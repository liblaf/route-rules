import aiocache

from sbr import GeoSite, Rule
from sbr.preset._ads import ads
from sbr.preset._cn import cn


@aiocache.cached()
async def download() -> Rule:
    rule = Rule()
    rule += await Rule.from_list_url("data/blackmatrix7/Developer.list")
    rule += await Rule.from_list_url("data/blackmatrix7/Download.list")
    rule += await Rule.from_list_url("data/blackmatrix7/OneDrive.list")
    rule += await Rule.from_json_url("custom/download.json")
    geosite: GeoSite = await GeoSite.from_url("data/MetaCubeX/geosite.db")
    rule += await geosite.export("category-dev")
    rule += await geosite.export("onedrive")
    rule -= await ads()
    rule -= await cn()
    return rule
