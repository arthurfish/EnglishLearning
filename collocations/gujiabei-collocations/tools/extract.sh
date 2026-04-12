#!/bin/sh

# 1. 基础参数校验
if [ "$#" -ne 3 ]; then
    echo "用法: $0 <命令: o|p|d> <起始序号> <结束序号>"
    echo "示例: $0 d 1 10"
    exit 1
fi

cmd=$1
from=$2
to=$3

case $cmd in
    "o") 
        # 增加 (HEADER FALSE) 避免导出表头
        duckdb.exe -c "copy (select chinese, phrase from 'gujiabei-collocations.csv' where seq between $from and $to order by seq) to '$from-$to.origin.csv' (HEADER FALSE);"
        mv "$from-$to.origin.csv" "$from-$to.origin.txt"
        echo "已生成: $from-$to.origin.txt"
        ;;

    "p") 
        duckdb.exe -c "copy (select chinese from 'gujiabei-collocations.csv' where seq between $from and $to order by seq) to '$from-$to.practice.csv' (HEADER FALSE);"
        # 兼容 Windows(\r\n) 和 Linux(\n) 换行符，将 -i 前置
        sed -i -E 's/\r?$/, /g' "$from-$to.practice.csv"
        echo "已生成: $from-$to.practice.csv"
        ;;

    "d") 
        duckdb.exe -c "copy (select chinese, phrase from 'gujiabei-collocations.csv' where seq between $from and $to order by seq) to '$from-$to.display.csv' (HEADER FALSE);"
        # 指定 awk 以逗号为分隔符 (注意: 假设数据内部不包含被引号包裹的逗号)
        awk -F ',' '{
            # 移除可能存在的 Windows 回车符 \r
            sub(/\r$/, "", $2); 
            printf "> %s\n%s\n\n", $1, $2
        }' "$from-$to.display.csv" > "$from-$to.display.txt"
        rm "$from-$to.display.csv"
        echo "已生成: $from-$to.display.txt"
        ;;
        
    *)
        echo "未知命令: $cmd"
        echo "支持的命令为: o (origin), p (practice), d (display)"
        exit 1
        ;;
esac