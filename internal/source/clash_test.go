package source

import "testing"

func TestParseClashPreservesNonMRSRules(t *testing.T) {
	t.Parallel()
	payload := `# comment
DOMAIN,api.example.com
DOMAIN-SUFFIX,example.org
DOMAIN-WILDCARD,*.example.net
DOMAIN-WILDCARD,*.qhimgs?.com
DOMAIN-KEYWORD,openai
DOMAIN-REGEX,^api[0-9]+\.example\.com$
IP-CIDR,192.0.2.0/24,no-resolve
PROCESS-NAME,Example
`
	result, err := parseClash(payload)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(result.Rules.Domains), 3; got != want {
		t.Fatalf("got %d MRS domain rules, expected %d", got, want)
	}
	if got, want := len(result.Rules.Classical), 3; got != want {
		t.Fatalf("got %d classical rules, expected %d", got, want)
	}
	if got, want := len(result.Rules.Prefixes), 1; got != want {
		t.Fatalf("got %d IP-CIDR rules, expected %d", got, want)
	}
	if got, want := result.Ignored["PROCESS-NAME"], 1; got != want {
		t.Fatalf("ignored %d process rules, expected %d", got, want)
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
