import csv
import os


def main():
    # --- 1. 读取 filtered_group_topic.csv ---
    groups = {}  # group_number -> [topic_name]
    with open("filtered_group_topic.csv", "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        for row in reader:
            gn = int(row["group_number"])
            tname = row["topic_name"]
            groups.setdefault(gn, []).append(tname)

    # --- 2. 读取 speaking.csv，按 topic_name 收集 Part1/Part2/Part3 ---
    topic_questions = {}  # topic_name -> {"Part1": [...], "Part2": [...], "Part3": [...]}
    with open("speaking.csv", "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        for row in reader:
            tname = row["topic_name"]
            qtype = row["question_type"]
            question = row["question"]
            if tname not in topic_questions:
                topic_questions[tname] = {"Part1": [], "Part2": [], "Part3": []}
            topic_questions[tname][qtype].append(question)

    # --- 3. 生成单个 markdown ---
    lines = []

    for gn in sorted(groups.keys()):
        tnames = groups[gn]
        lines.append(f"# Group {gn}\n")

        for tname in tnames:
            lines.append(f"## Topic [{tname}] - Group {gn}\n")
            q = topic_questions.get(tname, {"Part1": [], "Part2": [], "Part3": []})

            if q["Part1"]:
                lines.append("### Part1\n")
                for question in q["Part1"]:
                    lines.append(question)
                lines.append("")

            if q["Part2"]:
                lines.append("### Part2\n")
                for question in q["Part2"]:
                    lines.append(question)
                lines.append("")

            if q["Part3"]:
                lines.append("### Part3\n")
                for question in q["Part3"]:
                    lines.append(question)
                lines.append("")

    with open("group_md/IELTS_Speaking_Groups_filtered.md", "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    print(f"已生成 group_md/IELTS_Speaking_Groups_filtered.md ({len(groups)} groups, {sum(len(v) for v in groups.values())} topics)")


if __name__ == "__main__":
    main()
