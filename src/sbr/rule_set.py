import pathlib
import subprocess
from typing import Annotated, Iterable, Literal

import pydantic
import requests

StrList = Annotated[
    list[str], pydantic.BeforeValidator(lambda x: [x] if isinstance(x, str) else x)
]


def merge(*lists: Iterable[str]) -> list[str]:
    result: set[str] = set()
    for lst in lists:
        result.update(lst)
    return list(result)


def diff(a: Iterable[str], b: Iterable[str]) -> list[str]:
    return list(set(a) - set(b))


class Rule(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(extra="forbid")
    domain: StrList = []
    domain_suffix: StrList = []
    domain_keyword: StrList = []
    domain_regex: StrList = []
    ip_cidr: StrList = []

    def __add__(self, other: "Rule") -> "Rule":
        return Rule(
            **{k: merge(getattr(self, k), getattr(other, k)) for k in self.keys()}
        )

    def __sub__(self, other: "Rule") -> "Rule":
        return Rule(
            **{k: diff(getattr(self, k), getattr(other, k)) for k in self.keys()}
        )

    def keys(self) -> Iterable[str]:
        return self.model_fields.keys()

    def optimize(self) -> "Rule":
        data: dict[str, set[str]] = {k: set(v) for k, v in self.model_dump().items()}
        for d in self.domain:
            if "." + d in data["domain_suffix"]:
                data["domain"].remove(d)
                data["domain_suffix"].remove("." + d)
                data["domain_suffix"].add(d)
        return Rule(**{k: list(v) for k, v in data.items()})

    @pydantic.model_serializer
    def serialize_model(self) -> dict[str, list[str]]:
        data: dict[str, list[str]] = {k: getattr(self, k) for k in self.keys()}
        data = {k: v for k, v in data.items() if v}
        return data

    @classmethod
    def sum(cls, rules: Iterable["Rule"], start: "Rule | None" = None) -> "Rule":
        if start is None:
            start = cls()
        return sum(rules, start=start)


class RuleSet(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(extra="forbid")
    version: Literal[1] = 1
    rules: list[Rule] = []

    def __add__(self, other: "RuleSet") -> "RuleSet":
        rule: Rule = Rule.sum(self.rules + other.rules)
        return RuleSet(version=self.version, rules=[rule])

    def __sub__(self, other: "RuleSet") -> "RuleSet":
        self_rule: Rule = Rule.sum(self.rules)
        other_rule: Rule = Rule.sum(other.rules)
        return RuleSet(version=self.version, rules=[self_rule - other_rule])

    def optimize(self) -> "RuleSet":
        rule: Rule = Rule.sum(self.rules)
        rule = rule.optimize()
        return RuleSet(version=self.version, rules=[rule])

    def save(self, file: pathlib.Path | str) -> None:
        file = pathlib.Path(file)
        if file.suffix == ".json":
            file.write_text(self.model_dump_json())
        elif file.suffix == ".srs":
            source_file: pathlib.Path = file.with_suffix(".json")
            self.save(source_file)
            subprocess.run(
                ["sing-box", "rule-set", "compile", source_file, "--output", file],
                stdin=subprocess.DEVNULL,
                check=True,
            )
        else:
            raise NotImplementedError

    @classmethod
    def from_url(cls, url: str) -> "RuleSet":
        resp: requests.Response = requests.get(url)
        resp.raise_for_status()
        return cls(**resp.json())
