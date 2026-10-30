# route-rules

`route-rules` builds compact Mihomo rule providers from maintained upstream domain and IP-CIDR lists. The generated files are published daily to the [`mihomo` branch](https://github.com/liblaf/route-rules/tree/mihomo); current counts, sizes, update time, and download links are shown on [GitHub Pages](https://liblaf.github.io/route-rules/).

## Policy and evaluation order

Mihomo must evaluate these sets in this order:

1. `tailnet` → the outbound that reaches resources through the Tailnet
2. `tailscale` → ordinary `DIRECT` for Tailscale's official public services
3. `lan` → `DIRECT`
4. `cn` → `DIRECT`
5. `us` → a United States proxy
6. `global` → the normal proxy pool
7. unmatched traffic → the normal proxy pool

OpenAI and Anthropic coverage is merged into `us`, including the maintained [Claude supporting-service domains](config/sources/anthropic.domain.txt) and [API inbound IP ranges](config/sources/anthropic.ipcidr.txt). Requests from other applications to those shared telemetry, support, and analytics domains also follow the `us` outbound. Consumers of the former `openai` or `anthropic` providers should use the `us` providers and their domain, IP-CIDR, and classical companions where present.

Evaluate both `tailnet` providers before `lan`: Tailscale's IPv4 range overlaps the shared-address space and its IPv6 range is inside the broader ULA space. Evaluate the public `tailscale` domain provider before country and `global` providers too. Keep domain and classical providers ahead of public IP-CIDR providers after these rules. This lets `us` take precedence over the geographic location of a CDN address.

The set exclusions in [`config/rulesets.json`](config/rulesets.json) are applied during generation. IP exclusions are exact CIDR set subtraction. A domain rule is removed only when an earlier rule covers its entire match space; partial domain differences cannot be represented by Mihomo and are resolved by the evaluation order above.

`tailnet` preserves the existing `+.ts.net` grouping and the four exact custom service domains. [MagicDNS](https://tailscale.com/docs/features/magicdns) uses names under `<tailnet>.ts.net`, and [Funnel](https://tailscale.com/docs/features/tailscale-funnel#requirements-and-limitations) also uses that namespace for public access. A `.ts.net` suffix alone does not prove that an endpoint requires a private route. Consumers can narrow this provider to their own tailnet and shared device domains, or place specific public Funnel exceptions ahead of it.

## Generated files

For each non-empty behavior, the `mihomo` branch contains:

- `<name>.domain.mrs` and its auditable `<name>.domain.list` input.
- `<name>.ipcidr.mrs` and its auditable `<name>.ipcidr.list` input.
- `<name>.classical.list` when an upstream contains keyword, regular-expression, or partial-wildcard rules that MRS cannot represent faithfully.

`tailnet` has both domain and IP-CIDR artifacts. `tailscale` currently has no IP source, so no empty IP-CIDR artifact is fabricated for it. Classical companions are emitted only for rules that need them; the maintained `tailnet` and `tailscale` sources need none. The build fails instead of silently dropping malformed input, unsupported source syntax, missing selected V2Fly lists, assertion failures, or MRS round-trip differences.

The Tailnet/public split produces `tailnet.domain.{list,mrs}`, `tailnet.ipcidr.{list,mrs}`, and `tailscale.domain.{list,mrs}`. See the [consumer migration guide](docs/tailscale-tailnet-migration.md) for provider, outbound, and DNS changes in sub-store and native Mihomo profiles.

The GitHub Pages site lets you browse each generated ruleset and test a literal domain or IP address against them. It uses the generated lists in the browser and does not perform DNS lookups. Results are displayed in the configured evaluation priority; unsupported regular-expression classical rules are listed as unevaluated.

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

- `tailnet`: the existing `+.ts.net` [MagicDNS grouping](https://tailscale.com/docs/features/magicdns), exact custom service names in the [domain source](config/sources/tailnet.domain.txt), and Tailscale's official [device address ranges](https://tailscale.com/docs/reference/reserved-ip-addresses) in the [CIDR source](config/sources/tailnet.ipcidr.txt).
- `tailscale`: `+.tailscale.com` and `+.tailscale.io` in the [public-service domain source](config/sources/tailscale.domain.txt), based on Tailscale's official [DNS allowlist guidance](https://tailscale.com/docs/reference/troubleshooting/network-configuration/tailscale-domains-sites-unreachable). Both suffixes are excluded from later sets, including `global`. This domain-only group covers public service and DERP hostnames under those suffixes; IP-only DERP traffic needs a separate consumer policy.
- `lan`: SukkaW LAN, V2Fly `private`, and selected IANA local-address prefixes.
- `cn`: dnsmasq-china-list, china-operator-ip, selected SukkaW direct/domestic lists, selected V2Fly lists, and local overrides.
- `us`: selected V2Fly AI/Apple lists, including OpenAI, Anthropic, and category-ai lists; SukkaW AI/Apple Services domain and IP rules; and the maintained Claude supporting-service domain and API inbound IP sources.
- `global`: V2Fly `geolocation-!cn` and SukkaW Telegram/global domain and IP rules.

Add reviewed CN exceptions to [`config/overrides/cn.domain.txt`](config/overrides/cn.domain.txt). An override is omitted automatically when another CN source already covers it.

## Build locally

Requirements: Go 1.25 or newer, Node.js 22 or newer for browser-module tests, and Mihomo 1.19.30 or newer on `PATH`.

```console
go test ./...
node --test internal/output/web/*.test.mjs
go run ./cmd/route-rules
```

The generator writes `dist/` atomically. It validates the requested domain decisions, compiles every non-empty MRS with Mihomo, decodes it again, and compares the decoded rules with the optimized text input. Mihomo also loads every classical companion in an isolated validation configuration.

The [GitHub Actions workflow](.github/workflows/build.yml) runs on changes to `main`, pull requests, manual dispatch, and daily at 16:23 UTC. Non-PR builds replace the `mihomo` branch contents and deploy the same output to GitHub Pages.
