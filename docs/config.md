# Example [sing-box](https://sing-box.sagernet.org) Config

## DNS Rules

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

## Route Rules

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
