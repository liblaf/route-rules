import asyncio

from icecream import ic
from sbr import GeoSite, Rule


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/GlobalMedia.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("youtube")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite-lite.db")
    rule += await geosite.export("proxymedia")
    rule += await geosite.export("youtube")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("youtube")
    rule -= await Rule.from_json_url("output/rule-set/ads.json")
    ic(rule)
    rule.save("output/rule-set/media.json")


if __name__ == "__main__":
    asyncio.run(main())
