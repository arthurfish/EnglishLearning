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
	"strings"
)

// 配置信息
const (
	BaseURL    = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	Model      = "qwen-vl-max" // 视觉任务建议使用专门的 VL 模型
	InputDir   = "pdf_input"
	PromptFile = "prompt.txt" // 外部提示词文件路径
)

var ApiKey = os.Getenv("DASHSCOPE_API_KEY") // 请替换为你的 API Key

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
	// 1. 读取外部提示词文件
	promptBytes, err := os.ReadFile(PromptFile)
	if err != nil {
		fmt.Printf("无法读取提示词文件 %s: %v\n请确保该文件存在于当前目录。\n", PromptFile, err)
		return
	}
	systemPrompt := string(promptBytes)

	// 2. 获取所有 PDF 文件
	files, err := filepath.Glob(filepath.Join(InputDir, "*.pdf"))
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	for _, pdfPath := range files {
		fileName := filepath.Base(pdfPath)
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		fmt.Printf("正在处理: %s...\n", fileName)

		// 3. 将 PDF 转换为图像序列
		imgPaths, err := convertPdfToImages(pdfPath)
		if err != nil {
			fmt.Printf("PDF 转换失败 [%s]: %v\n", fileName, err)
			continue
		}

		// 4. 构建请求内容
		var contents []Content
		// 先放文字提示（从文件读取的内容）
		contents = append(contents, Content{Type: "text", Text: systemPrompt})

		// 再放入所有图片
		for _, imgPath := range imgPaths {
			b64, err := fileToBase64(imgPath)
			if err != nil {
				continue
			}
			contents = append(contents, Content{
				Type:     "image_url",
				ImageURL: &ImageURL{URL: fmt.Sprintf("data:image/png;base64,%s", b64)},
			})
		}

		// 5. 调用 API
		csvResult, err := callVLM(contents)
		if err != nil {
			fmt.Printf("API 调用失败 [%s]: %v\n", fileName, err)
			continue
		}

		// 6. 保存 CSV
		outputFile := baseName + ".csv"
		err = os.WriteFile(outputFile, []byte(csvResult), 0644)
		if err != nil {
			fmt.Printf("保存失败: %v\n", err)
		} else {
			fmt.Printf("成功！结果已保存至: %s\n", outputFile)
		}

		// 清理临时图片
		cleanupImages(imgPaths)
	}
}

// 使用 pdftoppm 将 PDF 转换为 PNG
func convertPdfToImages(pdfPath string) ([]string, error) {
	tempDir := "temp_imgs"
	os.Mkdir(tempDir, 0755)

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

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", BaseURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		message := choices[0].(map[string]interface{})["message"].(map[string]interface{})
		content := message["content"].(string)

		// 清理 Markdown 代码块标识
		content = strings.TrimPrefix(content, "```csv\n")
		content = strings.TrimPrefix(content, "```csv")
		content = strings.TrimPrefix(content, "```\n")
		content = strings.TrimSuffix(content, "\n```")
		content = strings.TrimSuffix(content, "```")
		return strings.TrimSpace(content), nil
	}

	return "", fmt.Errorf("API 响应异常: %s", string(body))
}

func cleanupImages(paths []string) {
	for _, p := range paths {
		os.Remove(p)
	}
}
