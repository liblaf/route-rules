import asyncio

from icecream import ic
from sbr import GeoSite, Rule
from sbr.geoip import GeoIP


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/ChinaMax.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Direct.list")
    geoip: GeoIP = await GeoIP.from_url("data/DustinWin/geoip-all.db")
    rule += await geoip.export("cn")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("cn")
    geoip = await GeoIP.from_url("data/MetaCubeX/geoip.db")
    rule += await geoip.export("cn")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("cn")
    for category in geosite.categories:
        if category.endswith(("-cn", "-cn", "@cn")) or "-ntp" in category:
            print(category)
            rule += await geosite.export(category)
    rule -= await Rule.from_json_url("output/rule-set/ads.json")
    rule -= await Rule.from_json_url("output/rule-set/private.json")
    ic(rule)
    rule.save("output/rule-set/cn.json")


if __name__ == "__main__":
    asyncio.run(main())
