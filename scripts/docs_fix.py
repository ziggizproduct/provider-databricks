#!/usr/bin/env python3

import glob
from pathlib import Path

# usage: version_diff.py <generated resource list> <base JSON schema path> <bumped JSON schema path>
# example usage: version_diff.py config/generated.lst .work/schema.json.3.38.0 config/schema.json
if __name__ == "__main__":
    directory = ".work/databricks/databricks/docs/resources"
    files = glob.glob(f"{directory}/*.md")
    for file_path in files:
        base_name = Path(file_path).stem
        # Open the md file and add headers for page_tile and description
        with open(file_path, encoding="utf-8") as file:
            lines = file.readlines()

        if any("page_title:" in s for s in lines):
            continue

        lines.insert(1, f'page_title: "databricks_{base_name} Resource"\n')
        lines.insert(2, "description: |-\n")
        lines.pop(3)

        with open(file_path, "w") as file:
            file.writelines(lines)

        print(f"{base_name} fixed.")
