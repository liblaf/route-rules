from typing import NamedTuple

import sbr
from sbr.container import Rule


class PresetConfig(NamedTuple):
    id: str
    name: str
    include: list[str]
    exclude: list[str]


PRESETS: list[PresetConfig] = [
    PresetConfig(
        "ads",
        "🛑 ADs",
        [
            "blackmatrix7:Advertising",
            "DustinWin/geosite-all:ads",
            "MetaCubeX/geosite:*-ads,*-ads-all,*@ads",
        ],
        [],
    ),
    PresetConfig(
        "private",
        "🔒 Private",
        [
            "blackmatrix7:Lan,NTPService",
            "DustinWin/geoip-all:private",
            "DustinWin/geosite-all:private",
            "MetaCubeX/geoip:private",
            "MetaCubeX/geosite:category-ntp*,private",
        ],
        ["preset:ads"],
    ),
    PresetConfig(
        "cn",
        "🇨🇳 CN",
        [
            "blackmatrix7:ChinaMax,Direct",
            "DustinWin/geoip-all:cn",
            "DustinWin/geosite-all:cn",
            "liblaf:cn",
            "MetaCubeX/geoip:cn",
            "MetaCubeX/geosite:cn,*-cn,*@cn",
        ],
        ["preset:ads", "preset:private"],
    ),
    PresetConfig(
        "proxy",
        "✈️ Proxy",
        [
            "blackmatrix7:Global",
            "DustinWin/geosite-all:proxy",
            "MetaCubeX/geosite:*!cn*",
        ],
        ["preset:ads", "preset:cn", "preset:private"],
    ),
    PresetConfig(
        "ai",
        "🤖 AI",
        [
            "blackmatrix7:Claude,Copilot,Gemini,OpenAI",
            "DustinWin/geosite-all:ai",
            "MetaCubeX/geosite:openai",
        ],
        ["preset:ads", "preset:cn", "preset:private"],
    ),
    PresetConfig(
        "download",
        "☁️ Download",
        [
            "blackmatrix7:Download,OneDrive",
            "MetaCubeX/geosite:onedrive",
        ],
        ["preset:ads", "preset:cn", "preset:private"],
    ),
    PresetConfig(
        "emby",
        "🍟 Emby",
        ["liblaf:emby", "NotSFC:Emby"],
        ["preset:ads", "preset:cn", "preset:private"],
    ),
    PresetConfig(
        "media",
        "📺 Media",
        [
            "blackmatrix7:GlobalMedia",
            "DustinWin/geosite-all:youtube",
            "MetaCubeX/geosite-lite:proxymedia,youtube",
            "MetaCubeX/geosite:youtube",
        ],
        ["preset:ads", "preset:cn", "preset:private"],
    ),
]


async def get_preset(_id: str, *, exclude: bool = True) -> Rule:
    for cfg in PRESETS:
        if cfg.id == _id:
            return await _get_preset(cfg, exclude=exclude)
    raise KeyError(_id)


async def _get_preset(cfg: PresetConfig, *, exclude: bool = True) -> Rule:
    rule: Rule = await sbr.get_rule(*cfg.include)
    if exclude:
        rule -= await sbr.get_rule(*cfg.exclude)
    return rule
