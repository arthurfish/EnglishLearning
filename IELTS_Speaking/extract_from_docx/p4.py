import csv
import os
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from openai import OpenAI

GROUP_SIZE = 200

PROMPT_TEMPLATE = """\
我正在备考雅思口语。以下是我当前背诵的词伙词组列表，以及这些词组所属的雅思话题分类：

{db_distinct_topics}

以下是可供匹配的雅思口语话题列表：
{spk_topics}

请根据词组的内容，找出与这个词组最相关的口语话题。每个话题返回 topic_id 即可，用空格隔开多个 topic_id，放在 <spk_topic_ids> 标签里。
如果没有匹配的话题，就返回 <spk_topic_ids></spk_topic_ids>。

require: {require}
"""


def main():
    # --- 1. 读取 db.csv ---
    db_rows = []
    with open("db.csv", "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            db_rows.append(row)

    # --- 2. 读取 speaking.csv，构造 spk_topics ---
    spk_rows = []
    with open("speaking.csv", "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            spk_rows.append(row)

    # 去重: topic_id -> topic_name
    seen = {}
    for row in spk_rows:
        tid = row["topic_id"].strip()
        tname = row["topic_name"].strip()
        if tid not in seen:
            seen[tid] = tname
    spk_topics = [(tid, tname) for tid, tname in seen.items()]
    spk_topics_str = "\n".join(f"  topic_id={tid}, topic_name={tname}" for tid, tname in spk_topics)

    # --- 3. 按 seq 分组 ---
    groups = {}
    for row in db_rows:
        seq = int(row["seq"].strip())
        group_number = (seq - 1) // GROUP_SIZE + 1
        if group_number not in groups:
            groups[group_number] = set()
        topic = row["topic"].strip()
        if topic:
            groups[group_number].add(topic)

    # --- 4. 构造 LLM 调用 ---
    client = OpenAI(
        api_key=os.environ["DASHSCOPE_API_KEY"],
        base_url="https://dashscope.aliyuncs.com/compatible-mode/v1",
    )

    results = {}  # group_number -> list of topic_ids

    def call_llm(group_number, distinct_topics):
        require = f"请只返回 <spk_topic_ids> 标签包裹的 topic_id 数字，用空格分隔。不要输出任何其他内容。"
        topics_list = "\n".join(f"  - {t}" for t in sorted(distinct_topics))
        prompt = PROMPT_TEMPLATE.format(
            db_distinct_topics=topics_list,
            spk_topics=spk_topics_str,
            require=require,
        )

        print(f"[Group {group_number}] 调用 LLM... ({len(distinct_topics)} distinct topics)")

        response = client.chat.completions.create(
            model="qwen3.6-plus",
            messages=[{"role": "user", "content": prompt}],
        )

        content = response.choices[0].message.content.strip()
        print(f"[Group {group_number}] 返回: {content}")

        # 提取 <spk_topic_ids> 标签里的内容
        import re
        m = re.search(r"<spk_topic_ids>(.*?)</spk_topic_ids>", content, re.DOTALL)
        if m:
            ids_text = m.group(1).strip()
            if ids_text:
                ids = [tid.strip() for tid in ids_text.split() if tid.strip().isdigit()]
            else:
                ids = []
        else:
            # fallback: 尝试直接从内容中提取数字
            ids = [s for s in content.split() if s.strip().isdigit()]

        results[group_number] = ids

    # 顺序调用，避免并发超限
    sorted_groups = sorted(groups.items())
    for group_number, distinct_topics in sorted_groups:
        try:
            call_llm(group_number, distinct_topics)
        except Exception as e:
            print(f"[Group {group_number}] 错误: {e}")
            results[group_number] = []
        time.sleep(1)

    # --- 5. 输出 group_topic.csv ---
    with open("group_topic.csv", "w", encoding="utf-8-sig", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["group_number", "topic_id", "topic_name"])

        topic_id_to_name = {tid: tname for tid, tname in spk_topics}

        for group_number in sorted(results.keys()):
            for tid in sorted(results[group_number], key=int):
                tname = topic_id_to_name.get(tid, "Unknown")
                writer.writerow([group_number, tid, tname])

    total = sum(len(ids) for ids in results.values())
    print(f"\n完成！共 {len(results)} 组，{total} 个匹配，已保存至 group_topic.csv")


if __name__ == "__main__":
    main()
