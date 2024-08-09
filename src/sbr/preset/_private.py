import aiocache

from sbr import GeoSite, Rule
from sbr.geoip import GeoIP
from sbr.preset._ads import ads


@aiocache.cached()
async def private() -> Rule:
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
    rule -= await ads()
    return rule
