import aiocache

from sbr import Rule


@aiocache.cached()
async def emby() -> Rule:
    rule = Rule()
    rule += await Rule.from_json_url("data/NotSFC/Emby.json")
    return rule
