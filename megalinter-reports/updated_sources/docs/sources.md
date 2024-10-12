# Data Sources

!!! note

    Exclusion is implemented as simple set difference, which does not mean the difference of rule sets. For example, [🇨🇳 CN](#cn) contains `DOMAIN,www.gstatic.com`, and [✈️ Proxy](#proxy) contains `DOMAIN-SUFFIX,gstatic.com`, then after set difference (`proxy -= cn`), [✈️ Proxy](#proxy) can still match `www.gstatic.com`.

## 🛑 ADs

- include:
  - [blackmatrix7/Advertising.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Advertising)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `ads`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*-ads` | `*-ads-all` | `*@ads`

## 🔒 Private

- include:
  - [blackmatrix7/Lan.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Lan)
  - [blackmatrix7/NTPService.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/NTPService)
  - [DustinWin/geoip-all.db](https://github.com/DustinWin/ruleset_geodata): `private`
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `private`
  - [MetaCubeX/geoip.db](https://github.com/MetaCubeX/meta-rules-dat): `private`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `category-ntp*`, `private`
- exclude:
  - [🛑 ADs](#ads)

## 🇨🇳 CN

- include:
  - [blackmatrix7/ChinaMax.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/ChinaMax)
  - [blackmatrix7/Direct.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Direct)
  - [DustinWin/geoip-all.db](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [liblaf/cn.json](https://github.com/liblaf/sing-box-rules/blob/main/custom/cn.json)
  - [MetaCubeX/geoip.db](https://github.com/MetaCubeX/meta-rules-dat): `cn`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*-cn` | `*@cn` | `cn`
- exclude:
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## ✈️ Proxy

- include:
  - [blackmatrix7/Global.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Global)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `proxy`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*!cn*`
- exclude:
  - [🇨🇳 CN](#cn)
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## 🤖 AI

- include:
  - [blackmatrix7/Claude.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Claude)
  - [blackmatrix7/Copilot.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Copilot)
  - [blackmatrix7/Gemini.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Gemini)
  - [blackmatrix7/OpenAI.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OpenAI)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `ai`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `openai`
- exclude:
  - [🇨🇳 CN](#cn)
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## ☁️ Download

🔴 latency insensitive, 🟢 high bandwidth

- include:
  - [blackmatrix7/Download.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Download)
  - [blackmatrix7/OneDrive.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `onedrive`
- exclude:
  - [🇨🇳 CN](#cn)
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## 🍟 Emby

- include:
  - [NotSFC/Emby.json](https://github.com/NotSFC/rulelist/blob/main/sing-box/Emby/Emby.json)
- exclude:
  - [🇨🇳 CN](#cn)
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## 📺 Media

🟢 low latency, 🟢 high bandwidth

- include:
  - [blackmatrix7/GlobalMedia.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/GlobalMedia)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `youtube`
  - [MetaCubeX/geosite-lite.db](https://github.com/MetaCubeX/meta-rules-dat): `proxymedia`, `youtube`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `youtube`
- exclude:
  - [🇨🇳 CN](#cn)
  - [🛑 ADs](#ads)
  - [🔒 Private](#private)

## Acknowledgement

- [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
- [DustinWin/ruleset_geodata](https://github.com/DustinWin/ruleset_geodata)
- [MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)
- [NotSFC/rulelist](https://github.com/NotSFC/rulelist)
