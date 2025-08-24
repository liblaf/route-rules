#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

python 'scripts/build.py'
python 'scripts/render-docs.py'
prettier --write --ignore-path '' 'dist/**/*.md'
