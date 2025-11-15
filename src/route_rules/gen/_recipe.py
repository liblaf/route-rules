from __future__ import annotations

import asyncio
from collections.abc import Iterable
from typing import Self

import attrs
import cachetools
import cachetools_async as cta
from slugify import slugify

from route_rules.core import RuleSet
from route_rules.provider import ProviderRegistry

from ._config import RecipeConfig


def _default_slug(self: Recipe) -> str:
    return slugify(self.name)


@attrs.define
class Recipe:
    name: str
    providers: list[str] = attrs.field(factory=list)
    excludes: list[str] = attrs.field(factory=list)
    registry: ProviderRegistry = attrs.field(
        repr=False, factory=ProviderRegistry.presets, kw_only=True
    )
    slug: str = attrs.field(
        default=attrs.Factory(_default_slug, takes_self=True), kw_only=True
    )

    _cache: cachetools.Cache = attrs.field(
        repr=False, factory=lambda: cachetools.Cache(maxsize=1), kw_only=True
    )

    @classmethod
    def from_config(cls, config: RecipeConfig) -> Self:
        return cls(
            name=config.name, providers=config.providers, excludes=config.excludes
        )

    @cta.cachedmethod(lambda self: self._cache)
    async def build(self) -> RuleSet:
        ruleset: RuleSet = await self._load(self.providers)
        ruleset -= await self._load(self.excludes)
        ruleset = ruleset.optimize()
        return ruleset

    async def _load(self, providers: Iterable[str]) -> RuleSet:
        if not providers:
            return RuleSet()
        return RuleSet.union(
            *(await asyncio.gather(*(self.registry.load(p) for p in providers)))
        )
