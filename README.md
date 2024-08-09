# sing-box Rules

## Rules

### RuleSet

#### 📵 RuleSet:ADs

- include:
  - [`blackmatrix7/Advertising.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Advertising)
  - [`DustinWin/geosite-all.db`](https://github.com/DustinWin/ruleset_geodata): `ads`
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `*-ads` | `*-ads-all` | `*@ads`

#### 🔒 RuleSet:Private

- include:
  - [`blackmatrix7/Lan.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Lan)
  - [`DustinWin/geoip-all.db`](https://github.com/DustinWin/ruleset_geodata): `private`
  - [`DustinWin/geosite-all.db`](https://github.com/DustinWin/ruleset_geodata): `private`
  - [`MetaCubeX/geoip.db`](https://github.com/MetaCubeX/meta-rules-dat): `private`
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `private`
- exclude:
  - [📵 RuleSet:ADs](#-rulesetads)

#### 🇨🇳 RuleSet:CN

- include:
  - [`blackmatrix7/ChinaMax.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/ChinaMax)
  - [`blackmatrix7/Direct.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Direct)
  - [`blackmatrix7/NTPService.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/NTPService)
  - [`DustinWin/geoip-all.db`](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [`DustinWin/geosite-all.db`](https://github.com/DustinWin/ruleset_geodata): `cn`
  - [`MetaCubeX/geoip.db`](https://github.com/MetaCubeX/meta-rules-dat): `cn`
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `*-cn` | `*-ntp*` | `*@cn` | `cn`
- exclude:
  - [📵 RuleSet:ADs](#-rulesetads)
  - [🔒 RuleSet:Private](#-rulesetprivate)

#### 🤖 RuleSet:AI

- include:
  - [`blackmatrix7/Claude.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Claude)
  - [`blackmatrix7/Copilot.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Copilot)
  - [`blackmatrix7/Gemini.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Gemini)
  - [`blackmatrix7/OpenAI.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OpenAI)
  - [`DustinWin/geosite-all.db`](https://github.com/DustinWin/ruleset_geodata): `ai`
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `openai`
- exclude:
  - [📵 RuleSet:ADs](#-rulesetads)

#### 📺 RuleSet:Media

🟢 low latency, 🟢 high bandwidth

- include:
  - [`blackmatrix7/GlobalMedia.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/GlobalMedia)
  - [`DustinWin/geosite-all.db`](https://github.com/DustinWin/ruleset_geodata): `youtube`
  - [`MetaCubeX/geosite-lite.db`](https://github.com/MetaCubeX/meta-rules-dat): `proxymedia`, `youtube`
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `youtube`
- exclude:
  - [📵 RuleSet:ADs](#-rulesetads)

#### ☁️ RuleSet:Download

🔴 latency insensitive, 🟢 high bandwidth

- include:
  - [`blackmatrix7/Developer.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Developer)
  - [`blackmatrix7/Download.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Download)
  - [`blackmatrix7/OneDrive.list`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
  - [`liblaf/download.json`](https://github.com/liblaf/sing-box-rules/blob/main/custom/download.json)
  - [`MetaCubeX/geosite.db`](https://github.com/MetaCubeX/meta-rules-dat): `category-dev` | `onedrive`
- exclude:
  - [📵 RuleSet:ADs](#-rulesetads)
  - [🇨🇳 RuleSet:CN](#-rulesetcn)

#### 🍟 RuleSet:Emby

- include:
  - [`NotSFC/Emby.json`](https://github.com/NotSFC/rulelist/blob/main/sing-box/Emby/Emby.json)

### GeoSite

#### 🔒 GeoSite:Private

#### 🇨🇳 GeoSite:CN
