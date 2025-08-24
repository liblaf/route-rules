import asyncio
import re

import attrs

from route_rules.provider import ProviderFactoryRegistry

from ._ruleset import RuleSet
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
    _cache: RuleSet | None = attrs.field(default=None, kw_only=True)

    async def build(self) -> RuleSet:
        if self._cache is None:
            ruleset: RuleSet = RuleSet.union(
                *(
                    await asyncio.gather(
                        *(self.factories.load(p) for p in self.providers)
                    )
                )
            )
            ruleset = ruleset.optimize()
            self._cache = ruleset
        return self._cache

    async def statistics(self) -> Statistics:
        statistics: Statistics = Statistics()
        for provider in self.providers:
            ruleset: RuleSet = await self.factories.load(provider)
            statistics += ruleset.statistics
        return statistics
