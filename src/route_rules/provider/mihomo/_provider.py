from typing import override

import attrs
import cachetools_async as cta
import httpx

from route_rules import utils
from route_rules.core import RuleSet
from route_rules.provider._abc import Provider

from ._decode import decode
from ._enum import Behavior, Format


@attrs.define
class ProviderMihomo(Provider):
    behavior: Behavior = attrs.field(kw_only=True)
    format: Format = attrs.field(default=Format.YAML, kw_only=True)

    @override
    @cta.cachedmethod(lambda self: self._cache)
    async def load(self, name: str) -> RuleSet:
        response: httpx.Response = await utils.download(self.download_url(name))
        return decode(response.text, behavior=self.behavior, format=self.format)
