import collections
from typing import Self

import attrs

from ._statistics import Statistics


@attrs.define
class RuleSet:
    """.

    References:
        1. <https://wiki.metacubex.one/en/config/rules/>
    """

    data: dict[str, set[str]] = attrs.field(
        factory=lambda: collections.defaultdict(set)
    )

    def __or__(self, other: Self, /) -> Self:
        data: dict[str, set[str]] = {}
        for typ in self.data.keys() | other.data.keys():
            data[typ] = self.data.get(typ, set()) | other.data.get(typ, set())
        return type(self)(data=data)

    @property
    def domain(self) -> set[str]:
        return self.data["DOMAIN"]

    @property
    def domain_suffix(self) -> set[str]:
        return self.data["DOMAIN-SUFFIX"]

    @property
    def statistics(self) -> Statistics:
        stats: Statistics = Statistics()
        for typ, values in self.data.items():
            stats[typ] = len(values)
        return stats

    @property
    def total(self) -> int:
        return sum(len(v) for v in self.data.values())

    def add(self, typ: str, value: str) -> None:
        self.data[typ].add(value)

    def optimize(self) -> Self:
        # TODO: implement
        return self

    def union(self, *others: Self) -> Self:
        data: dict[str, set[str]] = {}
        for typ in set(self.data.keys()).union(
            *(other.data.keys() for other in others)
        ):
            data[typ] = self.data.get(typ, set()).union(
                *(other.data.get(typ, set()) for other in others)
            )
        return type(self)(data=data)
