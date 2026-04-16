package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// 配置信息
const (
	BaseURL    = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	Model      = "qwen3.5-plus"
	InputFile  = "input.pdf" // 指定输入的单个 PDF 文件
	OutputDir  = "output"    // Markdown 输出文件夹
	PromptFile = "prompt.txt"
)

var ApiKey = os.Getenv("DASHSCOPE_API_KEY")

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

func main() {
	if ApiKey == "" {
		fmt.Println("错误：未找到 DASHSCOPE_API_KEY 环境变量，请先设置 API Key。")
		return
	}

	// 1. 读取外部提示词文件（如果不存在，给个默认的 OCR 提示词）
	systemPrompt := "请将这一页单词书中的内容完整地转化为 Markdown 格式。不需要保持原有的排版。但必须使用Markdown语法来保持原来的层级和有用的信息。不要输出任何额外的解释。你的输出会直接变成Markdown文件存储下来，所以请不要废话。"
	promptBytes, err := os.ReadFile(PromptFile)
	if err == nil {
		systemPrompt = string(promptBytes)
	} else {
		fmt.Printf("未找到 %s，将使用默认 OCR 提示词。\n", PromptFile)
	}

	// 2. 检查输入文件是否存在
	if _, err := os.Stat(InputFile); os.IsNotExist(err) {
		fmt.Printf("找不到输入文件: %s\n", InputFile)
		return
	}

	// 3. 创建输出目录
	os.MkdirAll(OutputDir, 0755)

	fmt.Printf("正在将 %s 转换为图片序列...\n", InputFile)

	// 4. 将 PDF 转换为图像序列
	imgPaths, err := convertPdfToImages(InputFile)
	if err != nil {
		fmt.Printf("PDF 转换失败: %v\n", err)
		return
	}
	defer cleanupImages(imgPaths) // 确保程序结束时清理临时图片

	// 准备提取页码的正则（匹配如 "-1.png", "-02.png" 里的数字）
	pageRegex := regexp.MustCompile(`-(\d+)\.png$`)

	// 5. 遍历每一页图片并单独调用 API
	for _, imgPath := range imgPaths {
		// 提取真实页码
		pageNum := 0
		matches := pageRegex.FindStringSubmatch(imgPath)
		if len(matches) > 1 {
			pageNum, _ = strconv.Atoi(matches[1])
		} else {
			fmt.Printf("跳过无法识别页码的图片: %s\n", imgPath)
			continue
		}

		fmt.Printf("正在对第 %d 页进行 OCR 识别...\n", pageNum)

		b64, err := fileToBase64(imgPath)
		if err != nil {
			fmt.Printf("图片读取失败 [%s]: %v\n", imgPath, err)
			continue
		}

		// 构建单页请求内容
		contents := []Content{
			{Type: "text", Text: systemPrompt},
			{
				Type:     "image_url",
				ImageURL: &ImageURL{URL: fmt.Sprintf("data:image/png;base64,%s", b64)},
			},
		}

		// 调用 API
		mdResult, err := callVLM(contents)
		if err != nil {
			fmt.Printf("第 %d 页 API 调用失败: %v\n", pageNum, err)
			continue
		}

		// 6. 保存 Markdown 文件
		outputFile := filepath.Join(OutputDir, fmt.Sprintf("%d.md", pageNum))
		err = os.WriteFile(outputFile, []byte(mdResult), 0644)
		if err != nil {
			fmt.Printf("第 %d 页保存失败: %v\n", pageNum, err)
		} else {
			fmt.Printf("-> 成功！第 %d 页结果已保存至: %s\n", pageNum, outputFile)
		}
	}
	fmt.Println("所有页面处理完毕！")
}

// 使用 pdftoppm 将 PDF 转换为 PNG
func convertPdfToImages(pdfPath string) ([]string, error) {
	tempDir := "temp_imgs"
	os.MkdirAll(tempDir, 0755)

	prefix := filepath.Join(tempDir, "page")
	cmd := exec.Command("pdftoppm", "-png", pdfPath, prefix)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	matches, _ := filepath.Glob(prefix + "-*.png")
	return matches, nil
}

func fileToBase64(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func callVLM(contents []Content) (string, error) {
	requestBody := map[string]interface{}{
		"model":    Model,
		"messages": []Message{{Role: "user", Content: contents}},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("JSON 序列化失败: %v", err)
	}

	// 1. 修复点：清理 BaseURL 可能包含的隐藏字符，并严格检查 req 错误
	cleanBaseURL := strings.TrimSpace(BaseURL)
	req, err := http.NewRequest("POST", cleanBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		// 如果这里报错，程序不会崩溃，而是会把真实的错误原因打印出来
		return "", fmt.Errorf("创建 HTTP 请求失败 (请检查 BaseURL 是否有误): %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API 请求发送失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取 API 响应失败: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("API 返回了非 JSON 数据: %s", string(body))
	}

	// 2. 修复点：安全地提取大模型的回复，防止接口报错(如限流/余额不足)引发 panic
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		// 走到这里说明 API 拒绝了请求，将完整响应打印出来供你排查
		return "", fmt.Errorf("API 未返回预期结果，可能发生了报错。完整响应: %s", string(body))
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("API 响应结构异常(choice): %s", string(body))
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("API 响应结构异常(message): %s", string(body))
	}

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("API 响应结构异常(content): %s", string(body))
	}

	// 清理大模型可能带有的 Markdown 代码块标识
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```markdown")
	content = strings.TrimPrefix(content, "```md")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	return strings.TrimSpace(content), nil
}

func cleanupImages(paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
	// 可选：清理完图片后顺手删除临时目录
	os.Remove("temp_imgs")
}
