import asyncio

from icecream import ic
from sbr import GeoSite, Rule


async def main() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Claude.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Copilot.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Gemini.list")
    rule += await Rule.from_list_url("data/blackmatrix7/OpenAI.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("ai")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("openai")
    rule -= await Rule.from_json_url("output/rule-set/ads.json")
    ic(rule)
    rule.save("output/rule-set/ai.json")


if __name__ == "__main__":
    asyncio.run(main())
