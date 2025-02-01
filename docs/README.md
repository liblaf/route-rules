# Route Rules

[![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/liblaf/route-rules/ci.yaml)](https://github.com/liblaf/route-rules/actions/workflows/ci.yaml)
[![GitHub last commit (branch)](https://img.shields.io/github/last-commit/liblaf/route-rules/sing?label=update)](https://github.com/liblaf/route-rules/tree/sing)
[![GitHub repo size](https://img.shields.io/github/repo-size/liblaf/route-rules)](https://github.com/liblaf/route-rules)
[![GitHub Repo stars](https://img.shields.io/github/stars/liblaf/route-rules)](https://github.com/liblaf/route-rules)
[![Read the Docs](https://img.shields.io/readthedocs/route-rules)](https://route-rules.readthedocs.io)

| Name                 | GitHub                                                                                        | Cloudflare                                                                      |
| -------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| 🛑 RuleSet: ADs      | [rule-set/ads.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/ads.srs)           | [rule-set/ads.srs](https://api.liblaf.me/rules/sing/rule-set/ads.srs)           |
| 🔒 RuleSet: Private  | [rule-set/private.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/private.srs)   | [rule-set/private.srs](https://api.liblaf.me/rules/sing/rule-set/private.srs)   |
| 🇨🇳 RuleSet: CN       | [rule-set/cn.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/cn.srs)             | [rule-set/cn.srs](https://api.liblaf.me/rules/sing/rule-set/cn.srs)             |
| ✈️ RuleSet: Proxy    | [rule-set/proxy.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/proxy.srs)       | [rule-set/proxy.srs](https://api.liblaf.me/rules/sing/rule-set/proxy.srs)       |
| 🤖 RuleSet: AI       | [rule-set/ai.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/ai.srs)             | [rule-set/ai.srs](https://api.liblaf.me/rules/sing/rule-set/ai.srs)             |
| ☁️ RuleSet: Download | [rule-set/download.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/download.srs) | [rule-set/download.srs](https://api.liblaf.me/rules/sing/rule-set/download.srs) |
| 🍟 RuleSet: Emby     | [rule-set/emby.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/emby.srs)         | [rule-set/emby.srs](https://api.liblaf.me/rules/sing/rule-set/emby.srs)         |
| 📺 RuleSet: Media    | [rule-set/media.srs](https://github.com/liblaf/route-rules/raw/sing/rule-set/media.srs)       | [rule-set/media.srs](https://api.liblaf.me/rules/sing/rule-set/media.srs)       |
| 🇨🇳 GeoIP: CN         | [geoip/cn.srs](https://github.com/liblaf/route-rules/raw/sing/geoip/cn.srs)                   | [geoip/cn.srs](https://api.liblaf.me/rules/sing/geoip/cn.srs)                   |
| 🛑 GeoSite: ADs      | [geosite/ads.srs](https://github.com/liblaf/route-rules/raw/sing/geosite/ads.srs)             | [geosite/ads.srs](https://api.liblaf.me/rules/sing/geosite/ads.srs)             |
| 🔒 GeoSite: Private  | [geosite/private.srs](https://github.com/liblaf/route-rules/raw/sing/geosite/private.srs)     | [geosite/private.srs](https://api.liblaf.me/rules/sing/geosite/private.srs)     |
| 🇨🇳 GeoSite: CN       | [geosite/cn.srs](https://github.com/liblaf/route-rules/raw/sing/geosite/cn.srs)               | [geosite/cn.srs](https://api.liblaf.me/rules/sing/geosite/cn.srs)               |
| ✈️ GeoSite: Proxy    | [geosite/proxy.srs](https://github.com/liblaf/route-rules/raw/sing/geosite/proxy.srs)         | [geosite/proxy.srs](https://api.liblaf.me/rules/sing/geosite/proxy.srs)         |

- [statistics](https://liblaf.github.io/route-rules/stats/)
- `GeoSite: *` does not contain `IP-CIDR` rules, useful for DNS Rule.
- `GeoIP: *` does not contain `DOMAIN*` rules, useful for DNS Rule.

## Optimization

[optimization results](https://liblaf.github.io/route-rules/stats/)

- remove duplicate rules
- remove unresolvable domains
- merge `DOMAIN` with `DOMAIN-SUFFIX`
- merge between `DOMAIN-SUFFIX`
- merge `DOMAIN` with `DOMAIN-KEYWORD`
- merge `DOMAIN-SUFFIX` with `DOMAIN-KEYWORD`
- merge `IP-CIDR`

## Acknowledgement

- [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
- [DustinWin/ruleset_geodata](https://github.com/DustinWin/ruleset_geodata)
- [MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)
- [NotSFC/rulelist](https://github.com/NotSFC/rulelist)
