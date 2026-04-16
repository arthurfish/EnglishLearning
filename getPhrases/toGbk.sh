#!/bin/bash

# 创建输出目录，防止覆盖原文件
OUTPUT_DIR="gbk_output"
mkdir -p "$OUTPUT_DIR"

echo "开始转换编码至 GBK..."

# 遍历当前目录下的所有 csv 文件
for file in *.csv; do
    # 检查文件是否存在（防止没有匹配到文件时报错）
    [ -e "$file" ] || continue

    echo "正在转换: $file"

    # 使用 iconv 进行转换
    # -f UTF-8: 原始编码（根据你的 VLM 输出通常是 UTF-8）
    # -t GBK//IGNORE: 目标编码为 GBK，//IGNORE 表示忽略无法转换的非法字符
    # -c: 丢弃无效字符
    iconv -f UTF-8 -t GBK//IGNORE "$file" > "$OUTPUT_DIR/$file"

    if [ $? -eq 0 ]; then
        echo "成功: $OUTPUT_DIR/$file"
    else
        echo "失败: $file"
    fi
done

echo "-----------------------------------"
echo "所有转换任务已完成。请在 $OUTPUT_DIR 目录查看结果。"