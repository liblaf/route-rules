import asyncio

from icecream import ic
from sbr import GeoSite, Rule
from sbr.geoip import GeoIP


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Lan.list")
    geoip: GeoIP = await GeoIP.from_url("data/DustinWin/geoip-all.db")
    rule += await geoip.export("private")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("private")
    geoip = await GeoIP.from_url("data/MetaCubeX/geoip.db")
    rule += await geoip.export("private")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("private")
    rule -= await Rule.from_json_url("output/rule-set/ads.json")
    ic(rule)
    rule.save("output/rule-set/private.json")


if __name__ == "__main__":
    asyncio.run(main())
