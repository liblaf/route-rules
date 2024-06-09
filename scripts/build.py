import re

from sbr.geosite import Geosite
from sbr.rule_set import RuleSet

URL_PREFIX: str = "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest"
geosite = Geosite(url=f"{URL_PREFIX}/geosite.db")

geosite_ai: RuleSet = geosite.export("bing", "google", "openai", "perplexity")
geosite_ai.save("rule-sets/ai.srs")
geosite_emby: RuleSet = RuleSet.from_url(
    "https://github.com/NotSFC/rulelist/raw/main/sing-box/Emby/Emby.json"
)
geosite_emby.save("rule-sets/emby.srs")
geosite_onedrive: RuleSet = geosite.export("onedrive")
geosite_onedrive.save("rule-sets/onedrive.srs")
geosite_proxy: RuleSet = geosite.export(
    *[category for category in geosite.list() if re.match(r".*!cn$", category)]
)
geosite_proxy.save("rule-sets/proxy.srs")
geosite_cn: RuleSet = geosite.export(
    *[category for category in geosite.list() if re.match(r"(.*[-@])?cn$", category)]
)
geosite_cn.save("rule-sets/cn.srs")
