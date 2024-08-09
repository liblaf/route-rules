import aiocache

from sbr import GeoSite, Rule
from sbr.preset._ads import ads


@aiocache.cached()
async def media() -> None:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/GlobalMedia.list")
    geosite: GeoSite = await GeoSite.from_url("data/DustinWin/geosite-all.db")
    rule += await geosite.export("youtube")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite-lite.db")
    rule += await geosite.export("proxymedia")
    rule += await geosite.export("youtube")
    geosite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("youtube")
    rule -= await ads()
    return rule
