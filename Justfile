default: gen-init lint

build:
    pyproject-build
    check-wheel-contents dist/*.whl
    twine check --strict dist/*

build-rules:
    ./scripts/build.py
    ./scripts/compile.sh

gen-init:
    ./scripts/gen-init.sh

lint: lint-python lint-toml

lint-python:
    ruff check --fix

lint-toml:
    sort-toml .ruff.toml pyproject.toml
