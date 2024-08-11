import asyncio

import aiocache

from sbr import GeoSite, Rule
from sbr.geoip import GeoIP
from sbr.preset._ads import ads


@aiocache.cached()
async def private() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Lan.list")
    rule += await Rule.from_list_url("data/blackmatrix7/NTPService.list")
    geoip: GeoIP = await GeoIP.from_url("data/DustinWin/geoip-all.db")
    rule += await geoip.export("private")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("private")
    geoip = await GeoIP.from_url("data/MetaCubeX/geoip.db")
    rule += await geoip.export("private")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("private")
    categories: list[str] = [
        category
        for category in geosite.categories
        if category.startswith("category-ntp")
    ]
    rule: Rule = sum(
        await asyncio.gather(*[geosite.export(category) for category in categories]),
        start=rule,
    )
    rule -= await ads()
    return rule
