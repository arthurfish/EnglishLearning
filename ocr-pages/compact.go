package main

import (
	"fmt"
	"os"
)

const startPage = 1
const endPage = 37
const pagesPerRequest = 6

func main() {
	compactCount := 1
	for i := startPage; i <= endPage; i += pagesPerRequest {

		fmt.Printf("Processing pages %d to %d\n", i, i+5)
		var accumulatedContent string
		for j := i; j < i+pagesPerRequest && j <= endPage; j++ {
			var currentContent string
			data, err := os.ReadFile(fmt.Sprintf("output/%d.md", j))
			if err != nil {
				fmt.Printf("Error reading file %d.md: %v\n", j, err)
				continue
			}
			currentContent = string(data)
			accumulatedContent += currentContent + "\n"
		}
		os.WriteFile(fmt.Sprintf("output/compact_%d.md", compactCount), []byte(accumulatedContent), 0644)
		compactCount++
	}
}
