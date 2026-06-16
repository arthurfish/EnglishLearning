import csv
import json


def json_to_csv(json_path, csv_path):
    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    with open(csv_path, "w", encoding="utf-8-sig", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["category", "topic_id", "date", "topic_name", "question"])

        for cat in data:
            for topic in cat["topics"]:
                for question in topic["questions"]:
                    writer.writerow([
                        cat["category"],
                        topic["topic_id"],
                        topic["date"],
                        topic["topic_name"],
                        question,
                    ])

    print(f"已导出至 {csv_path}")


json_to_csv("ielts_speaking_bank.json", "ielts_speaking_bank.csv")
