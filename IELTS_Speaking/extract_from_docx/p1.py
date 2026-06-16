import json
import re
from docx import Document


def parse_docx_to_json(docx_path, json_path):
    doc = Document(docx_path)

    result = []
    current_category = None
    current_topic = None
    in_part3 = False  # Track if we're inside a "Part 3" subsection

    # Regex: section header = 【...】...Part/必考/保留...（N道）
    # Real section headers always end with （N道）— intro paragraphs don't
    section_pattern = re.compile(r"【[^】]+】[^（(]*[（(]\d+道[）)]")
    # Regex: topic header like 【1】（5.28）Teachers  or  【1】Work/Study  or  【12】自行车/摩托车/电动车（1.30）
    topic_pattern = re.compile(r"^【(\d+)】[（(](\d[^）)]*)[）)]\s*(.*)")
    # Topic header without date: 【1】Pets and Animals
    topic_no_date_pattern = re.compile(r"^【(\d+)】(.+)")

    for paragraph in doc.paragraphs:
        text = paragraph.text.strip()
        if not text:
            continue

        # 1. Section header: e.g. 【5月新题】Part1（17道）, 【Part1必考题】（5道）, 【Part2&3】保留旧题（25道）
        if section_pattern.search(text):
            current_category = {"category": text, "topics": []}
            result.append(current_category)
            current_topic = None
            in_part3 = False
            continue

        # 2. Topic header with date: 【N】（M.DD）TopicName
        topic_match = topic_pattern.match(text)
        if topic_match:
            t_id, t_date, t_name = topic_match.groups()
            current_topic = {
                "topic_id": t_id,
                "date": t_date,
                "topic_name": t_name.strip(),
                "questions": [],
            }
            in_part3 = False
            if current_category:
                current_category["topics"].append(current_topic)
            continue

        # 3. Topic header without date: 【N】TopicName
        topic_match2 = topic_no_date_pattern.match(text)
        if topic_match2:
            t_id, t_name = topic_match2.groups()
            # Skip if it looks like a section header
            if "Part" in t_name or "必考" in t_name or "保留" in t_name:
                continue
            current_topic = {
                "topic_id": t_id,
                "date": "",
                "topic_name": t_name.strip(),
                "questions": [],
            }
            in_part3 = False
            if current_category:
                current_category["topics"].append(current_topic)
            continue

        # 4. "Part 3" sub-header inside a Part2&3 topic
        if text == "Part 3":
            in_part3 = True
            continue

        # 5. Question lines (under current topic)
        if current_topic is not None:
            current_topic["questions"].append(text)

    # Write JSON
    with open(json_path, "w", encoding="utf-8") as f:
        json.dump(result, f, ensure_ascii=False, indent=2)

    # Print summary
    total_topics = sum(len(c["topics"]) for c in result)
    total_questions = sum(
        len(q["questions"]) for c in result for q in c["topics"]
    )
    print(f"解析完成！共 {len(result)} 个分类, {total_topics} 个话题, {total_questions} 道题目")
    print(f"数据已保存至 {json_path}")
    for cat in result:
        print(f"  - {cat['category']}: {len(cat['topics'])} topics")


parse_docx_to_json("./2026年5-8月口语题库.docx", "ielts_speaking_bank.json")