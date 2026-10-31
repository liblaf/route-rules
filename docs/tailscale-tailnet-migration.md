# Migrate Tailnet and direct providers

The `lan`, `cn`, and public `tailscale` groups are merged into `direct`, which uses ordinary `DIRECT`. Private Tailnet resources stay in `tailnet`. Profiles still using the original private `tailscale` group must migrate that coverage to `tailnet`. These changes affect provider meaning and artifact names.

| Group | Maintained contents | Generated artifacts | Consumer outbound |
| --- | --- | --- | --- |
| `tailnet` | `+.ts.net`; exact `forgejo.liblaf.me`, `jellyfin.liblaf.me`, `oplist.liblaf.me`, `qbt.liblaf.me`; `100.64.0.0/10`; `fd7a:115c:a1e0::/48` | `tailnet.domain.list`, `tailnet.domain.mrs`, `tailnet.ipcidr.list`, `tailnet.ipcidr.mrs` | `TAILSCALE`, a direct outbound bound to `tailscale0` |
| `direct` | Former LAN and China sources; `+.tailscale.com`, `+.tailscale.io`; Cloudflare Tunnel destination port 7844; local and China IP ranges with Tailnet ranges subtracted | `direct.domain.list`, `direct.domain.mrs`, `direct.ipcidr.list`, `direct.ipcidr.mrs`, `direct.classical.txt` | Ordinary `DIRECT` |

The generated priority is `tailnet`, `direct`, `us`, `global`. The `direct` set excludes `tailnet`; `us` and `global` exclude both. In particular, the public Tailscale suffixes are removed from `global` by the existing exclusion mechanism. Consumer profiles must follow that priority because a broader domain or classical rule can still overlap part of an earlier group's match space.

The existing `+.liblaf.me` override, now in `direct`, is one such partial overlap: the four exact custom hosts select `tailnet` first, while the broader suffix remains. Generation asserts their routing priority and asserts exclusive membership for `.ts.net`, the private IP ranges, and the public Tailscale domains.

## Provider and DNS migration

Apply these changes together when consuming the new artifacts:

1. If still using the original private `tailscale` providers, move their URLs and cache paths to `tailnet.domain.mrs` and `tailnet.ipcidr.mrs`, and rename the providers to `domain-tailnet` and `ipcidr-tailnet`.
2. Replace the `lan`, `cn`, and public `tailscale` providers with `domain-direct`, `ipcidr-direct`, and `classical-direct`, using the `direct` artifact URLs and cache paths below. The old groups' artifacts are no longer generated.
3. Route the `direct` providers through ordinary `DIRECT`. After the two `tailnet` rules, evaluate domain and classical providers in priority order before public IP-CIDR providers. Load `direct.classical.txt` with `behavior: classical` and `format: text`, and put `RULE-SET,classical-direct,DIRECT` before proxy domain and IP rules.
4. Use `rule-set:domain-tailnet` for the Quad100 DNS policy selector. Profiles using `rule-set:domain-tailscale` for private DNS must update it so public Tailscale services use the profile's ordinary DNS policy.
5. For every retained ruleset, update classical provider URLs and cache paths from `<name>.classical.list` to `<name>.classical.txt`, keeping `behavior: classical` and `format: text`.

The resulting Mihomo fragment is:

```yaml
proxies:
  - name: TAILSCALE
    type: direct
    interface-name: tailscale0

rule-providers:
  domain-tailnet:
    type: http
    behavior: domain
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/tailnet.domain.mrs
    path: ./ruleset/tailnet.domain.mrs
    interval: 86400
  ipcidr-tailnet:
    type: http
    behavior: ipcidr
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/tailnet.ipcidr.mrs
    path: ./ruleset/tailnet.ipcidr.mrs
    interval: 86400
  domain-direct:
    type: http
    behavior: domain
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.domain.mrs
    path: ./ruleset/direct.domain.mrs
    interval: 86400
  ipcidr-direct:
    type: http
    behavior: ipcidr
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.ipcidr.mrs
    path: ./ruleset/direct.ipcidr.mrs
    interval: 86400
  classical-direct:
    type: http
    behavior: classical
    format: text
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.classical.txt
    path: ./ruleset/direct.classical.txt
    interval: 86400

dns:
  nameserver-policy:
    "rule-set:domain-tailnet": "100.100.100.100"

rules:
  - RULE-SET,domain-tailnet,TAILSCALE
  - RULE-SET,ipcidr-tailnet,TAILSCALE,no-resolve
  - RULE-SET,domain-direct,DIRECT
  - RULE-SET,classical-direct,DIRECT
  # Insert us and global domain/classical rules here, before public IP rules.
  - RULE-SET,ipcidr-direct,DIRECT,no-resolve
  # Follow with remaining IP-CIDR and fallback rules.
```

`direct.classical.txt` includes the provider line `DST-PORT,7844`; the consumer's `RULE-SET` selects `DIRECT`. It covers Cloudflare Tunnel UDP QUIC and TCP HTTP/2 transport and matches all traffic to destination port 7844. The preserved upstream `PROCESS-NAME,cloudflared` rule also selects `DIRECT` when Mihomo can identify the process. Connections on port 443 follow the remaining domain and process rules. See Cloudflare's [Tunnel firewall requirements](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/configure-tunnels/tunnel-with-firewall/).

Keep `TAILSCALE` as the direct outbound bound to `tailscale0`. Public Tailscale control-plane and DERP hostnames use ordinary `DIRECT`; binding that public infrastructure traffic to `tailscale0` can prevent it from reaching the network that carries the Tailnet connection.

For **sub-store**, update its maintained template or profile-construction source with the provider URLs, cache paths, rule references, and DNS selector above, then regenerate and inspect the resulting profile. For **native Mihomo profiles**, make the equivalent edits directly in each maintained profile. Existing deployments should adopt these references after the new artifacts are published. This repository constructs and validates the artifacts; deployment and profile reloads are separate consumer steps.

## Domain and DERP boundaries

Tailscale's [DNS allowlist guidance](https://tailscale.com/docs/reference/troubleshooting/network-configuration/tailscale-domains-sites-unreachable) lists `*.tailscale.com` and `*.tailscale.io`. The maintained `+.` rules cover each apex and its subdomains, including public control-plane and DERP hostnames under those suffixes.

[MagicDNS](https://tailscale.com/docs/features/magicdns) assigns device names under `<tailnet>.ts.net`. [Funnel](https://tailscale.com/docs/features/tailscale-funnel#requirements-and-limitations) also exposes public services under that namespace, so `.ts.net` is not proof that every endpoint requires a private route. The construction preserves the existing generic `+.ts.net` grouping. Consumers that use public Funnel endpoints can narrow their private domain provider to their own tailnet and shared device domains, or add specific ordinary `DIRECT` exceptions before `domain-tailnet`. Keep the four custom names exact unless their subdomains also need a Tailnet route.

Domain rules apply when Mihomo has the hostname. These two public suffixes do not guarantee coverage of IP-only DERP traffic or infrastructure under other domains. The merged `direct` IP sources retain their existing local and China coverage; no public DERP IP source is added by this merge. Collecting public DERP addresses from DERPMap would be a separate extension.
