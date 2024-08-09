import asyncio

import aiocache

from sbr import GeoSite, Rule


@aiocache.cached()
async def ads() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Advertising.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("ads")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    categories: list[str] = [
        category
        for category in geosite.categories
        if category.endswith(("-ads", "-ads-all", "@ads"))
    ]
    rule: Rule = sum(
        await asyncio.gather(*[geosite.export(category) for category in categories]),
        start=rule,
    )
    return rule
