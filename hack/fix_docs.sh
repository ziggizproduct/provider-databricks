#!/bin/bash

# Directory containing the files
directory=".work/databricks/databricks/docs/resources"

# Loop through each file in the directory
for file in "$directory"/*.md; do
  # Check if it's a file
  if [ -f "$file" ]; then
    # Insert a new line after the first line
    base_name=$(basename ${file})
    base_name=$(echo $base_name | cut -d"." -f1)
    echo $base_name
    sed -i '' '2a\
page_title: "databricks_'"$base_name"' Resource"\
' "$file"
    sed -i '' '2a\
description: |

' "$file"
  fi
done
