package build

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
)

func TestParseClashPreservesNonMRSRules(t *testing.T) {
	t.Parallel()
	payload := `# comment
DOMAIN,api.example.com
DOMAIN-SUFFIX,example.org
DOMAIN-WILDCARD,*.example.net
DOMAIN-WILDCARD,*.qhimgs?.com
DOMAIN-KEYWORD,openai
DOMAIN-REGEX,^api[0-9]{1,3}\.example\.com$
IP-CIDR,192.0.2.0/24,no-resolve
DST-PORT,7844
PROCESS-NAME,Example
`
	loaded, report, err := parseClash(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(loaded.Domains), 3; got != want {
		t.Fatalf("got %d MRS domain rules, expected %d", got, want)
	}
	if got, want := len(loaded.Classical), 5; got != want {
		t.Fatalf("got %d classical rules, expected %d", got, want)
	}
	if got, want := len(loaded.Prefixes), 1; got != want {
		t.Fatalf("got %d IP-CIDR rules, expected %d", got, want)
	}
	if report.Ignored["PROCESS-NAME"] != 0 || report.Ignored["DST-PORT"] != 0 {
		t.Fatalf("unexpected ignored counts: %+v", report.Ignored)
	}
	if loaded.Classical[3] != "DST-PORT,7844" {
		t.Fatalf("destination port rule was not preserved: %v", loaded.Classical)
	}
	if loaded.MatchesDomain("unrelated.example") {
		t.Fatal("destination port rule matched an unrelated domain")
	}
	if !loaded.MatchesDomain("api.openai.example") {
		t.Fatal("classical domain matching stopped working beside a port rule")
	}
	if !loaded.MatchesDomain("api42.example.com") || loaded.MatchesDomain("api1234.example.com") {
		t.Fatal("classical regex quantifier was not preserved")
	}
}

func TestParseClashRejectsInvalidDestinationPortAndUnknownKind(t *testing.T) {
	t.Parallel()
	for _, port := range []string{"", "https", "-1", "65536", "7844-", "7844-7845-7846"} {
		t.Run(port, func(t *testing.T) {
			_, _, err := parseClash("# comment\nDST-PORT," + port + "\n")
			if err == nil || !strings.Contains(err.Error(), "line 2:") {
				t.Fatalf("got %v, want invalid port error with source line", err)
			}
		})
	}
	if _, _, err := parseClash("MYSTERY,example.com\n"); err == nil {
		t.Fatal("unknown Clash rule kind did not fail")
	}
}

func TestParseClashPreservesContextualPayloadAndCountsIgnored(t *testing.T) {
	t.Parallel()
	lines := []string{
		"PROCESS-NAME,cloudflared",
		"PROCESS-PATH-REGEX,^/opt/app[0-9]{1,3}$",
		"IP-ASN,13335,no-resolve,src",
		"AND,((NETWORK,UDP),(DST-PORT,7844))",
		"IP-CIDR,192.0.2.0/24,no-resolve,src",
		"IP-CIDR,198.51.100.0/24,SRC",
		"IP-CIDR6,2001:db8::/32,src",
		"SRC-IP-CIDR,198.51.100.0/24",
		"SRC-PORT,1024-65535",
		"NETWORK,UDP",
		"DOMAIN-REGEX,^api[0-9]{1,3}\\.example\\.com$",
	}
	loaded, report, err := parseClash(strings.Join(lines, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Classical) != len(lines) || len(report.Ignored) != 0 || len(loaded.Prefixes) != 0 {
		t.Fatalf("contextual rules were lost or converted to destination prefixes: %+v %+v", loaded, report)
	}
	if !slices.Equal(loaded.Classical, lines) {
		t.Fatalf("got %v, want %v", loaded.Classical, lines)
	}
	if loaded.MatchesDomain("unrelated.example") {
		t.Fatal("contextual rules matched a domain without connection metadata")
	}
	for _, kind := range []string{"FINAL", "MATCH", "RULE-SET", "SUB-RULE", "URL-REGEX", "USER-AGENT"} {
		ignoredLoaded, ignoredReport, err := parseClash(kind + ",example")
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if ignoredReport.Ignored[kind] != 1 || len(ignoredLoaded.Classical) != 0 || len(ignoredLoaded.Domains) != 0 || len(ignoredLoaded.Prefixes) != 0 {
			t.Fatalf("%s was not explicitly counted as ignored: %+v %+v", kind, ignoredLoaded, ignoredReport)
		}
	}
}

func TestParseCIDR(t *testing.T) {
	t.Parallel()
	loaded, report, err := parseCIDR("# Tailscale ranges\n100.64.0.0/10\nfd7a:115c:a1e0::/48\n")
	if err != nil {
		t.Fatal(err)
	}
	if report.InputRules != 2 {
		t.Fatalf("got %d input rules, expected 2", report.InputRules)
	}
	for _, addr := range []netip.Addr{
		netip.MustParseAddr("100.64.0.1"),
		netip.MustParseAddr("fd7a:115c:a1e0::1"),
	} {
		if !loaded.Prefixes.Contains(addr) {
			t.Fatalf("parsed prefixes do not contain %s", addr)
		}
	}
}

func TestLoaderLocalClashFilePreservesPortWithoutMRSEntries(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(baseDir, "tunnel.txt"), []byte("# Cloudflare Tunnel\nDST-PORT,7844\nDST-PORT,7844\n"), 0600); err != nil {
		t.Fatal(err)
	}
	loader := newSourceLoader("https://unused.invalid/v2fly")
	loaded, report, err := loader.load(context.Background(), baseDir, SourceConfig{Name: "Cloudflare:tunnel", Kind: "clash-file", Path: "tunnel.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if report.InputRules != 2 || len(loaded.Classical) != 1 || loaded.Classical[0] != "DST-PORT,7844" {
		t.Fatalf("local Clash ingest did not retain and deduplicate port rules: %+v %+v", loaded, report)
	}
	if len(loaded.Domains) != 0 || len(loaded.Prefixes) != 0 || len(report.Fetches) != 0 || len(report.Ignored) != 0 {
		t.Fatalf("local port rules produced MRS entries, fetches or ignored rules: %+v %+v", loaded, report)
	}
	if _, _, err := loader.load(context.Background(), baseDir, SourceConfig{Name: "missing-tunnel", Kind: "clash-file", Path: "missing.txt"}); err == nil || !strings.Contains(err.Error(), "load missing-tunnel:") {
		t.Fatalf("got %v, want named source load error", err)
	}
}

func TestParseV2FlyRuleAttributesAndSelection(t *testing.T) {
	t.Parallel()
	kind, value, attrs, err := parseV2FlyRule("domain:example.com")
	if err != nil || kind != "domain" || value != "example.com" || attrs != "" || !hasV2FlyAttributes(attrs, nil) || hasV2FlyAttributes(attrs, []string{"cn"}) {
		t.Fatalf("untagged rule: got (%q, %q, %q, %v)", kind, value, attrs, err)
	}
	kind, value, attrs, err = parseV2FlyRule("full:firebase-settings.crashlytics.com:@cn,@telemetry")
	if err != nil || kind != "full" || value != "firebase-settings.crashlytics.com" ||
		!hasV2FlyAttributes(attrs, []string{"cn", "telemetry"}) || hasV2FlyAttributes(attrs, []string{"cn", "ads"}) {
		t.Fatalf("multi-tag rule: got (%q, %q, %q, %v)", kind, value, attrs, err)
	}
	for _, bad := range []string{
		"badrule", "domain:", ":example.cn", "domain::@cn",
		"domain:example.cn:@", "domain:example.cn:@cn,@", "domain:example.cn:@,@cn", "domain:example.cn:@cn,ads",
	} {
		if _, _, _, err := parseV2FlyRule(bad); err == nil {
			t.Errorf("malformed V2Fly rule %q did not fail", bad)
		}
	}

	const fixture = `lists:
  - name: alpha
    length: 2
    rules:
      - domain:untagged.alpha.example
      - domain:cn.alpha.example:@cn
  - name: beta
    length: 4
    rules:
      - domain:untagged.beta.example
      - full:multi.beta.example:@cn,@telemetry
      - keyword:tracker:@cn
      - regexp:^api[0-9]+\.example\.com$:@cn
`
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(server.Close)

	loader := newSourceLoader(server.URL)
	named, _, err := loader.load(context.Background(), "", SourceConfig{Name: "v2fly:alpha", Kind: "v2fly", Lists: []string{"alpha"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(named.Domains, []string{"+.cn.alpha.example", "+.untagged.alpha.example"}) {
		t.Fatalf("named list rules: got %v", named.Domains)
	}
	allCN, report, err := loader.load(context.Background(), "", SourceConfig{Name: "v2fly:*@cn", Kind: "v2fly", AllLists: true, RequireAttributes: []string{"cn"}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(allCN.Domains, []string{"+.cn.alpha.example", "multi.beta.example"}) || len(allCN.Classical) != 2 || report.InputRules != 4 {
		t.Fatalf("all_lists @cn: got %+v %+v", allCN, report)
	}
	if _, _, err := loader.load(context.Background(), "", SourceConfig{Name: "v2fly:missing", Kind: "v2fly", Lists: []string{"alpha", "missing"}}); err == nil || !strings.Contains(err.Error(), `V2Fly list "missing" not found`) {
		t.Fatalf("got %v, want missing list error", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("V2Fly document fetched %d times, want 1", requests.Load())
	}
}

func TestFetcherCanceledContextAndParseIANALocalCSV(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	loader := newSourceLoader("https://unused.invalid/v2fly")
	if _, _, err := loader.fetch(ctx, "https://unused.invalid/canceled"); err == nil || strings.Contains(err.Error(), "after 3 attempts") {
		t.Fatalf("canceled error = %v, want direct cancellation error", err)
	}

	csvPayload := "Address Block,Name\n10.0.0.0/8 [RFC1918],Private-Use\n203.0.113.0/24,Documentation\n"
	prefixes, err := parseIANALocalCSV([]byte(csvPayload))
	if err != nil {
		t.Fatal(err)
	}
	if len(prefixes) != 1 || prefixes[0].String() != "10.0.0.0/8" {
		t.Fatalf("parseIANALocalCSV = %v, want [10.0.0.0/8]", prefixes)
	}
}
