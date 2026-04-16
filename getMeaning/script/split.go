package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// 1. 创建 split 文件夹
	outputDir := "split"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		log.Fatalf("无法创建文件夹 %s: %v", outputDir, err)
	}

	// 2. 读取当前目录下的所有文件
	files, err := os.ReadDir(".")
	if err != nil {
		log.Fatalf("无法读取当前目录: %v", err)
	}

	for _, file := range files {
		// 只处理 .md 文件，跳过文件夹
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".md") {
			continue
		}

		fmt.Printf("正在处理文件: %s\n", file.Name())

		// 读取文件内容
		content, err := os.ReadFile(file.Name())
		if err != nil {
			log.Printf("读取文件 %s 失败: %v", file.Name(), err)
			continue
		}

		// 3. 按行分割，并删除空行，得到无空行的行字符串数组
		rawLines := strings.Split(string(content), "\n")
		var validLines []string
		for _, line := range rawLines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				validLines = append(validLines, trimmedLine)
			}
		}

		// 4. 用 “---” 分割成多个部分
		var currentPart []string
		for _, line := range validLines {
			if line == "---" {
				// 遇到分割线，处理当前累积的部分
				processPart(currentPart, outputDir)
				currentPart = nil // 清空当前部分，准备接收下一部分
			} else {
				currentPart = append(currentPart, line)
			}
		}
		// 处理文件末尾的最后一部分
		processPart(currentPart, outputDir)
	}
	
	fmt.Println("所有文件处理完成！请查看 split 文件夹。")
}

// processPart 处理每个分割后的部分
func processPart(lines []string, outputDir string) {
	if len(lines) == 0 {
		return // 忽略空的部分
	}

	firstLine := lines[0]
	// 5. 保证最开始有个类似 “# List 1” 开头的行
	if !strings.HasPrefix(firstLine, "# ") {
		log.Printf("警告: 发现一个不以 '# ' 开头的部分，已跳过。开头为: %s", firstLine)
		return
	}

	// 6. 建立文件名为 “# ” 后部分的文件 (例如 List 1.txt)
	fileName := strings.TrimPrefix(firstLine, "# ")
	fileName = strings.TrimSpace(fileName) + ".txt"
	filePath := filepath.Join(outputDir, fileName)

	// 7. 文件内容为分割后部分除去标题行的所有内容
	// 将剩余的行重新用换行符拼接
	fileContent := strings.Join(lines[1:], "\n")

	// 写入文件
	err := os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		log.Printf("写入文件 %s 失败: %v", filePath, err)
	} else {
		fmt.Printf("  -> 成功生成文件: %s\n", fileName)
	}
}