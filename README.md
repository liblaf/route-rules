# sing-box Rules

| Name                                       | Download Link                                                                                         |
| ------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| [📵 RuleSet: ADs](#-ruleset-ads)           | [rule-set/ads.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/ads.srs)           |
| [🔒 RuleSet: Private](#-ruleset-private)   | [rule-set/private.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/private.srs)   |
| [🇨🇳 RuleSet: CN](#-ruleset-cn)             | [rule-set/cn.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/cn.srs)             |
| [🌐 RuleSet: Proxy](#-ruleset-proxy)       | [rule-set/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/proxy.srs)       |
| [🤖 RuleSet: AI](#-ruleset-ai)             | [rule-set/ai.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/ai.srs)             |
| [🍟 RuleSet: Emby](#-ruleset-emby)         | [rule-set/emby.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/emby.srs)         |
| [☁️ RuleSet: Download](#-ruleset-download) | [rule-set/download.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/download.srs) |
| [📺 RuleSet: Media](#-ruleset-media)       | [rule-set/media.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/rule-set/media.srs)       |
| [📵 GeoSite: ADs](#-ruleset-ads)           | [geosite/ads.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/geosite/ads.srs)             |
| [🔒 GeoSite: Private](#-ruleset-private)   | [geosite/private.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/geosite/private.srs)     |
| [🇨🇳 GeoSite: CN](#-ruleset-cn)             | [geosite/cn.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/geosite/cn.srs)               |
| [🌐 GeoSite: Proxy](#-ruleset-proxy)       | [geosite/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/geosite/proxy.srs)         |
| [🇨🇳 GeoIP: CN](#-ruleset-cn)               | [geoip/cn.srs](https://github.com/liblaf/sing-box-rules/raw/rule-sets/geoip/cn.srs)                   |

- [statistics](https://github.com/liblaf/sing-box-rules/blob/rule-sets/README.md)
- `GeoSite: *` does not contain `IP-CIDR` rules, useful for DNS Rule.
- `GeoIP: *` does not contain `DOMAIN*` rules, useful for DNS Rule.

## Optimization

[optimization results](https://github.com/liblaf/sing-box-rules/blob/rule-sets/README.md)

- remove duplicate rules
- merge `DOMAIN` with `DOMAIN-SUFFIX`
- merge between `DOMAIN-SUFFIX`
- merge `DOMAIN` with `DOMAIN-KEYWORD`
- merge `DOMAIN-SUFFIX` with `DOMAIN-KEYWORD`
- merge `IP-CIDR`

### 📵 RuleSet: ADs

- include:
  - [blackmatrix7/Advertising.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Advertising)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `ads`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*-ads` | `*-ads-all` | `*@ads`

### 🔒 RuleSet: Private

- include:
  - [blackmatrix7/Lan.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Lan)
  - [blackmatrix7/NTPService.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/NTPService)
  - [DustinWin/geoip-all.db](https://github.com/DustinWin/ruleset_geodata): `private`
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `private`
  - [MetaCubeX/geoip.db](https://github.com/MetaCubeX/meta-rules-dat): `private`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `category-ntp*`, `private`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)

### 🇨🇳 RuleSet: CN

- include:
  - [blackmatrix7/ChinaMax.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/ChinaMax)
  - [blackmatrix7/Direct.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Direct)
  - [DustinWin/geoip-all.db](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [MetaCubeX/geoip.db](https://github.com/MetaCubeX/meta-rules-dat): `cn`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*-cn` | `*@cn` | `cn`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)
  - [🌐 RuleSet: Proxy](#-ruleset-proxy)

### 🌐 RuleSet: Proxy

- include:
  - [blackmatrix7/Global.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Global)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `proxy`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*!cn*`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

### 🤖 RuleSet: AI

- include:
  - [blackmatrix7/Claude.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Claude)
  - [blackmatrix7/Copilot.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Copilot)
  - [blackmatrix7/Gemini.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Gemini)
  - [blackmatrix7/OpenAI.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OpenAI)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `ai`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `openai`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)

### 🍟 RuleSet: Emby

- include:
  - [NotSFC/Emby.json](https://github.com/NotSFC/rulelist/blob/main/sing-box/Emby/Emby.json)

### ☁️ RuleSet: Download

🔴 latency insensitive, 🟢 high bandwidth

- include:
  - [blackmatrix7/Developer.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Developer)
  - [blackmatrix7/Download.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Download)
  - [blackmatrix7/OneDrive.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
  - [liblaf/download.json](https://github.com/liblaf/sing-box-rules/blob/main/custom/download.json)
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `category-dev` | `onedrive`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)

### 📺 RuleSet: Media

🟢 low latency, 🟢 high bandwidth

- include:
  - [blackmatrix7/GlobalMedia.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/GlobalMedia)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `youtube`
  - [MetaCubeX/geosite-lite.db](https://github.com/MetaCubeX/meta-rules-dat): `proxymedia`, `youtube`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `youtube`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)

## Acknowledgement

- [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
- [DustinWin/ruleset_geodata](https://github.com/DustinWin/ruleset_geodata)
- [MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)
- [NotSFC/rulelist](https://github.com/NotSFC/rulelist)
