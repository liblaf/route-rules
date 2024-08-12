import asyncio

import aiocache

from sbr import GeoSite, Rule
from sbr.geoip import GeoIP
from sbr.preset._ads import ads
from sbr.preset._private import private
from sbr.preset._proxy import proxy


@aiocache.cached()
async def cn() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/ChinaMax.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Direct.list")
    geoip: GeoIP = await GeoIP.from_url("data/DustinWin/geoip-all.db")
    rule += await geoip.export("cn")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("cn")
    rule += await Rule.from_json_url("custom/cn.json")
    geoip = await GeoIP.from_url("data/MetaCubeX/geoip.db")
    rule += await geoip.export("cn")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("cn")
    categories: list[str] = [
        category
        for category in geosite.categories
        if category.endswith(("-cn", "-cn", "@cn"))
    ]
    rule: Rule = sum(
        await asyncio.gather(*[geosite.export(category) for category in categories]),
        start=rule,
    )
    rule -= await ads()
    rule -= await private()
    rule -= await proxy()
    return rule
