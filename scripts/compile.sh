#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

for source in output/**/*.json; do
  sing-box rule-set compile "$source"
done
