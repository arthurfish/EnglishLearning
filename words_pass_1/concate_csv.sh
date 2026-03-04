cd csv
awk 1 *.csv > all.csv
sed -i 's/，/,/g' all.csv
iconv -f UTF-8 -t GBK all.csv > _all.csv
mv _all.csv all.csv