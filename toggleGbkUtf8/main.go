package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: toggleGbkUtf8 <文件名>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		os.Exit(1)
	}

	var convertedData []byte
	var action string

	// 探测编码情况：如果是合法的 UTF-8 (包括纯 ASCII)，则当作 UTF-8 转换为 GBK；否则当作 GBK 转换为 UTF-8
	if utf8.Valid(data) {
		action = "UTF-8 -> GBK"
		convertedData, err = doTransform(data, simplifiedchinese.GBK.NewEncoder())
	} else {
		action = "GBK -> UTF-8"
		convertedData, err = doTransform(data, simplifiedchinese.GBK.NewDecoder())
	}

	if err != nil {
		fmt.Printf("❌ 转换过程失败 (%s): %v\n", action, err)
		os.Exit(1)
	}

	// 优化判断：如果是纯 ASCII 文本，转换前后字节完全一样，无需写入磁盘
	if bytes.Equal(data, convertedData) {
		fmt.Printf("ℹ️ 转换跳过: %s (文件可能是纯ASCII文本，UTF-8与GBK编码相同)\n", filename)
		return
	}

	// 将转换后的数据就地覆写回原文件
	// 使用与原文件相同或默认的安全权限 (0644)
	err = os.WriteFile(filename, convertedData, 0644)
	if err != nil {
		fmt.Printf("❌ 就地写入文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 转换成功: %s (%s)\n", filename, action)
}

// doTransform 接受原始字节切片和转换器，返回转换后的字节切片
func doTransform(src []byte, transformer transform.Transformer) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(src), transformer)
	return io.ReadAll(reader)
}
