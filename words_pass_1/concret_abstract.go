package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// 提示词模板
const promptTemplate = `
我正在背雅思单词。请根据我提供的单词列表生成一个csv文件。这个csv文件分三列：english, concrete meaning, abstract meaning。english就是英文单词原文，concrete meaning是指单词的一种非常具体的含义（如果有多种具体的含义，随便挑一个常见的就行），而abstract meaning是单词一种有些抽象的含义（如果有多种抽象的含义，随便挑一个常见的就行）。注意这两个含义必须使用中文，最好是语义相差很少的单个中文词语，但如果多个字能更好的表达那就用更多字。如果这两个meaning各自没有的话，留空就行，但每个单词总会至少有一个中文意思的。注意栏间隔是英文逗号，csv文件不要有表头！

例如：
plough,犁,耕耘

我会把你的输出直接作为csv文件存储，所以不要输出任何多余的Markdown内容。以下是我的单词列表，每行一个单词。

单词列表如下：
%s`

func main() {
	inputDir := "split"
	outputDir := "csv"

	// 1. 创建 csv 文件夹
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("无法创建输出文件夹 %s: %v", outputDir, err)
	}

	// 2. 完全使用你提供的代码初始化 Client
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("DASHSCOPE_API_KEY")),
		option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1"),
	)

	// 3. 读取 split 文件夹下的所有 txt 文件
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("无法读取输入目录 %s: %v", inputDir, err)
	}

	var txtFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			txtFiles = append(txtFiles, file.Name())
		}
	}

	if len(txtFiles) == 0 {
		fmt.Println("在 split 文件夹中没有找到 txt 文件。")
		return
	}

	// 4. 设置并发 Worker 池
	numWorkers := 10
	jobs := make(chan string, len(txtFiles))
	var wg sync.WaitGroup

	// 启动 N=10 个 worker
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg, &client, inputDir, outputDir)
	}

	// 5. 将任务分配给 jobs 通道
	for _, fileName := range txtFiles {
		jobs <- fileName
	}
	close(jobs) // 所有任务下发完毕，关闭通道

	// 6. 等待所有 worker 完成
	wg.Wait()
	fmt.Println("所有文件均已通过千问处理完成！")
}

// worker 执行具体的任务
func worker(id int, jobs <-chan string, wg *sync.WaitGroup, client *openai.Client, inputDir, outputDir string) {
	defer wg.Done()

	for fileName := range jobs {
		inputPath := filepath.Join(inputDir, fileName)

		// 读取 txt 文件内容
		content, err := os.ReadFile(inputPath)
		if err != nil {
			log.Printf("[Worker %d] 读取文件 %s 失败: %v\n", id, fileName, err)
			continue
		}

		// 将内容嵌入提示词模板
		prompt := fmt.Sprintf(promptTemplate, string(content))

		// --- 以下完全使用你提供的大模型调用代码结构 ---
		chatCompletion, err := client.Chat.Completions.New(
			context.TODO(), openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(prompt),
				},
				Model: "qwen3.5-plus",
			},
		)

		if err != nil {
			log.Printf("[Worker %d] 调用模型失败 (%s): %v\n", id, fileName, err)
			continue
		}

		resultContent := chatCompletion.Choices[0].Message.Content
		// --- 大模型调用代码结构结束 ---

		// 确定输出文件的路径 (后缀改为 .csv)
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		outputFileName := baseName + ".csv"
		outputPath := filepath.Join(outputDir, outputFileName)

		// 将模型的输出写入 CSV 文件
		err = os.WriteFile(outputPath, []byte(resultContent), 0644)
		if err != nil {
			log.Printf("[Worker %d] 写入文件 %s 失败: %v\n", id, outputPath, err)
			continue
		}

		// 任务完成后在终端说一嘴
		fmt.Printf("[Worker %d] 完成处理: %s -> %s\n", id, fileName, outputFileName)
	}
}
