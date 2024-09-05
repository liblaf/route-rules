# sing-box Rules

| Name                                       | Download Link                                                                                    |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------ |
| [📵 RuleSet: ADs](#-ruleset-ads)           | [rule-set/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ads.srs)           |
| [🔒 RuleSet: Private](#-ruleset-private)   | [rule-set/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/private.srs)   |
| [🇨🇳 RuleSet: CN](#-ruleset-cn)             | [rule-set/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/cn.srs)             |
| [🌐 RuleSet: Proxy](#-ruleset-proxy)       | [rule-set/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/proxy.srs)       |
| [🤖 RuleSet: AI](#-ruleset-ai)             | [rule-set/ai.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ai.srs)             |
| [🍟 RuleSet: Emby](#-ruleset-emby)         | [rule-set/emby.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/emby.srs)         |
| [☁️ RuleSet: Download](#-ruleset-download) | [rule-set/download.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/download.srs) |
| [📺 RuleSet: Media](#-ruleset-media)       | [rule-set/media.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/media.srs)       |
| [📵 GeoSite: ADs](#-ruleset-ads)           | [geosite/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/ads.srs)             |
| [🔒 GeoSite: Private](#-ruleset-private)   | [geosite/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/private.srs)     |
| [🇨🇳 GeoSite: CN](#-ruleset-cn)             | [geosite/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/cn.srs)               |
| [🌐 GeoSite: Proxy](#-ruleset-proxy)       | [geosite/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/proxy.srs)         |
| [🇨🇳 GeoIP: CN](#-ruleset-cn)               | [geoip/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geoip/cn.srs)                   |

- [statistics](https://github.com/liblaf/sing-box-rules/blob/sing/README.md)
- `GeoSite: *` does not contain `IP-CIDR` rules, useful for DNS Rule.
- `GeoIP: *` does not contain `DOMAIN*` rules, useful for DNS Rule.

## Optimization

[optimization results](https://github.com/liblaf/sing-box-rules/blob/sing/README.md)

- remove duplicate rules
- merge `DOMAIN` with `DOMAIN-SUFFIX`
- merge between `DOMAIN-SUFFIX`
- merge `DOMAIN` with `DOMAIN-KEYWORD`
- merge `DOMAIN-SUFFIX` with `DOMAIN-KEYWORD`
- merge `IP-CIDR`

## Example [sing-box](https://sing-box.sagernet.org) Config

### DNS Rules

```json
{
  "dns": {
    "servers": [
      {
        "tag": "dns:proxy",
        "address": "https://cloudflare-dns.com/dns-query",
        "address_resolver": "dns:bootstrap"
      },
      { "tag": "dns:local", "address": "local" },
      { "tag": "dns:reject", "address": "rcode://refused" }
    ],
    "rules": [
      { "outbound": "any", "server": "dns:local" },
      {
        "rule_set": "geosite:ads",
        "server": "dns:reject",
        "disable_cache": true
      },
      { "rule_set": "geosite:private", "server": "dns:local" },
      { "clash_mode": "direct", "server": "dns:local" },
      { "clash_mode": "global", "server": "dns:proxy" },
      { "rule_set": "geosite:cn", "server": "dns:local" },
      {
        "type": "logical",
        "mode": "and",
        "rules": [
          { "rule_set": "geosite:proxy", "invert": true },
          { "rule_set": "geoip:cn" }
        ],
        "server": "dns:proxy",
        "client_subnet": "101.6.6.6"
      }
    ],
    "final": "dns:proxy",
    "independent_cache": true
  }
}
```

### Route Rules

```json
{
  "route": {
    "rules": [
      {
        "type": "logical",
        "mode": "or",
        "rules": [{ "protocol": "dns" }, { "port": 53 }],
        "outbound": "dns"
      },
      { "rule_set": "rule-set:ads", "outbound": "REJECT" },
      {
        "ip_is_private": true,
        "rule_set": "rule-set:private",
        "outbound": "DIRECT"
      },
      { "clash_mode": "direct", "outbound": "DIRECT" },
      { "clash_mode": "global", "outbound": "PROXY" },
      {
        "type": "logical",
        "mode": "or",
        "rules": [
          { "port": 853 },
          { "network": "udp", "port": 443 },
          { "protocol": "stun" }
        ],
        "outbound": "REJECT"
      },
      {
        "type": "logical",
        "mode": "and",
        "rules": [
          { "rule_set": "rule-set:proxy", "invert": true },
          { "rule_set": "rule-set:cn" }
        ],
        "outbound": "DIRECT"
      },
      { "rule_set": "rule-set:ai", "outbound": "🤖 AI" },
      { "rule_set": "rule-set:emby", "outbound": "🍟 Emby" },
      { "rule_set": "rule-set:download", "outbound": "☁️ Download" },
      { "rule_set": "rule-set:media", "outbound": "📺 Media" }
    ],
    "final": "PROXY",
    "auto_detect_interface": true
  }
}
```

## Data Sources

> [!NOTE]
> Exclusion is implemented as simple set difference, which does not mean the difference of rule sets. For example, [🇨🇳 RuleSet: CN](#-ruleset-cn) contains `DOMAIN,www.gstatic.com`, and [🌐 RuleSet: Proxy](#-ruleset-proxy) contains `DOMAIN-SUFFIX,gstatic.com`, then after set difference, [🌐 RuleSet: Proxy](#-ruleset-proxy) can still match `www.gstatic.com`.

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
  - [liblaf/cn.json](https://github.com/liblaf/sing-box-rules/blob/main/custom/cn.json)
  - [MetaCubeX/geoip.db](https://github.com/MetaCubeX/meta-rules-dat): `cn`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*-cn` | `*@cn` | `cn`
- exclude:
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

### 🌐 RuleSet: Proxy

- include:
  - [blackmatrix7/Global.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Global)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `proxy`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `*!cn*`
- exclude:
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)
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
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

### 🍟 RuleSet: Emby

- include:
  - [NotSFC/Emby.json](https://github.com/NotSFC/rulelist/blob/main/sing-box/Emby/Emby.json)
- exclude:
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

### ☁️ RuleSet: Download

🔴 latency insensitive, 🟢 high bandwidth

- include:
  - [blackmatrix7/Developer.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Developer)
  - [blackmatrix7/Download.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/Download)
  - [blackmatrix7/OneDrive.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
  - [liblaf/download.json](https://github.com/liblaf/sing-box-rules/blob/main/custom/download.json)
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `category-dev` | `onedrive`
- exclude:
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

### 📺 RuleSet: Media

🟢 low latency, 🟢 high bandwidth

- include:
  - [blackmatrix7/GlobalMedia.list](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/GlobalMedia)
  - [DustinWin/geosite-all.db](https://github.com/DustinWin/ruleset_geodata): `youtube`
  - [MetaCubeX/geosite-lite.db](https://github.com/MetaCubeX/meta-rules-dat): `proxymedia`, `youtube`
  - [MetaCubeX/geosite.db](https://github.com/MetaCubeX/meta-rules-dat): `youtube`
- exclude:
  - [🇨🇳 RuleSet: CN](#-ruleset-cn)
  - [📵 RuleSet: ADs](#-ruleset-ads)
  - [🔒 RuleSet: Private](#-ruleset-private)

## Acknowledgement

- [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
- [DustinWin/ruleset_geodata](https://github.com/DustinWin/ruleset_geodata)
- [MetaCubeX/meta-rules-dat](https://github.com/MetaCubeX/meta-rules-dat)
- [NotSFC/rulelist](https://github.com/NotSFC/rulelist)
