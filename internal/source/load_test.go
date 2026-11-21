package source

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/liblaf/route-rules/internal/config"
)

func TestLoaderLocalClashFilePreservesPortWithoutMRSEntries(t *testing.T) {
	t.Parallel()
	baseDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(baseDir, "tunnel.txt"), []byte("# Cloudflare Tunnel\nDST-PORT,7844\nDST-PORT,7844\n"), 0600); err != nil {
		t.Fatal(err)
	}
	loader := NewLoader("https://unused.invalid/v2fly")
	result, err := loader.Load(context.Background(), baseDir, config.Source{Name: "Cloudflare:tunnel", Kind: "clash-file", Path: "tunnel.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if result.InputRules != 2 || len(result.Rules.Classical) != 1 || result.Rules.Classical[0].Text() != "DST-PORT,7844" {
		t.Fatalf("local Clash ingest did not retain and deduplicate port rules: %+v", result)
	}
	if len(result.Rules.Domains) != 0 || len(result.Rules.Prefixes) != 0 || len(result.Fetches) != 0 || len(result.Ignored) != 0 {
		t.Fatalf("local port rules produced MRS entries, fetches or ignored rules: %+v", result)
	}
	if _, err := loader.Load(context.Background(), baseDir, config.Source{Name: "missing-tunnel", Kind: "clash-file", Path: "missing.txt"}); err == nil || !strings.Contains(err.Error(), "load missing-tunnel:") {
		t.Fatalf("got %v, want named source load error", err)
	}
}

func TestLoaderV2FlySelectionAndAttributes(t *testing.T) {
	t.Parallel()
	const fixture = `lists:
  - name: alpha
    length: 2
    rules:
      - domain:untagged.alpha.example
      - domain:cn.alpha.example:@cn
  - name: beta
    length: 3
    rules:
      - domain:untagged.beta.example
      - full:multi.beta.example:@cn,@telemetry
      - full:other.beta.example:@ads
`
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte(fixture))
	}))
	t.Cleanup(server.Close)

	loader := NewLoader(server.URL)

	named, err := loader.Load(context.Background(), "", config.Source{
		Name:  "v2fly:alpha",
		Kind:  "v2fly",
		Lists: []string{"alpha"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := domainRuleTexts(named), []string{"+.cn.alpha.example", "+.untagged.alpha.example"}; !slices.Equal(got, want) {
		t.Fatalf("named list rules: got %v, want %v", got, want)
	}

	allCN, err := loader.Load(context.Background(), "", config.Source{
		Name:              "v2fly:*@cn",
		Kind:              "v2fly",
		AllLists:          true,
		RequireAttributes: []string{"cn"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := domainRuleTexts(allCN), []string{"+.cn.alpha.example", "multi.beta.example"}; !slices.Equal(got, want) {
		t.Fatalf("all_lists @cn rules: got %v, want %v", got, want)
	}
	if allCN.InputRules != 2 {
		t.Fatalf("all_lists @cn input rules = %d, want 2", allCN.InputRules)
	}

	if _, err := loader.Load(context.Background(), "", config.Source{
		Name:  "v2fly:missing",
		Kind:  "v2fly",
		Lists: []string{"alpha", "missing"},
	}); err == nil || !strings.Contains(err.Error(), `V2Fly list "missing" not found`) {
		t.Fatalf("got %v, want missing list error", err)
	}

	if got := requests.Load(); got != 1 {
		t.Fatalf("V2Fly document fetched %d times, want 1", got)
	}
}

func domainRuleTexts(result Result) []string {
	texts := make([]string, 0, len(result.Rules.Domains))
	for _, rule := range result.Rules.Domains {
		texts = append(texts, rule.Text())
	}
	return texts
}
