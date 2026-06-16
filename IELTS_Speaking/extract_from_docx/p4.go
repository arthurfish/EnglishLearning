package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type dbRow struct {
	topic string
	seq   int
}

type spkTopic struct {
	id   string
	name string
}

const groupSize = 200

const promptTemplate = `我正在备考雅思口语。以下是我当前背诵的词伙词组列表，以及这些词组所属的分类：

%s

以下是可供匹配的雅思口语话题列表（topic_id, topic_name）：
%s

请根据词组的语义内容，找出与这些词组最相关的口语话题。只返回匹配的 topic_id，用空格分隔，放在 <spk_topic_ids> 标签里。
如果没有匹配的话题，就返回 <spk_topic_ids></spk_topic_ids>。

%s`

func main() {
	// --- 1. 读取 db.csv ---
	dbRows := readDB("db.csv")

	// --- 2. 读取 speaking.csv，去重得到 spk_topics (topic_id -> topic_name) ---
	spkTopics := readSpeaking("speaking.csv")
	spkTopicsStr := formatSpkTopics(spkTopics)

	// --- 3. 按 seq 分组，得到每组的 distinct topic ---
	groups := groupBySeq(dbRows)
	sortedKeys := make([]int, 0, len(groups))
	for k := range groups {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Ints(sortedKeys)

	// --- 4. 初始化 Client（和 reference.go 一样）---
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	)

	// --- 5. 并发调用 LLM ---
	type groupResult struct {
		groupNumber int
		topicIDs    []string
	}
	results := make(map[int][]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	type job struct {
		groupNumber int
		topics      []string
	}
	jobs := make(chan job, len(sortedKeys))

	// 下发所有任务
	for _, gn := range sortedKeys {
		jobs <- job{groupNumber: gn, topics: groups[gn]}
	}
	close(jobs)

	// 启动 worker 池（10 个并发）
	numWorkers := 10
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(wid int, client *openai.Client, spkStr string) {
			defer wg.Done()
			for jb := range jobs {
				distinctStr := formatDistinctTopics(jb.topics)
				require := "只返回 topic_id，用空格分隔，放在 <spk_topic_ids> 标签里。不要输出任何其他内容。"
				prompt := fmt.Sprintf(promptTemplate, distinctStr, spkStr, require)

				fmt.Printf("[Worker %d] Group %d (%d topics)... ", wid, jb.groupNumber, len(jb.topics))

				chatCompletion, err := client.Chat.Completions.New(
					context.TODO(), openai.ChatCompletionNewParams{
						Messages: []openai.ChatCompletionMessageParamUnion{
							openai.UserMessage(prompt),
						},
						Model: "qwen3.6-plus",
					},
				)
				if err != nil {
					log.Printf("[Worker %d] Group %d 调用失败: %v\n", wid, jb.groupNumber, err)
					mu.Lock()
					results[jb.groupNumber] = nil
					mu.Unlock()
					continue
				}

				content := chatCompletion.Choices[0].Message.Content
				ids := extractTopicIDs(content)
				fmt.Printf("匹配 %d 个 topic\n", len(ids))

				mu.Lock()
				results[jb.groupNumber] = ids
				mu.Unlock()
			}
		}(w, &client, spkTopicsStr)
	}

	wg.Wait()
	fmt.Println("\n所有组处理完毕。")

	// --- 6. 按 group_number 排序后输出 ---
	topicIDToName := buildTopicIDMap(spkTopics)

	out, err := os.Create("group_topic.csv")
	if err != nil {
		log.Fatalf("创建输出文件失败: %v", err)
	}
	defer out.Close()

	writer := csv.NewWriter(out)
	defer writer.Flush()

	writer.Write([]string{"group_number", "topic_id", "topic_name"})

	total := 0
	for _, gn := range sortedKeys {
		tids := results[gn]
		for _, tid := range tids {
			tname := topicIDToName[tid]
			writer.Write([]string{
				strconv.Itoa(gn),
				tid,
				tname,
			})
			total++
		}
	}

	fmt.Printf("完成！共 %d 组，%d 个匹配关系，已保存至 group_topic.csv\n", len(sortedKeys), total)
}

// extractTopicIDs 从 LLM 返回中提取 <spk_topic_ids> 标签内的内容
func extractTopicIDs(content string) []string {
	start := strings.Index(content, "<spk_topic_ids>")
	end := strings.Index(content, "</spk_topic_ids>")
	if start == -1 || end == -1 {
		// fallback: 直接从内容中提取数字
		var ids []string
		for _, token := range strings.Fields(content) {
			if _, err := strconv.Atoi(token); err == nil {
				ids = append(ids, token)
			}
		}
		return ids
	}
	body := content[start+len("<spk_topic_ids>") : end]
	var ids []string
	for _, token := range strings.Fields(body) {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if _, err := strconv.Atoi(token); err == nil {
			ids = append(ids, token)
		}
	}
	return ids
}

func formatDistinctTopics(topics []string) string {
	var sb strings.Builder
	for _, t := range topics {
		sb.WriteString(fmt.Sprintf("  - %s\n", t))
	}
	return sb.String()
}

func formatSpkTopics(topics []spkTopic) string {
	var sb strings.Builder
	for _, t := range topics {
		sb.WriteString(fmt.Sprintf("  topic_id=%s, topic_name=%s\n", t.id, t.name))
	}
	return sb.String()
}

func buildTopicIDMap(topics []spkTopic) map[string]string {
	m := make(map[string]string)
	for _, t := range topics {
		if _, ok := m[t.id]; !ok {
			m[t.id] = t.name
		}
	}
	return m
}

func groupBySeq(rows []dbRow) map[int][]string {
	groups := make(map[int]map[string]bool)
	for _, row := range rows {
		groupNum := (row.seq-1)/groupSize + 1
		if groups[groupNum] == nil {
			groups[groupNum] = make(map[string]bool)
		}
		if row.topic != "" {
			groups[groupNum][row.topic] = true
		}
	}
	result := make(map[int][]string)
	for k, v := range groups {
		var topics []string
		for t := range v {
			topics = append(topics, t)
		}
		result[k] = topics
	}
	return result
}

func readDB(path string) []dbRow {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("无法打开 %s: %v", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取 %s 失败: %v", path, err)
	}

	var rows []dbRow
	for i, record := range records {
		if i == 0 {
			continue // skip header
		}
		if len(record) < 4 {
			continue
		}
		seq, err := strconv.Atoi(strings.TrimSpace(record[3]))
		if err != nil {
			log.Printf("跳过第 %d 行, seq 解析失败: %v\n", i, err)
			continue
		}
		rows = append(rows, dbRow{
			topic: strings.TrimSpace(record[0]),
			seq:   seq,
		})
	}
	return rows
}

func readSpeaking(path string) []spkTopic {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("无法打开 %s: %v", path, err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("读取 %s 失败: %v", path, err)
	}

	seen := make(map[string]bool)
	var topics []spkTopic
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) < 5 {
			continue
		}
		tid := strings.TrimSpace(record[1])
		tname := strings.TrimSpace(record[3])
		key := tname
		if seen[key] {
			continue
		}
		seen[key] = true
		topics = append(topics, spkTopic{id: tid, name: tname})
	}
	return topics
}
