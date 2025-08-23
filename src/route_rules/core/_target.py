import asyncio
import re

import attrs

from route_rules.provider import ProviderFactoryRegistry

from ._rule_set import RuleSet
from ._statistics import Statistics


def _default_slug(self: "Target") -> str:
    slug: str = re.sub(r"[^\x00-\x7F]+", "", self.name)
    slug = slug.strip().lower()
    return slug


@attrs.define
class Target:
    name: str = attrs.field()
    providers: list[str] = attrs.field()
    slug: str = attrs.field(default=attrs.Factory(_default_slug, takes_self=True))
    factories: ProviderFactoryRegistry = attrs.field(
        factory=ProviderFactoryRegistry.presets, kw_only=True
    )

    def __attrs_post_init__(self) -> None:
        if not self.name:
            self.name = self.slug

    async def build(self) -> RuleSet:
        rule_set: RuleSet = RuleSet.union(
            *(await asyncio.gather(*(self.factories.load(p) for p in self.providers)))
        )
        rule_set = rule_set.optimize()
        return rule_set

    async def statistics(self) -> Statistics:
        statistics: Statistics = Statistics()
        for provider in self.providers:
            rule_set: RuleSet = await self.factories.load(provider)
            statistics += rule_set.statistics
        return statistics
