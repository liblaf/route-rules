import re

from sbr.geosite import Geosite
from sbr.rule_set import RuleSet

URL_PREFIX: str = "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest"
geosite = Geosite(url=f"{URL_PREFIX}/geosite.db")

proxy: RuleSet = geosite.export(
    *[category for category in geosite.list() if re.match(r".*!cn$", category)]
)
proxy.save("rule-sets/proxy.srs")

cn: RuleSet = geosite.export(
    *[category for category in geosite.list() if re.match(r"(.*[-@])?cn$", category)]
)
cn.save("rule-sets/cn.srs")

ai: RuleSet = geosite.export("bing", "google", "openai", "perplexity")
ai -= geosite.export("youtube")
ai.save("rule-sets/ai.srs")

emby: RuleSet = RuleSet.from_url(
    "https://github.com/NotSFC/rulelist/raw/main/sing-box/Emby/Emby.json"
)
emby.save("rule-sets/emby.srs")

media: RuleSet = geosite.export("netflix", "youtube")
media.save("rule-sets/media.srs")
