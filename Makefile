default: fmt

fmt: fmt-toml

fmt-toml: fmt-toml/pyproject.toml

fmt-toml/%: %
	toml-sort --in-place --all "$<"
	taplo fmt "$<"
