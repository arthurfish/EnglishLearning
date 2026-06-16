import csv

COLS = 3

def main():
    # --- 1. 读取 db.csv，按 group 收集 phrase ---
    groups = {}
    with open("db.csv", "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            seq = int(row["seq"].strip())
            gn = (seq - 1) // 200 + 1
            groups.setdefault(gn, []).append(row["phrase"].strip())

    # --- 2. 生成 markdown ---
    lines = []

    for gn in sorted(groups.keys()):
        phrases = groups[gn]
        lines.append(f"# Group {gn}\n")

        # 按列数分组
        for i in range(0, len(phrases), COLS):
            row = phrases[i:i + COLS]
            lines.append("| " + " | ".join(row) + " |")
            if i == 0:
                lines.append("| " + " | ".join("---" for _ in range(COLS)) + " |")

        lines.append("")

    with open("group_md/IELTS_Phrase_Tables.md", "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    total = sum(len(v) for v in groups.values())
    print(f"已生成 group_md/IELTS_Phrase_Tables.md ({len(groups)} groups, {total} phrases, {COLS} columns)")


if __name__ == "__main__":
    main()
