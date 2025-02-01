import operator
from collections.abc import Callable
from typing import Annotated, Any

import pydantic
from pydantic import BaseModel, ConfigDict

import route_rules as rr
from route_rules.container import optim
from route_rules.typing import StrPath

Set = Annotated[set[str], pydantic.BeforeValidator(rr.utils.as_set)]


class Rule(BaseModel):
    model_config = ConfigDict(extra="forbid")
    domain: Set = set()
    domain_suffix: Set = set()
    domain_keyword: Set = set()
    domain_regex: Set = set()
    ip_cidr: Set = set()

    @classmethod
    def from_file(cls, path: StrPath) -> "Rule":
        rule_set: rr.RuleSet = rr.RuleSet.from_file(path)
        return Rule().union(*rule_set.rules)

    def __getitem__(self, key: str) -> set[str]:
        return getattr(self, key)

    def __len__(self) -> int:
        return sum(len(v) for k, v in self)

    def __or__(self, other: "Rule") -> "Rule":
        return self.op(operator.or_, other)

    def __sub__(self, other: "Rule") -> "Rule":
        return self.op(operator.sub, other)

    def union(self, *others: "Rule") -> "Rule":
        return Rule(**{k: v.union(*(r[k] for r in others)) for k, v in self})

    def difference(self, *others: "Rule") -> "Rule":
        return Rule(**{k: v.difference(*(r[k] for r in others)) for k, v in self})

    def geoip(self) -> "Rule":
        return Rule(ip_cidr=self.ip_cidr)

    def geosite(self) -> "Rule":
        return Rule(
            domain=self.domain,
            domain_suffix=self.domain_suffix,
            domain_keyword=self.domain_keyword,
            domain_regex=self.domain_regex,
        )

    def op(self, op: Callable[[Any, Any], Any], other: "Rule") -> "Rule":
        return Rule(**{k: op(self[k], other[k]) for k, v in self})

    def optimize(self) -> None:
        self.domain = optim.remove_unresolvable(self.domain)
        self.domain, self.domain_suffix = optim.merge_domain_with_suffix(
            self.domain, self.domain_suffix
        )
        self.domain_suffix = optim.merge_between_suffix(self.domain_suffix)
        self.domain, self.domain_keyword = optim.merge_domain_with_keyword(
            self.domain, self.domain_keyword
        )
        self.domain_suffix, self.domain_keyword = optim.merge_suffix_with_keyword(
            self.domain_suffix, self.domain_keyword
        )
        self.ip_cidr = optim.merge_ip_cidr(self.ip_cidr)

    def save(self, path: StrPath) -> None:
        rr.RuleSet(version=1, rules=[self]).save(path)

    def summary(self) -> str:
        res: str = ""
        for k, v in self:
            if v:
                name: str = k.upper().replace("_", "-")
                res += f"{name}: {len(v)}\n"
        res += f"TOTAL: {len(self)}"
        return res
