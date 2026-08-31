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
