package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
