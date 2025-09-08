import asyncio

import attrs
import cachetools
import cachetools_async as cta

from . import utils
from .core import RuleSet
from .provider import ProviderRegistry


@attrs.define
class Recipe:
    name: str = attrs.field()
    providers: list[str] = attrs.field()
    registry: ProviderRegistry = attrs.field(
        factory=ProviderRegistry.presets, kw_only=True
    )
    slug: str = attrs.field(
        default=attrs.Factory(utils.default_slug, takes_self=True), kw_only=True
    )

    _cache: cachetools.Cache = attrs.field(
        factory=lambda: cachetools.Cache(maxsize=1), kw_only=True
    )

    @cta.cachedmethod(lambda self: self._cache)
    async def build(self) -> RuleSet:
        ruleset: RuleSet = RuleSet.union(
            *(await asyncio.gather(*(self.registry.load(p) for p in self.providers)))
        )
        ruleset = ruleset.optimize()
        return ruleset
