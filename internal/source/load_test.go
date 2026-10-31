package source

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
