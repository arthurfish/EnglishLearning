import csv

# Manually filtered matches based on reviewing each group
# Key = (group_number, topic_name), value = True if kept
FILTERED = {
    # === Group 1 ===
    # DB: Academic Skills, Air Travel, Animal Testing, Architecture, Behavioural Problems,
    #     Business, Car Emissions, Celebrities, Charity, Clothes, Communication, Community,
    #     Computers, Consumer Society
    (1, "Teachers"):                 True,   # Academic Skills
    (1, "Social media"):             True,   # Communication
    (1, "Music"):                    False,  # 无关联
    (1, "Tidiness"):                 False,  # 无关联
    (1, "Websites"):                 True,   # Computers / Communication
    (1, "Watch"):                    False,  # 无关联
    (1, "Shopping"):                 True,   # Consumer Society
    (1, "Parks"):                    True,   # Community
    (1, "Cars"):                     True,   # Car Emissions
    (1, "Science"):                  True,   # Computers / Academic Skills
    (1, "Singing"):                  False,  # 无关联
    (1, "Clothing"):                 True,   # Clothes
    (1, "电子产品故障"):              True,   # Computers
    (1, "重要决定"):                  False,  # 无关联
    (1, "擅长学习和说语言的人"):       True,  # Academic Skills
    (1, "长时间未收到回复"):           False,  # 无关联
    (1, "遇到的科技问题"):            True,   # Computers
    (1, "推荐旅行过的地方"):          True,   # Air Travel
    (1, "最近看过的电视/网络节目"):    False,  # 无关联
    (1, "喜欢拜访但不想住的家"):       False,  # 无关联
    (1, "包含动物的故事或书"):         True,  # Animal Testing
    (1, "别人帮助解决问题"):           True,   # Charity / Community
    (1, "想颁布的环保法律"):           True,   # Car Emissions

    # === Group 2 ===
    # DB: Cover letter, Creativity, Crime, Discipline, Educational Equality,
    #     Energy Consumption, Environmental Problems, Environmental Protection,
    #     Exams, Family, Film, Food, Food Technology, Friends
    (2, "Teachers"):                 True,   # Educational Equality
    (2, "Social media"):             False,  # 无关联
    (2, "Tidiness"):                 False,  # 无关联
    (2, "Websites"):                 False,  # 无关联
    (2, "Shopping"):                 False,  # 无关联
    (2, "Feeling bored"):            False,  # 无关联
    (2, "Parks"):                    False,  # 无关联
    (2, "Cars"):                     False,  # 无关联
    (2, "Science"):                  True,   # Food Technology / Energy
    (2, "Telling Jokes"):            True,   # Creativity
    (2, "Headphone"):                False,  # 无关联
    (2, "Singing"):                  False,  # 无关联
    (2, "Clothing"):                 False,  # 无关联
    (2, "电子产品故障"):              True,   # Food Technology (tech)
    (2, "遇到的科技问题"):            True,   # Food Technology
    (2, "推荐旅行过的地方"):          False,  # 无关联
    (2, "最近看过的电视/网络节目"):    True,  # Film
    (2, "想颁布的环保法律"):           True,   # Environmental Protection / Problems

    # === Group 3 ===
    # DB: Friends, Gender Discrimination, Global Warming, Government Intervention,
    #     Greenspaces, Happiness, Health, Housing, Internet, Job Creation,
    #     Job Satisfaction, Job Seeking, Knowledge, Language, Laws, Lifestyle
    (3, "Teachers"):                 False,  # 无关联
    (3, "Social media"):             True,   # Internet
    (3, "Music"):                    False,  # 无关联
    (3, "Websites"):                 True,   # Internet
    (3, "Feeling bored"):            False,  # 无关联
    (3, "Parks"):                    True,   # Greenspaces
    (3, "Science"):                  True,   # Knowledge
    (3, "Headphone"):                False,  # 无关联
    (3, "擅长学习和说语言的人"):       True,  # Language / Knowledge
    (3, "遇到的科技问题"):            False,  # 无关联
    (3, "想颁布的环保法律"):           True,   # Laws / Government Intervention / Global Warming

    # === Group 4 ===
    # DB: Lifestyle, Living Standards, Materialistic World, Mechanisation,
    #     Medical Services, Motivated Students, Museums, Music, News,
    #     Older People, Online News, Online Shopping, Partying, Photos, Practical Work
    (4, "Teachers"):                 True,   # Motivated Students
    (4, "Social media"):             True,   # News / Online News
    (4, "Music"):                    True,   # Music
    (4, "Shopping"):                 True,   # Online Shopping / Materialistic World
    (4, "Feeling bored"):            True,   # Partying / Lifestyle
    (4, "Parks"):                    True,   # Lifestyle
    (4, "Science"):                  True,   # Mechanisation / Medical Services
    (4, "Telling Jokes"):            True,   # Partying
    (4, "Headphone"):                False,  # 无关联
    (4, "Clothing"):                 True,   # Lifestyle
    (4, "擅长学习和说语言的人"):       False,  # 无关联
    (4, "遇到的科技问题"):            True,   # Mechanisation
    (4, "推荐旅行过的地方"):          False,  # 无关联
    (4, "喜欢拜访但不想住的家"):       False,  # 无关联

    # === Group 5 ===
    # DB: Practical Work, Qualities, Reduce Poverty, Rules, Selfish People,
    #     Shopping, Skills, Social Activities, Social Life, Social Norms,
    #     Space Technology, Sport, Stress, Study Abroad
    (5, "Teachers"):                 True,   # Skills
    (5, "Music"):                    False,  # 无关联
    (5, "Tidiness"):                 False,  # 无关联
    (5, "Websites"):                 False,  # 无关联
    (5, "Shopping"):                 True,   # Shopping
    (5, "Feeling bored"):            False,  # 无关联
    (5, "Cars"):                     False,  # 无关联
    (5, "Outer space and stars"):    True,   # Space Technology
    (5, "Science"):                  True,   # Space Technology / Practical Work
    (5, "Mirrors"):                  False,  # 无关联
    (5, "Telling Jokes"):            False,  # 无关联
    (5, "Headphone"):                False,  # 无关联
    (5, "Singing"):                  False,  # 无关联
    (5, "重要决定"):                  True,   # Qualities
    (5, "擅长学习和说语言的人"):       True,  # Study Abroad / Skills
    (5, "遇到的科技问题"):            False,  # 无关联
    (5, "推荐旅行过的地方"):          True,   # Study Abroad
    (5, "喜欢拜访但不想住的家"):       False,  # 无关联
    (5, "想颁布的环保法律"):           True,   # Rules
    (5, "克服困难成功的人"):           True,   # Selfish People (contrast) / Qualities

    # === Group 6 ===
    # DB: Study Abroad, Talent, Teachers/Schools, Teamwork, Technology,
    #     Telecommunications, Teleworking, Tourism, Traffic Congestion,
    #     Transportation, Trees, University Subjects, Urbanisation, Waste, Water
    (6, "Teachers"):                 True,   # Teachers/Schools
    (6, "Social media"):             True,   # Telecommunications
    (6, "Websites"):                 True,   # Telecommunications / Technology
    (6, "Parks"):                    True,   # Trees / Urbanisation
    (6, "Cars"):                     True,   # Traffic Congestion / Transportation
    (6, "Outer space and stars"):    False,  # 无关联
    (6, "Science"):                  True,   # Technology / University Subjects
    (6, "Mirrors"):                  False,  # 无关联
    (6, "擅长学习和说语言的人"):       True,  # Study Abroad / University Subjects / Talent
    (6, "遇到的科技问题"):            True,   # Technology
    (6, "推荐旅行过的地方"):          True,   # Tourism / Study Abroad
    (6, "想颁布的环保法律"):           True,   # Waste / Water / Trees

    # === Group 7 ===
    # DB: Water, Wildlife Preservation, Women in the Workforce, Work Overtime, Zoo
    (7, "Teachers"):                 False,  # 无关联
    (7, "Parks"):                    True,   # Zoo
    (7, "Science"):                  True,   # Water / Wildlife
    (7, "Headphone"):                False,  # 无关联
    (7, "推荐旅行过的地方"):          False,  # 无关联
    (7, "想颁布的环保法律"):           True,   # Water / Wildlife Preservation
}


def main():
    # Read original, filter, write new
    kept = []
    with open("group_topic.csv", "r", encoding="utf-8-sig") as f:
        reader = csv.DictReader(f)
        for row in reader:
            gn = int(row["group_number"])
            tname = row["topic_name"]
            if FILTERED.get((gn, tname), False):
                kept.append(row)

    with open("filtered_group_topic.csv", "w", encoding="utf-8-sig", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=["group_number", "topic_id", "topic_name"])
        writer.writeheader()
        writer.writerows(kept)

    print(f"原始 {sum(1 for _ in open('group_topic.csv', encoding='utf-8-sig')) - 1} 条 -> 过滤后 {len(kept)} 条")


if __name__ == "__main__":
    main()
