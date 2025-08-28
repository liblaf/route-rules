import functools
from typing import Self

import attrs

from route_rules.core import RuleSet

from ._abc import Provider
from .mihomo import Behavior, Format, ProviderMihomo


@attrs.define
class ProviderRegistry:
    registry: dict[str, Provider] = attrs.field(factory=dict)

    @classmethod
    @functools.cache
    def presets(cls) -> Self:
        self: Self = cls()
        self.register(
            ProviderMihomo(
                "blackmatrix7",
                "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/{name}/{name}.list",
                "https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/{name}",
                behavior=Behavior.CLASSICAL,
                format=Format.TEXT,
            ),
            ProviderMihomo(
                "dler-io",
                "https://raw.githubusercontent.com/dler-io/Rules/main/Clash/Provider/{name}.yaml",
                "https://github.com/dler-io/Rules/blob/main/Clash/Provider/{name}.yaml",
                behavior=Behavior.CLASSICAL,
                format=Format.YAML,
            ),
            ProviderMihomo(
                "liblaf",
                "https://raw.githubusercontent.com/liblaf/route-rules/main/rules/{name}.list",
                "https://github.com/liblaf/route-rules/blob/main/rules/{name}.list",
                behavior=Behavior.DOMAIN,
                format=Format.TEXT,
            ),
            ProviderMihomo(
                "MetaCubeX/geosite",
                "https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/meta/geo/geosite/{name}.yaml",
                "https://github.com/MetaCubeX/meta-rules-dat/blob/meta/geo/geosite/{name}.yaml",
                behavior=Behavior.DOMAIN,
                format=Format.YAML,
            ),
            ProviderMihomo(
                "SukkaW/classical",
                "https://ruleset.skk.moe/Clash/{name}.txt",
                behavior=Behavior.CLASSICAL,
                format=Format.TEXT,
            ),
            ProviderMihomo(
                "SukkaW/domain",
                "https://ruleset.skk.moe/Clash/{name}.txt",
                behavior=Behavior.DOMAIN,
                format=Format.TEXT,
            ),
        )
        return self

    def download_url(self, name: str) -> str:
        provider: Provider
        ruleset_name: str
        provider, ruleset_name = self._parse_name(name)
        return provider.download_url(ruleset_name)

    async def load(self, name: str) -> RuleSet:
        provider: Provider
        ruleset_name: str
        provider, ruleset_name = self._parse_name(name)
        return await provider.load(ruleset_name)

    def preview_url(self, name: str) -> str:
        provider: Provider
        ruleset_name: str
        provider, ruleset_name = self._parse_name(name)
        return provider.preview_url(ruleset_name)

    def register(self, *providers: Provider) -> None:
        for provider in providers:
            self.registry[provider.name] = provider

    def _parse_name(self, name: str) -> tuple[Provider, str]:
        provider_name: str
        ruleset_name: str
        provider_name, _, ruleset_name = name.partition(":")
        provider: Provider = self.registry[provider_name]
        return provider, ruleset_name
