import json
import os
import pathlib
from typing import Annotated, Literal

import pydantic

from sbr import utils
from sbr.typing import StrPath

StrSet = Annotated[set[str], pydantic.BeforeValidator(utils.as_list)]


class Rule(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(
        extra="forbid",
    )
    domain: StrSet = set()
    domain_suffix: StrSet = set()
    domain_keyword: StrSet = set()
    domain_regex: StrSet = set()
    ip_cidr: StrSet = set()
    process_name: StrSet = set()

    @classmethod
    def from_json(cls, text: str) -> "Rule":
        rule_set = RuleSet(**json.loads(text))
        return rule_set.rule

    @classmethod
    async def from_json_url(cls, url: StrPath) -> "Rule":
        text: str = await utils.text_from_url(url)
        return Rule.from_json(text)

    @classmethod
    def from_list(cls, text: str) -> "Rule":
        rule = Rule()
        for line in utils.strip_comments(text):
            words: list[str] = line.split(",")
            match words[0]:
                case "DOMAIN":
                    rule.domain.add(words[1])
                case "DOMAIN-SUFFIX":
                    rule.domain_suffix.add(words[1])
                case "DOMAIN-KEYWORD":
                    rule.domain_keyword.add(words[1])
                case "DOMAIN-REGEX":
                    rule.domain_regex.add(words[1])
                case "IP-CIDR":
                    rule.ip_cidr.add(words[1])
                case "IP-CIDR6":
                    rule.ip_cidr.add(words[1])
                case "IP-ASN":
                    pass  # TODO
                case "PROCESS-NAME":
                    rule.process_name.add(words[1])
                case _:
                    msg: str = f"Unknown rule: {line}"
                    raise ValueError(msg)
        return rule

    @classmethod
    async def from_list_url(cls, url: StrPath) -> "Rule":
        text: str = await utils.text_from_url(url)
        return Rule.from_list(text)

    def __add__(self, other: "Rule") -> "Rule":
        return Rule(
            domain=self.domain | other.domain,
            domain_suffix=self.domain_suffix | other.domain_suffix,
            domain_keyword=self.domain_keyword | other.domain_keyword,
            domain_regex=self.domain_regex | other.domain_regex,
            ip_cidr=self.ip_cidr | other.ip_cidr,
            process_name=self.process_name | other.process_name,
        )

    def __and__(self, other: "Rule") -> "Rule":
        return Rule(
            domain=self.domain & other.domain,
            domain_suffix=self.domain_suffix & other.domain_suffix,
            domain_keyword=self.domain_keyword & other.domain_keyword,
            domain_regex=self.domain_regex & other.domain_regex,
            ip_cidr=self.ip_cidr & other.ip_cidr,
            process_name=self.process_name & other.process_name,
        )

    def __len__(self) -> int:
        return self.size

    def __repr__(self) -> str:
        return self.summary

    def __str__(self) -> str:
        return self.summary

    def __sub__(self, other: "Rule") -> "Rule":
        return Rule(
            domain=self.domain - other.domain,
            domain_suffix=self.domain_suffix - other.domain_suffix,
            domain_keyword=self.domain_keyword - other.domain_keyword,
            domain_regex=self.domain_regex - other.domain_regex,
            ip_cidr=self.ip_cidr - other.ip_cidr,
            process_name=self.process_name - other.process_name,
        )

    def save(self, filename: str | os.PathLike[str]) -> None:
        filename = pathlib.Path(filename)
        rule_set = RuleSet(version=1, rules=[self])
        text: str = rule_set.model_dump_json(exclude_defaults=True)
        filename.parent.mkdir(parents=True, exist_ok=True)
        filename.write_text(text)

    @property
    def size(self) -> int:
        return sum(
            [
                len(self.domain),
                len(self.domain_suffix),
                len(self.domain_keyword),
                len(self.domain_regex),
                len(self.ip_cidr),
                len(self.process_name),
            ]
        )

    @property
    def summary(self) -> str:
        res: str = ""
        if self.domain:
            res += f"DOMAIN: {len(self.domain)}\n"
        if self.domain_suffix:
            res += f"DOMAIN-SUFFIX: {len(self.domain_suffix)}\n"
        if self.domain_keyword:
            res += f"DOMAIN-KEYWORD: {len(self.domain_keyword)}\n"
        if self.domain_regex:
            res += f"DOMAIN-REGEX: {len(self.domain_regex)}\n"
        if self.ip_cidr:
            res += f"IP-CIDR: {len(self.ip_cidr)}\n"
        if self.process_name:
            res += f"PROCESS-NAME: {len(self.process_name)}\n"
        res += f"TOTAL: {self.size}"
        return res


class RuleSet(pydantic.BaseModel):
    model_config = pydantic.ConfigDict(extra="forbid")
    version: Literal[1, 2]
    rules: list[Rule]

    @property
    def rule(self) -> Rule:
        return sum(self.rules, start=Rule())
