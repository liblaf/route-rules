# route-rules

`route-rules` builds compact Mihomo rule providers from maintained upstream domain and IP-CIDR lists. The generated files are published daily to the [`mihomo` branch](https://github.com/liblaf/route-rules/tree/mihomo); current counts, sizes, update time, and download links are shown on [GitHub Pages](https://liblaf.github.io/route-rules/).

## Policy and evaluation order

Mihomo must evaluate these sets in this order:

1. `tailnet` → the outbound that reaches the Tailscale tailnet
2. `lan` → `DIRECT`
3. `cn` → `DIRECT`
4. `crypto` → a globally supported non-US region
5. `us` → a United States proxy
6. `proxy` → the normal proxy pool
7. unmatched traffic → the normal proxy pool

Evaluate both `tailnet` providers before `lan`: Tailscale's IPv4 range overlaps the shared-address space and its IPv6 range is inside the broader ULA space. Keep domain and classical providers ahead of public IP-CIDR providers after these private-destination rules. This lets an explicit service decision such as `us` or `crypto` take precedence over the geographic location of a CDN address.

The set exclusions in [`config/rulesets.json`](config/rulesets.json) are applied during generation. IP exclusions are exact CIDR set subtraction. A domain rule is removed only when an earlier rule covers its entire match space; partial domain differences cannot be represented by Mihomo and are resolved by the evaluation order above.

## Generated files

For each non-empty behavior, the `mihomo` branch contains:

- `<name>.domain.mrs` and its auditable `<name>.domain.list` input.
- `<name>.ipcidr.mrs` and its auditable `<name>.ipcidr.list` input.
- `<name>.classical.list` when an upstream contains keyword, regular-expression, or partial-wildcard rules that MRS cannot represent faithfully.

`tailnet` has both domain and IP-CIDR artifacts. `crypto` currently has no IP source, so no empty `crypto.ipcidr.mrs` is fabricated. The build fails instead of silently dropping malformed input, unsupported source syntax, missing selected V2Fly lists, assertion failures, or MRS round-trip differences.

Example provider:

```yaml
rule-providers:
  cn-domain:
    type: http
    behavior: domain
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/cn.domain.mrs
    path: ./ruleset/cn.domain.mrs
    interval: 86400
  cn-ipcidr:
    type: http
    behavior: ipcidr
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/cn.ipcidr.mrs
    path: ./ruleset/cn.ipcidr.mrs
    interval: 86400
  cn-classical:
    type: http
    behavior: classical
    format: text
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/cn.classical.list
    path: ./ruleset/cn.classical.list
    interval: 86400
```

Use all non-empty companions for full source fidelity. GitHub Raw and jsDelivr links for every MRS artifact are generated on the statistics page.

## Sources

The declarative source map is in [`config/rulesets.json`](config/rulesets.json). It combines:

- `tailnet`: Tailscale [MagicDNS names](https://tailscale.com/docs/features/magicdns) under `.ts.net` and its official [device address ranges](https://tailscale.com/docs/reference/reserved-ip-addresses). Public control-plane destinations such as `tailscale.com` and `tailscale.io` stay outside this set.
- `lan`: SukkaW LAN, V2Fly `private`, and selected IANA local-address prefixes.
- `cn`: dnsmasq-china-list, china-operator-ip, selected SukkaW direct/domestic lists, selected V2Fly lists plus every `@cn` entry, and local overrides.
- `crypto`: V2Fly `category-cryptocurrency`.
- `us`: selected V2Fly AI/Apple lists and SukkaW AI/Apple Services domain and IP rules.
- `proxy`: V2Fly `geolocation-!cn` and SukkaW Telegram/global domain and IP rules.

Add reviewed CN exceptions to [`config/overrides/cn.domain.txt`](config/overrides/cn.domain.txt). An override is omitted automatically when another CN source already covers it.

## Build locally

Requirements: Go 1.25 or newer and Mihomo 1.19.30 or newer on `PATH`.

```console
go test ./...
go run ./cmd/route-rules
```

The generator writes `dist/` atomically. It validates the requested domain decisions, compiles every non-empty MRS with Mihomo, decodes it again, and compares the decoded rules with the optimized text input. Mihomo also loads every classical companion in an isolated validation configuration.

The [GitHub Actions workflow](.github/workflows/build.yml) runs on changes to `main`, pull requests, manual dispatch, and daily at 16:23 UTC. Non-PR builds replace the `mihomo` branch contents and deploy the same output to GitHub Pages.
