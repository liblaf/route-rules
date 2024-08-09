import asyncio

from icecream import ic
from sbr import GeoSite, Rule


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Developer.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Download.list")
    rule += await Rule.from_list_url("data/blackmatrix7/OneDrive.list")
    rule += await Rule.from_json_url("custom/download.json")
    geosite: GeoSite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("category-dev")
    rule += await geosite.export("onedrive")
    rule -= await Rule.from_json_url("output/rule-set/ads.json")
    rule -= await Rule.from_json_url("output/rule-set/cn.json")
    ic(rule)
    rule.save("output/rule-set/download.json")


if __name__ == "__main__":
    asyncio.run(main())
