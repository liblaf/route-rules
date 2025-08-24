import os
from pathlib import Path
from typing import override

import attrs
import httpx

from route_rules import utils
from route_rules.core import RuleSet
from route_rules.provider._abc import Provider, ProviderFactory

from ._decode import decode
from ._encode import encode
from ._enum import Behavior, Format


@attrs.define
class ProviderMihomo(Provider):
    url: str = attrs.field()
    behavior: Behavior = attrs.field(kw_only=True)
    format: Format = attrs.field(default=Format.YAML, kw_only=True)

    @classmethod
    def save(
        cls,
        file: str | os.PathLike[str],
        ruleset: RuleSet,
        *,
        behavior: Behavior,
        format: Format = Format.YAML,  # noqa: A002
    ) -> None:
        data: bytes = encode(ruleset, behavior=behavior, format=format)
        file = Path(file)
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_bytes(data)

    @override
    async def load(self) -> RuleSet:
        response: httpx.Response = await utils.download(self.url)
        return decode(response.text, self.behavior, self.format)


@attrs.define
class ProviderMihomoFactory(ProviderFactory):
    name: str = attrs.field()
    url: str = attrs.field()
    behavior: Behavior = attrs.field(kw_only=True)
    format: Format = attrs.field(default=Format.YAML, kw_only=True)

    @override
    def create(self, name: str, /) -> ProviderMihomo:
        return ProviderMihomo(
            url=self.url.format(name=name),
            behavior=self.behavior,
            format=self.format,
        )
