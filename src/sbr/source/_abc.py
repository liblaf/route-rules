import abc
import asyncio

from boltons.cacheutils import LRU

from sbr import Rule


class Source(abc.ABC):
    name: str
    _key_cache: list[str] | None = None
    _rule_cache: LRU[str, Rule]

    def __init__(self) -> None:
        self._rule_cache = LRU()

    async def get(self, *key: str) -> Rule:
        return Rule().union(*(await asyncio.gather(*(self._get(k) for k in key))))

    async def keys(self) -> list[str]:
        if self._key_cache is not None:
            return self._key_cache
        self._key_cache = await self._keys_nocache()
        return self._key_cache

    @abc.abstractmethod
    async def _get_nocache(self, key: str) -> Rule: ...

    async def _get(self, key: str) -> Rule:
        if (r := self._rule_cache.get(key)) is not None:
            return r
        rule: Rule = await self._get_nocache(key)
        self._rule_cache[key] = rule
        return rule

    @abc.abstractmethod
    async def _keys_nocache(self) -> list[str]: ...
