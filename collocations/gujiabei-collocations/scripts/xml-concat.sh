#!/bin/bash
# 用法: ./merge_xml.sh <目录路径> > output.xml

# 1. 参数与目录校验
if [[ -z "$1" ]]; then
    echo "用法: $0 <目录路径>" >&2
    exit 1
fi

dir="$1"
if [[ ! -d "$dir" ]]; then
    echo "错误: '$dir' 不是一个有效的目录。" >&2
    exit 1
fi

# 2. XML 字符转义函数（必须优先转义 &）
xml_escape() {
    sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g'
}

echo "<files>"
for file in "$dir"/*; do
    # 兼容空目录：若 glob 未展开，跳过字面量路径
    [[ -e "$file" ]] || continue
    # 仅处理普通文件（跳过子目录、软链接等）
    [[ -f "$file" ]] || continue

    # 提取文件名并转义
    filename=$(basename "$file")
    escaped_title=$(printf '%s' "$filename" | xml_escape)

    echo "  <file>"
    echo "    <file-title>$escaped_title</file-title>"
    echo "    <file-content>"
    # 流式读取并转义内容，保留原始换行，不占用额外内存
    xml_escape < "$file"
    echo "    </file-content>"
    echo "  </file>"
done
echo "</files>"