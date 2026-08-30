package mihomo

import (
	"context"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/liblaf/route-rules/internal/rules"
)

func TestConverterRoundTripsMRS(t *testing.T) {
	path, err := exec.LookPath("mihomo")
	if err != nil {
		t.Skip("mihomo is not installed")
	}
	temporary := t.TempDir()
	converter := New(path)

	domainRule, err := rules.NewDomainRule(rules.DomainSuffix, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	domainSource := filepath.Join(temporary, "domain.txt")
	if err := os.WriteFile(domainSource, []byte("+.example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := converter.ConvertDomain(context.Background(), domainSource, filepath.Join(temporary, "domain.mrs"), rules.DomainSet{domainRule}); err != nil {
		t.Fatal(err)
	}

	prefix := netip.MustParsePrefix("192.0.2.0/24")
	ipSource := filepath.Join(temporary, "ipcidr.txt")
	if err := os.WriteFile(ipSource, []byte(prefix.String()+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := converter.ConvertIPCIDR(context.Background(), ipSource, filepath.Join(temporary, "ipcidr.mrs"), rules.PrefixSet{prefix}); err != nil {
		t.Fatal(err)
	}

	classicalSource := filepath.Join(temporary, "classical.list")
	if err := os.WriteFile(classicalSource, []byte("DOMAIN-REGEX,^api[0-9]+\\.example\\.com$\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := converter.ValidateClassical(context.Background(), classicalSource); err != nil {
		t.Fatal(err)
	}
	invalidClassicalSource := filepath.Join(temporary, "invalid-classical.list")
	if err := os.WriteFile(invalidClassicalSource, []byte("MYSTERY,example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = converter.ValidateClassical(context.Background(), invalidClassicalSource)
	if err == nil {
		t.Fatal("invalid classical ruleset passed Mihomo validation")
	}
	if !strings.Contains(err.Error(), "MYSTERY") {
		t.Fatalf("Mihomo did not report the invalid rule: %v", err)
	}
}
