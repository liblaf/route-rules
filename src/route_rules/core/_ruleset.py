import collections
from collections.abc import Iterable, Mapping
from typing import Self, override

import cytoolz as toolz

_ALIASES: dict[str, str] = {
    "IP-CIDR6": "IP-CIDR"  # IP-CIDR and IP-CIDR6 have the same effect, with IP-CIDR6 being an alias.
}


# ref: <https://wiki.metacubex.one/en/config/rules/>
class RuleSet(collections.UserDict[str, set[str]]):
    def __missing__(self, key: str) -> set[str]:
        self[key] = set()
        return self[key]

    @override
    def __or__(self, other: Mapping[str, Iterable[str]], /) -> Self:  # pyright: ignore[reportIncompatibleMethodOverride]
        return self.union(other)

    @property
    def domain(self) -> set[str]:
        return self["DOMAIN"]

    @property
    def domain_suffix(self) -> set[str]:
        return self["DOMAIN-SUFFIX"]

    @property
    def ip_cidr(self) -> set[str]:
        return self["IP-CIDR"]

    def add(self, typ: str, value: str) -> None:
        typ = _ALIASES.get(typ, typ)
        self[typ].add(value)

    def optimize(self) -> Self:
        # TODO: implement
        return self

    def union(self, *others: Mapping[str, Iterable[str]]) -> Self:
        return toolz.merge_with(
            lambda lst: set.union(*lst), self, *others, factory=type(self)
        )
