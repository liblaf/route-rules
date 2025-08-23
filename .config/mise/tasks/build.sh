#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

python 'scripts/build.py'
prettier --write --ignore-path '' 'dist/README.md'
