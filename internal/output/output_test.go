package output

import (
	"context"
	"encoding/json"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/liblaf/route-rules/internal/build"
	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/mihomo"
	"github.com/liblaf/route-rules/internal/rules"
)

func TestWriteIndexUsesBrowserLocalAndRelativeGenerationTime(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "index.html")
	manifest := Manifest{
		GeneratedAt:   time.Date(2026, time.August, 31, 8, 9, 10, 0, time.FixedZone("UTC+8", 8*60*60)),
		MihomoVersion: "v-test",
	}
	if err := writeIndex(path, manifest); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(payload)
	for _, want := range []string{
		`<time id="generated-at" datetime="2026-08-31T00:09:10Z">2026-08-31 00:09:10 UTC</time>`,
		`new Intl.DateTimeFormat`,
		`new Intl.RelativeTimeFormat`,
		`setInterval(updateGeneratedTime, 60 * 1000)`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("generated index does not contain %q", want)
		}
	}
}

func TestWriteIndexWritesEmbeddedAssets(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := writeIndex(filepath.Join(directory, "index.html"), Manifest{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range webAssetNames {
		want, err := webFiles.ReadFile("web/" + name)
		if err != nil {
			t.Fatalf("read embedded asset %s: %v", name, err)
		}
		got, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Errorf("generated site does not contain %s: %v", name, err)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("generated %s differs from the embedded asset", name)
		}
	}
	directoryEntries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(directoryEntries))
	for _, entry := range directoryEntries {
		names = append(names, entry.Name())
	}
	if got, want := strings.Join(names, ","), "explorer.css,explorer.mjs,index.html,ruleset-data.mjs,ruleset-matcher.mjs"; got != want {
		t.Errorf("generated site files = %q, want %q", got, want)
	}
}

func TestWritePreservesPortRulesInClassicalCompanion(t *testing.T) {
	path, err := exec.LookPath("mihomo")
	if err != nil {
		t.Skip("mihomo is not installed")
	}
	domain, err := rules.NewDomainRule(rules.DomainSuffix, "example.com")
	if err != nil {
		t.Fatal(err)
	}
	port, err := rules.NewClassicalRule("DST-PORT", "7844")
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(t.TempDir(), "dist")
	manifest, err := Write(context.Background(), directory,
		config.Config{Repository: "example/rules", Priority: []string{"direct"}, Fallback: "PROXY"},
		build.Result{RuleSets: []build.RuleSet{{Name: "direct", Rules: rules.Collection{
			Domains:   rules.DomainSet{domain},
			Prefixes:  rules.PrefixSet{netip.MustParsePrefix("192.0.2.0/24")},
			Classical: rules.ClassicalSet{port},
		}}}}, mihomo.New(path), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"direct.classical.txt": "DST-PORT,7844\n",
		"direct.domain.list":   "+.example.com\n",
		"direct.ipcidr.list":   "192.0.2.0/24\n",
	} {
		payload, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(payload) != want {
			t.Errorf("%s = %q, want %q", name, payload, want)
		}
	}
	if _, err := os.Stat(filepath.Join(directory, "direct.classical.list")); !os.IsNotExist(err) {
		t.Errorf("obsolete classical filename exists or cannot be inspected: %v", err)
	}
	var companion *Artifact
	for i := range manifest.Artifacts {
		if manifest.Artifacts[i].Behavior == "classical" {
			companion = &manifest.Artifacts[i]
		}
	}
	if companion == nil {
		t.Fatal("manifest omits classical companion")
	}
	if companion.Path != "direct.classical.txt" || companion.Format != "text" || companion.RuleCount != 1 {
		t.Errorf("classical artifact = %+v", companion)
	}
	if companion.GitHubURL != "https://raw.githubusercontent.com/example/rules/mihomo/direct.classical.txt" || companion.JSDelivr != "https://cdn.jsdelivr.net/gh/example/rules@mihomo/direct.classical.txt" {
		t.Errorf("incorrect classical download links: %+v", companion)
	}
	payload, err := os.ReadFile(filepath.Join(directory, "stats.json"))
	if err != nil {
		t.Fatal(err)
	}
	var published Manifest
	if err := json.Unmarshal(payload, &published); err != nil {
		t.Fatal(err)
	}
	if len(published.Artifacts) != 5 {
		t.Errorf("published artifact count = %d, want 5", len(published.Artifacts))
	}
	for _, name := range []string{"stats.json", "index.html", "README.md"} {
		payload, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(payload), "direct.classical.txt") || strings.Contains(string(payload), "direct.classical.list") {
			t.Errorf("%s does not publish the classical .txt companion", name)
		}
	}
}
