import functools
from typing import Self

import attrs

from route_rules.core import RuleSet

from ._abc import Provider, ProviderFactory
from ._mihomo import Behavior, Format, ProviderMihomoFactory


@attrs.define
class ProviderFactoryRegistry:
    registry: dict[str, ProviderFactory] = attrs.field(factory=dict)

    @classmethod
    @functools.lru_cache
    def presets(cls) -> Self:
        self: Self = cls()
        self.register(
            ProviderMihomoFactory(
                "MetaCubeX/geosite",
                "https://cdn.jsdelivr.net/gh/MetaCubeX/meta-rules-dat@meta/geo/geosite/{name}.yaml",
                behavior=Behavior.DOMAIN,
                format=Format.YAML,
            )
        )
        return self

    async def load(self, provider: str) -> RuleSet:
        factory_name: str
        rule_set_name: str
        factory_name, _, rule_set_name = provider.partition(":")
        factory: ProviderFactory = self.registry[factory_name]
        provider: Provider = factory.create(rule_set_name)
        return await provider.load()

    def register(self, factory: ProviderFactory) -> None:
        self.registry[factory.name] = factory
