package source

import (
	"net/netip"
	"strings"
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
	result, err := parseClash(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(result.Rules.Domains), 3; got != want {
		t.Fatalf("got %d MRS domain rules, expected %d", got, want)
	}
	if got, want := len(result.Rules.Classical), 5; got != want {
		t.Fatalf("got %d classical rules, expected %d", got, want)
	}
	if got, want := len(result.Rules.Prefixes), 1; got != want {
		t.Fatalf("got %d IP-CIDR rules, expected %d", got, want)
	}
	if got, want := result.Ignored["PROCESS-NAME"], 0; got != want {
		t.Fatalf("ignored %d process rules, expected %d", got, want)
	}
	if result.Ignored["DST-PORT"] != 0 {
		t.Fatal("destination port rule was ignored")
	}
	if result.Rules.Classical[3].Text() != "DST-PORT,7844" {
		t.Fatalf("destination port rule was not preserved: %v", result.Rules.Classical)
	}
	if result.Rules.MatchesDomain("unrelated.example") {
		t.Fatal("destination port rule matched an unrelated domain")
	}
	if !result.Rules.MatchesDomain("api.openai.example") {
		t.Fatal("classical domain matching stopped working beside a port rule")
	}
	if !result.Rules.MatchesDomain("api42.example.com") || result.Rules.MatchesDomain("api1234.example.com") {
		t.Fatal("classical regex quantifier was not preserved")
	}
}

func TestParseClashRejectsInvalidDestinationPort(t *testing.T) {
	t.Parallel()
	for _, port := range []string{"", "https", "-1", "65536", "7844-", "7844-7845-7846"} {
		t.Run(port, func(t *testing.T) {
			_, err := parseClash("# comment\nDST-PORT," + port + "\n")
			if err == nil || !strings.Contains(err.Error(), "line 2:") {
				t.Fatalf("got %v, want invalid port error with source line", err)
			}
		})
	}
}

func TestParseClashRejectsUnknownKind(t *testing.T) {
	t.Parallel()
	if _, err := parseClash("MYSTERY,example.com\n"); err == nil {
		t.Fatal("unknown Clash rule kind did not fail")
	}
}

func TestParseV2FlyRuleAttributes(t *testing.T) {
	t.Parallel()
	kind, value, attributes, err := parseV2FlyRule("domain:example.cn:@cn")
	if err != nil {
		t.Fatal(err)
	}
	if kind != "domain" || value != "example.cn" {
		t.Fatalf("got kind=%q value=%q", kind, value)
	}
	if _, exists := attributes["cn"]; !exists {
		t.Fatal("cn attribute was not parsed")
	}
}

func TestParseCIDR(t *testing.T) {
	t.Parallel()
	result, err := parseCIDR("# Tailscale ranges\n100.64.0.0/10\nfd7a:115c:a1e0::/48\n")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := result.InputRules, 2; got != want {
		t.Fatalf("got %d input rules, expected %d", got, want)
	}
	for _, address := range []netip.Addr{
		netip.MustParseAddr("100.64.0.1"),
		netip.MustParseAddr("fd7a:115c:a1e0::1"),
	} {
		if !result.Rules.Prefixes.Contains(address) {
			t.Fatalf("parsed prefixes do not contain %s", address)
		}
	}
}

func TestParseClashPreservesContextualPayloadAndOptions(t *testing.T) {
	t.Parallel()
	lines := []string{
		"PROCESS-NAME,cloudflared",
		"PROCESS-PATH-REGEX,^/opt/app[0-9]{1,3}$",
		"IP-ASN,13335,no-resolve,src",
		"AND,((NETWORK,UDP),(DST-PORT,7844))",
		"IP-CIDR,192.0.2.0/24,no-resolve,src",
		"IP-CIDR6,2001:db8::/32,src",
		"SRC-IP-CIDR,198.51.100.0/24",
		"SRC-PORT,1024-65535",
		"NETWORK,UDP",
		"DOMAIN-REGEX,^api[0-9]{1,3}\\.example\\.com$",
	}
	result, err := parseClash(strings.Join(lines, "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rules.Classical) != len(lines) || len(result.Ignored) != 0 || len(result.Rules.Prefixes) != 0 {
		t.Fatalf("contextual rules were lost or converted to destination prefixes: %+v", result)
	}
	for i, line := range lines {
		if got := result.Rules.Classical[i].Text(); got != line {
			t.Errorf("rule %d = %q, want %q", i, got, line)
		}
	}
	if result.Rules.MatchesDomain("unrelated.example") {
		t.Fatal("contextual rules matched a domain without connection metadata")
	}
}

func TestParseClashCountsRulesUnsupportedByClassicalProviders(t *testing.T) {
	t.Parallel()
	for kind := range ignoredClashKinds {
		result, err := parseClash(kind + ",example")
		if err != nil {
			t.Fatalf("%s: %v", kind, err)
		}
		if result.Ignored[kind] != 1 || len(result.Rules.Classical) != 0 || len(result.Rules.Domains) != 0 || len(result.Rules.Prefixes) != 0 {
			t.Fatalf("%s was not explicitly counted as ignored: %+v", kind, result)
		}
	}
}
