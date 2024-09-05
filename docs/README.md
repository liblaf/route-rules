# sing-box Rules

| Name                 | Download Link                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------------ |
| 📵 RuleSet: ADs      | [rule-set/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ads.srs)           |
| 🔒 RuleSet: Private  | [rule-set/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/private.srs)   |
| 🇨🇳 RuleSet: CN       | [rule-set/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/cn.srs)             |
| 🌐 RuleSet: Proxy    | [rule-set/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/proxy.srs)       |
| 🤖 RuleSet: AI       | [rule-set/ai.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/ai.srs)             |
| 🍟 RuleSet: Emby     | [rule-set/emby.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/emby.srs)         |
| ☁️ RuleSet: Download | [rule-set/download.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/download.srs) |
| 📺 RuleSet: Media    | [rule-set/media.srs](https://github.com/liblaf/sing-box-rules/raw/sing/rule-set/media.srs)       |
| 📵 GeoSite: ADs      | [geosite/ads.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/ads.srs)             |
| 🔒 GeoSite: Private  | [geosite/private.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/private.srs)     |
| 🇨🇳 GeoSite: CN       | [geosite/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/cn.srs)               |
| 🌐 GeoSite: Proxy    | [geosite/proxy.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geosite/proxy.srs)         |
| 🇨🇳 GeoIP: CN         | [geoip/cn.srs](https://github.com/liblaf/sing-box-rules/raw/sing/geoip/cn.srs)                   |

-   [statistics](https://github.com/liblaf/sing-box-rules/blob/sing/README.md)
-   `GeoSite: *` does not contain `IP-CIDR` rules, useful for DNS Rule.
-   `GeoIP: *` does not contain `DOMAIN*` rules, useful for DNS Rule.

## Optimization

[optimization results](https://github.com/liblaf/sing-box-rules/blob/sing/README.md)

-   remove duplicate rules
-   merge `DOMAIN` with `DOMAIN-SUFFIX`
-   merge between `DOMAIN-SUFFIX`
-   merge `DOMAIN` with `DOMAIN-KEYWORD`
-   merge `DOMAIN-SUFFIX` with `DOMAIN-KEYWORD`
-   merge `IP-CIDR`

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
