import aiocache

from sbr import GeoSite, Rule
from sbr.preset._ads import ads


@aiocache.cached()
async def ai() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Claude.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Copilot.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Gemini.list")
    rule += await Rule.from_list_url("data/blackmatrix7/OpenAI.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("ai")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("openai")
    rule -= await ads()
    return rule
