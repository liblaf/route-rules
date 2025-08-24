import abc
import urllib.parse

import attrs
import cachetools

from route_rules.core import RuleSet


def _default_preview_url_template(self: "Provider") -> str:
    return self.download_url_template


@attrs.define
class Provider(abc.ABC):
    name: str = attrs.field()
    download_url_template: str = attrs.field()
    preview_url_template: str = attrs.field(
        default=attrs.Factory(_default_preview_url_template, takes_self=True),
    )

    _cache: cachetools.Cache = attrs.field(
        factory=lambda: cachetools.LRUCache(maxsize=65536), kw_only=True
    )

    def download_url(self, name: str) -> str:
        name = urllib.parse.quote(name)
        return self.download_url_template.format(name=name)

    @abc.abstractmethod
    async def load(self, name: str) -> RuleSet:
        raise NotImplementedError

    def preview_url(self, name: str) -> str:
        name = urllib.parse.quote(name)
        return self.preview_url_template.format(name=name)
