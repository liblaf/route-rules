import functools
from typing import Self

import attrs

from route_rules.core import RuleSet

from ._abc import ProviderFactory
from .mihomo import Behavior, Format, ProviderMihomoFactory


@attrs.define
class ProviderFactoryRegistry:
    registry: dict[str, ProviderFactory] = attrs.field(factory=dict)
    _cache: dict[str, RuleSet] = attrs.field(factory=dict, kw_only=True)

    @classmethod
    @functools.cache
    def presets(cls) -> Self:
        self: Self = cls()
        self.register(
            ProviderMihomoFactory(
                "blackmatrix7",
                "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/{name}/{name}.list",
                behavior=Behavior.CLASSICAL,
                format=Format.TEXT,
            ),
            ProviderMihomoFactory(
                "dler-io",
                "https://raw.githubusercontent.com/dler-io/Rules/main/Clash/Provider/{name}.yaml",
                behavior=Behavior.CLASSICAL,
                format=Format.YAML,
            ),
            ProviderMihomoFactory(
                "MetaCubeX/geosite",
                "https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/meta/geo/geosite/{name}.yaml",
                behavior=Behavior.DOMAIN,
                format=Format.YAML,
            ),
            ProviderMihomoFactory(
                "SukkaW/classical",
                "https://ruleset.skk.moe/Clash/{name}.txt",
                behavior=Behavior.CLASSICAL,
                format=Format.TEXT,
            ),
            ProviderMihomoFactory(
                "SukkaW/domain",
                "https://ruleset.skk.moe/Clash/{name}.txt",
                behavior=Behavior.DOMAIN,
                format=Format.TEXT,
            ),
        )
        return self

    async def load(self, provider: str) -> RuleSet:
        if provider not in self._cache:
            factory_name: str
            ruleset_name: str
            factory_name, _, ruleset_name = provider.partition(":")
            factory: ProviderFactory = self.registry[factory_name]
            self._cache[provider] = await factory.load(ruleset_name)
        return self._cache[provider]

    def register(self, *factories: ProviderFactory) -> None:
        for factory in factories:
            self.registry[factory.name] = factory
