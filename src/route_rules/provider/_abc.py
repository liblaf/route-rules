import abc

import attrs

from route_rules.core import RuleSet


@attrs.define
class Provider(abc.ABC):
    _cache: RuleSet | None = attrs.field(default=None, kw_only=True)

    async def load(self) -> RuleSet:
        if self._cache is None:
            self._cache = await self._load()
        return self._cache

    @abc.abstractmethod
    async def _load(self) -> RuleSet: ...


@attrs.define
class ProviderFactory(abc.ABC):
    name: str = attrs.field()
    _cache: dict[str, RuleSet] = attrs.field(factory=dict, kw_only=True)

    @abc.abstractmethod
    def create(self, name: str, /) -> Provider: ...

    async def load(self, name: str, /) -> RuleSet:
        if name not in self._cache:
            self._cache[name] = await self.create(name).load()
        return self._cache[name]
