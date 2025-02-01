from route_rules.source import ClashClassicalText, GeoIP, GeoSite, Preset, Source
from route_rules.source._singbox import SingBoxRuleSet

SOURCES: list[Source] = [
    Preset(),
    ClashClassicalText(
        "blackmatrix7",
        "https://github.com/blackmatrix7/ios_rule_script/raw/master/rule/Clash/${key}/${key}.list",
        "data/blackmatrix7/",
    ),
    GeoIP(
        "DustinWin/geoip-all",
        "https://github.com/DustinWin/ruleset_geodata/releases/download/sing-box/geoip-all.db",
        "data/DustinWin/geoip-all/",
    ),
    GeoSite(
        "DustinWin/geosite-all",
        "https://github.com/DustinWin/ruleset_geodata/releases/download/sing-box/geosite-all.db",
        "data/DustinWin/geosite-all/",
    ),
    SingBoxRuleSet(
        "liblaf",
        "https://github.com/liblaf/route-rules/raw/main/custom/${key}.json",
        "data/liblaf",
    ),
    GeoIP(
        "MetaCubeX/geoip",
        "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.db",
        "data/MetaCubeX/geoip/",
    ),
    GeoSite(
        "MetaCubeX/geosite",
        "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.db",
        "data/MetaCubeX/geosite/",
    ),
    GeoSite(
        "MetaCubeX/geosite-lite",
        "https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite-lite.db",
        "data/MetaCubeX/geosite-lite/",
    ),
    SingBoxRuleSet(
        "NotSFC",
        "https://github.com/NotSFC/rulelist/raw/main/sing-box/${key}/${key}.json",
        "data/NotSFC",
    ),
]


def get_source(name: str) -> Source:
    for source in SOURCES:
        if source.name == name:
            return source
    raise KeyError(name)
