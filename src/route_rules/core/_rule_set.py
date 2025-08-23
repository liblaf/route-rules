from typing import Self

import attrs

from ._statistics import Statistics


@attrs.define
class RuleSet:
    """.

    References:
        1. <https://wiki.metacubex.one/en/config/rules/>
    """

    domain: set[str] = attrs.field(factory=set)
    domain_suffix: set[str] = attrs.field(factory=set)
    domain_keyword: set[str] = attrs.field(factory=set)
    domain_wildcard: set[str] = attrs.field(factory=set)

    def __or__(self, value: Self, /) -> Self:
        fields: dict[str, set[str]] = {}
        for field in attrs.fields(type(self)):
            field: attrs.Attribute
            fields[field.name] = getattr(self, field.name) | getattr(value, field.name)
        return type(self)(**fields)

    def optimize(self) -> Self:
        # TODO: implement
        return self

    @property
    def statistics(self) -> Statistics:
        stats: Statistics = Statistics()
        for field in attrs.fields(type(self)):
            field: attrs.Attribute
            name: str = field.name.replace("_", "-").upper()
            stats[name] = len(getattr(self, field.name))
        return stats

    def union(self, *others: Self) -> Self:
        fields: dict[str, set[str]] = {}
        for field in attrs.fields(type(self)):
            field: attrs.Attribute
            fields[field.name] = getattr(self, field.name).union(
                *(getattr(other, field.name) for other in others)
            )
        return type(self)(**fields)
