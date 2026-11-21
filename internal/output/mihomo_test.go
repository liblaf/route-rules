package output

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

func TestMihomoRoundTripsMRSAndValidatesClassical(t *testing.T) {
	path, err := exec.LookPath("mihomo")
	if err != nil {
		t.Skip("mihomo is not installed")
	}
	dir := t.TempDir()
	cli := mihomoCLI(path)

	domainRule, err := rules.SuffixDomain("example.com")
	if err != nil {
		t.Fatal(err)
	}
	domainSource := filepath.Join(dir, "domain.txt")
	if err := os.WriteFile(domainSource, []byte("+.example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.convertDomain(context.Background(), domainSource, filepath.Join(dir, "domain.mrs"), rules.DomainSet{domainRule}); err != nil {
		t.Fatal(err)
	}

	prefix := netip.MustParsePrefix("192.0.2.0/24")
	ipSource := filepath.Join(dir, "ipcidr.txt")
	if err := os.WriteFile(ipSource, []byte(prefix.String()+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.convertIPCIDR(context.Background(), ipSource, filepath.Join(dir, "ipcidr.mrs"), rules.PrefixSet{prefix}); err != nil {
		t.Fatal(err)
	}

	classicalSource := filepath.Join(dir, "classical.txt")
	if err := os.WriteFile(classicalSource, []byte("DOMAIN-REGEX,^api[0-9]+\\.example\\.com$\nDST-PORT,7844\nPROCESS-NAME,cloudflared\nAND,((DST-PORT,7844),(NETWORK,UDP))\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.validateClassical(context.Background(), classicalSource); err != nil {
		t.Fatal(err)
	}
	invalidClassicalSource := filepath.Join(dir, "invalid-classical.txt")
	if err := os.WriteFile(invalidClassicalSource, []byte("DST-PORT,7844\nMYSTERY,example.com\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = cli.validateClassical(context.Background(), invalidClassicalSource)
	if err == nil {
		t.Fatal("invalid classical ruleset passed Mihomo validation")
	}
	if !strings.Contains(err.Error(), "MYSTERY") {
		t.Fatalf("Mihomo did not report the invalid rule: %v", err)
	}
}
