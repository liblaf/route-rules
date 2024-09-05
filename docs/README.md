# sing-box Rules

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/liblaf/sing-box-rules/docs.yaml?label=docs)](https://liblaf.github.io/sing-box-rules/)
[![GitHub last commit (branch)](https://img.shields.io/github/last-commit/liblaf/sing-box-rules/sing?label=update)](https://github.com/liblaf/sing-box-rules/tree/sing)
[![GitHub repo size](https://img.shields.io/github/repo-size/liblaf/sing-box-rules)](https://github.com/liblaf/sing-box-rules)
[![GitHub Repo stars](https://img.shields.io/github/stars/liblaf/sing-box-rules)](https://github.com/liblaf/sing-box-rules)

| Name                 | Download Link                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------------ |
| 🛑 RuleSet: ADs      | [rule-set/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ads.srs)           |
| 🔒 RuleSet: Private  | [rule-set/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/private.srs)   |
| 🇨🇳 RuleSet: CN       | [rule-set/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/cn.srs)             |
| ✈️ RuleSet: Proxy    | [rule-set/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/proxy.srs)       |
| 🤖 RuleSet: AI       | [rule-set/ai.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ai.srs)             |
| ☁️ RuleSet: Download | [rule-set/download.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/download.srs) |
| 🍟 RuleSet: Emby     | [rule-set/emby.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/emby.srs)         |
| 📺 RuleSet: Media    | [rule-set/media.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/media.srs)       |
| 🇨🇳 GeoIP: CN         | [geoip/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geoip/cn.srs)                   |
| 🛑 GeoSite: ADs      | [geosite/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/ads.srs)             |
| 🔒 GeoSite: Private  | [geosite/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/private.srs)     |
| 🇨🇳 GeoSite: CN       | [geosite/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/cn.srs)               |
| ✈️ GeoSite: Proxy    | [geosite/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/proxy.srs)         |

-   [statistics](https://liblaf.github.io/sing-box-rules/stats/)
-   `GeoSite: *` does not contain `IP-CIDR` rules, useful for DNS Rule.
-   `GeoIP: *` does not contain `DOMAIN*` rules, useful for DNS Rule.

## Optimization

[optimization results](https://liblaf.github.io/sing-box-rules/stats/)

-   remove duplicate rules
-   remove unresolvable domains
-   merge `DOMAIN` with `DOMAIN-SUFFIX`
-   merge between `DOMAIN-SUFFIX`
-   merge `DOMAIN` with `DOMAIN-KEYWORD`
-   merge `DOMAIN-SUFFIX` with `DOMAIN-KEYWORD`
-   merge `IP-CIDR`

## Acknowledgement

-   [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
-   [DustinWin/ruleset_geodata](https://github.com/DustinWin/ruleset_geodata)
-   [MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)
-   [NotSFC/rulelist](https://github.com/NotSFC/rulelist)
