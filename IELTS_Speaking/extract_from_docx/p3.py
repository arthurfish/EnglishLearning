import csv
import json
import re
from docx import Document


def parse_docx_to_csv(docx_path, csv_path):
    doc = Document(docx_path)

    section_pattern = re.compile(r"【[^】]+】[^（(]*[（(]\d+道[）)]")
    topic_with_date = re.compile(r"^【(\d+)】[（(](\d[^）)]*)[）)]\s*(.*)")
    topic_no_date = re.compile(r"^【(\d+)】(.+)")

    rows = []
    current_category = None
    current_topic = None
    state = "idle"  # idle | cue_card | part3

    for paragraph in doc.paragraphs:
        text = paragraph.text.strip()
        if not text:
            continue

        # Section header
        if section_pattern.search(text):
            current_category = text
            current_topic = None
            state = "idle"
            continue

        # Topic header (with date)
        m = topic_with_date.match(text)
        if m:
            current_topic = {
                "topic_id": m.group(1),
                "date": m.group(2),
                "topic_name": m.group(3).strip(),
            }
            state = "idle"
            continue

        # Topic header (no date)
        m2 = topic_no_date.match(text)
        if m2 and current_category:
            name = m2.group(2).strip()
            if "Part" in name or "必考" in name or "保留" in name:
                continue
            current_topic = {
                "topic_id": m2.group(1),
                "date": "",
                "topic_name": name,
            }
            state = "idle"
            continue

        if current_topic is None:
            continue

        # --- Collect questions based on state ---

        # "You should say:" → switch to cue_card mode, skip this line
        if text == "You should say:":
            state = "cue_card"
            continue

        # "Part 3" / "Part3" / "Part3:" → switch to part3 mode
        if text.startswith("Part 3") or text.startswith("Part3"):
            state = "part3"
            continue

        # Cue card bullet points → store in a separate field, skip from questions
        if state == "cue_card":
            continue

        # Determine question type
        if state == "part3":
            q_type = "Part3"
        elif current_topic["topic_name"] and any(
            c in current_category for c in ("Part 2", "Part2")
        ):
            # Part 2&3 section, but not yet in Part 3 → this is the Describe prompt
            q_type = "Part2"
            state = "idle"  # After Describe, go back to idle (may hit cue_card or part3 next)
        else:
            q_type = "Part1"

        rows.append({
            "category": current_category,
            "topic_id": current_topic["topic_id"],
            "date": current_topic["date"],
            "topic_name": current_topic["topic_name"],
            "question_type": q_type,
            "question": text,
        })

    # Write CSV
    with open(csv_path, "w", encoding="utf-8-sig", newline="") as f:
        writer = csv.DictWriter(
            f,
            fieldnames=["category", "topic_id", "date", "topic_name", "question_type", "question"],
        )
        writer.writeheader()
        writer.writerows(rows)

    print(f"已导出 {len(rows)} 道题目至 {csv_path}")

    # Summary
    from collections import Counter
    counts = Counter(r["question_type"] for r in rows)
    print(f"  Part1: {counts.get('Part1', 0)}")
    print(f"  Part2: {counts.get('Part2', 0)}")
    print(f"  Part3: {counts.get('Part3', 0)}")


parse_docx_to_csv("./2026年5-8月口语题库.docx", "ielts_speaking_bank.csv")
