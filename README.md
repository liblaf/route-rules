# route-rules

`route-rules` builds compact Mihomo rule providers from maintained upstream domain and IP-CIDR lists. The generated files are published daily to the [`mihomo` branch](https://github.com/liblaf/route-rules/tree/mihomo); current counts, sizes, update time, and download links are shown on [GitHub Pages](https://liblaf.github.io/route-rules/).

## Policy and evaluation order

Mihomo must evaluate these sets in this order:

1. `tailnet` → the outbound that reaches resources through the Tailnet
2. `direct` → ordinary `DIRECT` for LAN, China, Tailscale's official public services, and Cloudflare Tunnel transport
3. `us` → a United States proxy
4. `global` → the normal proxy pool
5. unmatched traffic → the normal proxy pool

OpenAI and Anthropic coverage is merged into `us`, including the maintained [Claude supporting-service domains](config/sources/anthropic.domain.txt) and [API inbound IP ranges](config/sources/anthropic.ipcidr.txt). Requests from other applications to those shared telemetry, support, and analytics domains also follow the `us` outbound. Consumers of the former `openai` or `anthropic` providers should use the `us` providers and their domain, IP-CIDR, and classical companions where present.

Evaluate both `tailnet` providers before `direct`: Tailscale's IPv4 range overlaps the shared-address space and its IPv6 range is inside the broader ULA space. After the `tailnet` rules, evaluate domain and classical providers in priority order before public IP-CIDR providers. This lets `us` take precedence over the geographic location of a CDN address.

The set exclusions in [`config/rulesets.json`](config/rulesets.json) are applied during generation. IP exclusions are exact CIDR set subtraction. A domain rule is removed only when an earlier rule covers its entire match space; partial domain differences cannot be represented by Mihomo and are resolved by the evaluation order above.

`tailnet` preserves the existing `+.ts.net` grouping and the four exact custom service domains. [MagicDNS](https://tailscale.com/docs/features/magicdns) uses names under `<tailnet>.ts.net`, and [Funnel](https://tailscale.com/docs/features/tailscale-funnel#requirements-and-limitations) also uses that namespace for public access. A `.ts.net` suffix alone does not prove that an endpoint requires a private route. Consumers can narrow this provider to their own tailnet and shared device domains, or place specific public Funnel exceptions ahead of it.

## Generated files

For each non-empty behavior, the `mihomo` branch contains:

- `<name>.domain.mrs` and its auditable `<name>.domain.list` input.
- `<name>.ipcidr.mrs` and its auditable `<name>.ipcidr.list` input.
- `<name>.classical.txt` for rules that domain and IP-CIDR providers cannot represent, including keyword, regular-expression, partial-wildcard, port, process, ASN, and logical rules.

`tailnet` and `direct` have both domain and IP-CIDR artifacts. Classical companions are emitted only for rules that need them; the maintained `tailnet` sources need none. The build fails instead of silently dropping malformed input, unsupported source syntax, missing selected V2Fly lists, assertion failures, or MRS round-trip differences.

The former `lan`, `cn`, and public `tailscale` providers are merged into `direct.domain.{list,mrs}`, `direct.ipcidr.{list,mrs}`, and any `direct.classical.txt` companion. See the [consumer migration guide](docs/tailscale-tailnet-migration.md) for provider, outbound, and DNS changes in sub-store and native Mihomo profiles.

The GitHub Pages site lets you browse each generated ruleset and test a literal domain or IP address against them. It uses the generated lists in the browser and does not perform DNS lookups. Results are displayed in the configured evaluation priority; classical rules that require request context or unsupported regular-expression syntax are listed as unevaluated.

Example providers:

```yaml
rule-providers:
  direct-domain:
    type: http
    behavior: domain
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.domain.mrs
    path: ./ruleset/direct.domain.mrs
    interval: 86400
  direct-ipcidr:
    type: http
    behavior: ipcidr
    format: mrs
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.ipcidr.mrs
    path: ./ruleset/direct.ipcidr.mrs
    interval: 86400
  direct-classical:
    type: http
    behavior: classical
    format: text
    url: https://raw.githubusercontent.com/liblaf/route-rules/mihomo/direct.classical.txt
    path: ./ruleset/direct.classical.txt
    interval: 86400
```

Use all non-empty companions for full source fidelity. Consumers of the former `<name>.classical.list` files must update their URLs and cache paths to `<name>.classical.txt`. After both Tailnet rules, add `RULE-SET,direct-classical,DIRECT` before proxy domain and IP rules. GitHub Raw and jsDelivr links for every artifact are generated on the statistics page.

`direct.classical.txt` includes `DST-PORT,7844`, with the outbound selected by the consumer's `RULE-SET` rule. This routes Cloudflare Tunnel's UDP QUIC and TCP HTTP/2 transport through ordinary `DIRECT`, including IP-only flows. It matches all traffic to destination port 7844. The preserved upstream `PROCESS-NAME,cloudflared` rule also selects `DIRECT` when Mihomo can identify the process. Connections on port 443 follow the remaining domain and process rules. See Cloudflare's [Tunnel firewall requirements](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/configure-tunnels/tunnel-with-firewall/).

## Sources

The declarative source map is in [`config/rulesets.json`](config/rulesets.json). It combines:

- `tailnet`: the existing `+.ts.net` [MagicDNS grouping](https://tailscale.com/docs/features/magicdns), exact custom service names in the [domain source](config/sources/tailnet.domain.txt), and Tailscale's official [device address ranges](https://tailscale.com/docs/reference/reserved-ip-addresses) in the [CIDR source](config/sources/tailnet.ipcidr.txt).
- `direct`: SukkaW LAN, V2Fly `private`, selected IANA local-address prefixes, dnsmasq-china-list, china-operator-ip, selected SukkaW direct/domestic lists, selected V2Fly lists, local overrides, [Cloudflare Tunnel transport](config/sources/cloudflared.classical.txt), and Tailscale's `+.tailscale.com` / `+.tailscale.io` [public-service domains](config/sources/tailscale.domain.txt). The Tailscale suffixes follow the official [DNS allowlist guidance](https://tailscale.com/docs/reference/troubleshooting/network-configuration/tailscale-domains-sites-unreachable) and are excluded from later sets. IP-only DERP traffic still needs a separate consumer policy.
- `us`: selected V2Fly AI/Apple lists, including OpenAI, Anthropic, and category-ai lists; SukkaW AI/Apple Services domain and IP rules; and the maintained Claude supporting-service domain and API inbound IP sources.
- `global`: V2Fly `geolocation-!cn` and SukkaW Telegram/global domain and IP rules.

Add reviewed direct-routing exceptions to [`config/overrides/cn.domain.txt`](config/overrides/cn.domain.txt). An override is omitted automatically when another `direct` source already covers it.

## Build locally

Requirements: Go 1.25 or newer, Node.js 22 or newer for browser-module tests, and Mihomo 1.19.30 or newer on `PATH`.

```console
go test ./...
node --test internal/output/web/*.test.mjs
go run ./cmd/route-rules
```

The generator writes `dist/` atomically. It validates domain, IP, and destination-port assertions, compiles every non-empty MRS with Mihomo, decodes it again, and compares the decoded rules with the optimized text input. Mihomo also loads every classical companion in an isolated validation configuration; the controller must confirm that every entry loaded successfully.

The [GitHub Actions workflow](.github/workflows/build.yml) runs on changes to `main`, pull requests, manual dispatch, and daily at 16:23 UTC. Non-PR builds replace the `mihomo` branch contents and deploy the same output to GitHub Pages.
