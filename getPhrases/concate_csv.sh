cd result_2
awk 1 技能类短语.csv 效率类短语.csv 权利类短语.csv 责任类短语.csv 环境类短语.csv 健康类短语.csv 财务类短语.csv 安全类短语.csv 文化类短语.csv 乐趣类短语.csv > all.csv
sed -i 's/，/,/g' all.csv
iconv -f UTF-8 -t GBK all.csv > _all.csv
mv _all.csv all.csv