# Migrate the Tailscale and Tailnet providers

The former `tailscale` group becomes `tailnet`. The new `tailscale` group contains Tailscale's official public service domains and uses ordinary `DIRECT`. This is a breaking change to provider meaning and artifact names.

| Group | Maintained contents | Generated artifacts | Consumer outbound |
| --- | --- | --- | --- |
| `tailnet` | `+.ts.net`; exact `forgejo.liblaf.me`, `jellyfin.liblaf.me`, `oplist.liblaf.me`, `qbt.liblaf.me`; `100.64.0.0/10`; `fd7a:115c:a1e0::/48` | `tailnet.domain.list`, `tailnet.domain.mrs`, `tailnet.ipcidr.list`, `tailnet.ipcidr.mrs` | `TAILSCALE`, a direct outbound bound to `tailscale0` |
| `tailscale` | `+.tailscale.com`, `+.tailscale.io` | `tailscale.domain.list`, `tailscale.domain.mrs` | Ordinary `DIRECT` |

Both groups precede `lan`, country, service, and `global` sets in the generated priority. Later sets explicitly exclude both groups. In particular, the public suffixes are removed from `global` by the existing exclusion mechanism. Consumer profiles must follow that priority because a broader domain or classical rule can still overlap part of an earlier group's match space.

The existing CN override `+.liblaf.me` is one such partial overlap: the four exact custom hosts select `tailnet` first, while the broader CN suffix remains. Generation asserts their routing priority and asserts exclusive membership for `.ts.net`, the private IP ranges, and the public Tailscale domains.

## Provider and DNS migration

Apply these changes together when consuming the new artifacts:

1. Move the private domain provider from `tailscale.domain.mrs` to `tailnet.domain.mrs`, rename it to `domain-tailnet`, and update its local cache path.
2. Move the private IP provider from `tailscale.ipcidr.mrs` to `tailnet.ipcidr.mrs`, rename it to `ipcidr-tailnet`, and update its local cache path. `tailscale.ipcidr.mrs` is no longer generated.
3. Use `domain-tailscale` for the new public `tailscale.domain.mrs` and route it through ordinary `DIRECT`.
4. Change the Quad100 DNS policy selector from `rule-set:domain-tailscale` to `rule-set:domain-tailnet` so the public service group uses the profile's ordinary DNS policy.

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
  domain-tailscale:
    type: http
    behavior: domain
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/tailscale.domain.mrs
    path: ./ruleset/tailscale.domain.mrs
    interval: 86400

dns:
  nameserver-policy:
    "rule-set:domain-tailnet": "100.100.100.100"

rules:
  - RULE-SET,domain-tailnet,TAILSCALE
  - RULE-SET,ipcidr-tailnet,TAILSCALE,no-resolve
  - RULE-SET,domain-tailscale,DIRECT
  # Follow with LAN, country, service, global, and fallback rules.
```

Keep `TAILSCALE` as the direct outbound bound to `tailscale0`. Public Tailscale control-plane and DERP hostnames use ordinary `DIRECT`; binding that public infrastructure traffic to `tailscale0` can prevent it from reaching the network that carries the Tailnet connection.

For **sub-store**, update its maintained template or profile-construction source with the provider URLs, cache paths, rule references, and DNS selector above, then regenerate and inspect the resulting profile. For **native Mihomo profiles**, make the equivalent edits directly in each maintained profile. Existing deployments should adopt these references after the new artifacts are published. This repository constructs and validates the artifacts; deployment and profile reloads are separate consumer steps.

## Domain and DERP boundaries

Tailscale's [DNS allowlist guidance](https://tailscale.com/docs/reference/troubleshooting/network-configuration/tailscale-domains-sites-unreachable) lists `*.tailscale.com` and `*.tailscale.io`. The maintained `+.` rules cover each apex and its subdomains, including public control-plane and DERP hostnames under those suffixes.

[MagicDNS](https://tailscale.com/docs/features/magicdns) assigns device names under `<tailnet>.ts.net`. [Funnel](https://tailscale.com/docs/features/tailscale-funnel#requirements-and-limitations) also exposes public services under that namespace, so `.ts.net` is not proof that every endpoint requires a private route. The construction preserves the existing generic `+.ts.net` grouping. Consumers that use public Funnel endpoints can narrow their private domain provider to their own tailnet and shared device domains, or add specific ordinary `DIRECT` exceptions before `domain-tailnet`. Keep the four custom names exact unless their subdomains also need a Tailnet route.

Domain rules apply when Mihomo has the hostname. IP-only DERP traffic is a separate concern: these two public suffixes do not guarantee coverage of connections identified only by IP address or infrastructure under other domains. There is no public DERP IP source in this split; collecting public DERP addresses from DERPMap would be a separate extension.
