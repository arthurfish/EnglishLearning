import csv
import re

COLS = 3


def latex_escape(s):
    # Order matters: & first, then \ (to not double-escape backslashes)
    s = s.replace("&", r"\&")
    s = s.replace("%", r"\%")
    s = s.replace("$", r"\$")
    s = s.replace("#", r"\#")
    s = s.replace("_", r"\_")
    s = s.replace("{", r"\{")
    s = s.replace("}", r"\}")
    s = s.replace("~", r"\textasciitilde{}")
    s = s.replace("^", r"\textasciicircum{}")
    return s


def main():
    # --- 1. 读取 db.csv，按 group 收集 phrase ---
    groups = {}
    with open("db.csv", "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            seq = int(row["seq"].strip())
            gn = (seq - 1) // 200 + 1
            groups.setdefault(gn, []).append(row["phrase"].strip())

    # --- 2. 生成 LaTeX ---
    lines = []
    lines.append(r"""\documentclass[10pt,a4paper]{article}
\usepackage[UTF8]{ctex}
\usepackage{geometry}
\usepackage{booktabs}
\usepackage{xcolor}
\usepackage{longtable}
\usepackage{colortbl}
\geometry{left=1.5cm,right=1.5cm,top=2cm,bottom=2cm}

\begin{document}
""")

    for gn in sorted(groups.keys()):
        phrases = groups[gn]
        lines.append(f"\\section*{{Group {gn}}}\n")
        lines.append(r"\setlength{\tabcolsep}{8pt}")
        lines.append(r"\renewcommand{\arraystretch}{1.3}")
        lines.append(r"{\large")
        col_width = str(round(1.0 / COLS, 4))
        col_spec = "*" + "{" + str(COLS) + "}{" + r"p{" + col_width + r"\textwidth}" + "}"
        lines.append(r"\begin{longtable}{@{}" + col_spec + r"@{}}")
        lines.append(r"\toprule")

        # Column headers
        lines.append(" & ".join(r"\textbf{Phrase " + str(i+1) + "}" for i in range(COLS)) + r" \\ \midrule")
        lines.append(r"\endfirsthead")

        # Repeating header on new pages
        lines.append(r"\multicolumn{" + str(COLS) + r"}{r}{\footnotesize (continued)} \\")
        lines.append(r"\toprule")
        lines.append(" & ".join(r"\textbf{Phrase " + str(i+1) + "}" for i in range(COLS)) + r" \\ \midrule")
        lines.append(r"\endhead")

        # Footer for page breaks
        lines.append(r"\midrule")
        lines.append(r"\multicolumn{" + str(COLS) + r"}{r}{\footnotesize continued on next page} \\")
        lines.append(r"\endfoot")

        lines.append(r"\bottomrule")
        lines.append(r"\endlastfoot")

        # Rows
        for i in range(0, len(phrases), COLS):
            row = phrases[i:i + COLS]
            escaped = [latex_escape(p) for p in row]
            while len(escaped) < COLS:
                escaped.append("")
            lines.append(" & ".join(escaped) + r" \\")

        lines.append(r"\end{longtable}}")
        lines.append("")

    lines.append(r"\end{document}")

    with open("group_md/IELTS_Phrase_Tables.tex", "w", encoding="utf-8") as f:
        f.write("\n".join(lines))

    total = sum(len(v) for v in groups.values())
    print(f"已生成 group_md/IELTS_Phrase_Tables.tex ({len(groups)} groups, {total} phrases, {COLS} columns)")


if __name__ == "__main__":
    main()
