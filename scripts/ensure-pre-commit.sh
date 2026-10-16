#!/usr/bin/env bash
# Workspace standard, copied from ziggizproduct/arch. Do not edit in product
# repos. Installs pre-commit (if missing) and its git hook, so every commit in
# this checkout runs the formatting and linting standard. Run it once after
# cloning; agents run it at the start of every session (AGENTS.md).
# Never fails: a session must start even when offline.
set -uo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)" || exit 0

if ! command -v pre-commit >/dev/null 2>&1; then
  if ! command -v uv >/dev/null 2>&1; then
    echo "pre-commit is not installed and uv is missing; install uv (https://docs.astral.sh/uv/), then run this script again."
    exit 0
  fi
  uv tool install --quiet pre-commit >/dev/null 2>&1
  PATH="$(uv tool dir --bin):$PATH"
fi

if ! command -v pre-commit >/dev/null 2>&1; then
  echo "pre-commit is not installed and could not be installed; run 'uv tool install pre-commit' before committing."
  exit 0
fi

if pre-commit install >/dev/null 2>&1; then
  echo "pre-commit git hook is installed: commits run the formatting and linting standard. Run 'pre-commit run --files <paths>' before committing to fix issues early."
else
  echo "Could not install the pre-commit git hook; run 'pre-commit install'."
fi

if grep -q 'filter=lfs' compliance/sources/.gitattributes 2>/dev/null; then
  if git lfs install --local >/dev/null 2>&1; then
    git lfs pull >/dev/null 2>&1 || echo "Could not fetch Git LFS files; run 'git lfs pull'."
  else
    echo "This repo stores files in Git LFS and git-lfs is not installed; install it (docs/standards/engineering/formatting-and-linting.md, Tools you install), then run this script again."
  fi
fi
exit 0
