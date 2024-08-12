import asyncio

import aiocache

from sbr import GeoSite, Rule
from sbr.preset._ads import ads
from sbr.preset._cn import cn
from sbr.preset._private import private


@aiocache.cached()
async def proxy() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Global.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("proxy")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    categories: list[str] = [
        category for category in geosite.categories if "!cn" in category
    ]
    rule: Rule = sum(
        await asyncio.gather(*[geosite.export(category) for category in categories]),
        start=rule,
    )
    rule -= await ads()
    rule -= await private()
    rule -= await cn()
    return rule
