import asyncio

from icecream import ic
from sbr import GeoSite, Rule


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Advertising.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("ads")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    for category in geosite.categories:
        if category.endswith(("-ads", "-ads-all", "@ads")):
            print(category)
            rule += await geosite.export(category)
    ic(rule)
    rule.save("output/rule-set/ads.json")


if __name__ == "__main__":
    asyncio.run(main())
