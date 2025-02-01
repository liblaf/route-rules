import asyncio
import fnmatch

from loguru import logger

import route_rules as rr
from route_rules import Rule, Source


async def get_rule(*spec: str) -> Rule:
    return Rule().union(*(await asyncio.gather(*(_get_rule(s) for s in spec))))


async def _get_rule(spec: str) -> Rule:
    source_name: str
    key_spec: str
    source_name, _, key_spec = spec.partition(":")
    source: Source = rr.get_source(source_name)
    keys: list[str] = []
    for k in rr.utils.split_strip(key_spec):
        if "*" in k:
            keys += fnmatch.filter(await source.keys(), k)
        else:
            keys.append(k)
    logger.debug("{} -> {}", spec, keys)
    return await source.get(*keys)
