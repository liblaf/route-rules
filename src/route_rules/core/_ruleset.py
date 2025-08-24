import collections
from collections.abc import Mapping
from collections.abc import Set as AbstractSet
from typing import Self, override


class RuleSet(collections.UserDict[str, set[str]]):
    """.

    References:
        1. <https://wiki.metacubex.one/en/config/rules/>
    """

    @override
    def __or__(self, other: Mapping[str, AbstractSet[str]], /) -> Self:  # pyright: ignore[reportIncompatibleMethodOverride]
        result: Self = type(self)()
        for typ in self.keys() | other.keys():
            result[typ] = self.get(typ, set()) | other.get(typ, set())
        return result

    def __missing__(self, key: str) -> set[str]:
        self[key] = set()
        return self[key]

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
        # IP-CIDR and IP-CIDR6 have the same effect, with IP-CIDR6 being an alias.
        if typ == "IP-CIDR6":
            typ = "IP-CIDR"
        self[typ].add(value)

    def optimize(self) -> Self:
        # TODO: implement
        return self

    def union(self, *others: Mapping[str, AbstractSet[str]]) -> Self:
        result: Self = type(self)()
        for typ in set(self.keys()).union(*(m.keys() for m in others)):
            result[typ] = self.get(typ, set()).union(
                *(m.get(typ, set()) for m in others)
            )
        return result
