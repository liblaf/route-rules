package output

import (
	"bytes"
	"cmp"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/liblaf/route-rules/internal/build"
)

type Artifact struct {
	Path      string `json:"path"`
	RuleSet   string `json:"ruleset"`
	Behavior  string `json:"behavior"`
	Format    string `json:"format"`
	RuleCount int    `json:"rule_count"`
	Bytes     int64  `json:"bytes"`
	GitHubURL string `json:"github_url"`
	JSDelivr  string `json:"jsdelivr_url"`
}

type Manifest struct {
	GeneratedAt   time.Time       `json:"generated_at"`
	Repository    string          `json:"repository"`
	Priority      []string        `json:"priority"`
	Fallback      string          `json:"fallback"`
	MihomoVersion string          `json:"mihomo_version"`
	RuleSets      []build.RuleSet `json:"rulesets"`
	Artifacts     []Artifact      `json:"artifacts"`
}

func Write(ctx context.Context, outputDir string, cfg build.Config, built build.Result, mihomoPath string, generatedAt time.Time) (Manifest, error) {
	if filepath.Clean(outputDir) == "." {
		return Manifest{}, fmt.Errorf("output directory must not be the current directory or a filesystem root")
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return Manifest{}, fmt.Errorf("resolve output directory: %w", err)
	}
	parent := filepath.Dir(absOutput)
	if parent == absOutput {
		return Manifest{}, fmt.Errorf("output directory must not be a filesystem root")
	}
	if _, err := os.Stat(filepath.Join(absOutput, ".git")); err == nil {
		return Manifest{}, fmt.Errorf("output directory %s is a Git working tree", absOutput)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Manifest{}, fmt.Errorf("inspect output directory: %w", err)
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Manifest{}, fmt.Errorf("create output parent: %w", err)
	}
	staging, err := os.MkdirTemp(parent, "."+filepath.Base(absOutput)+"-build-")
	if err != nil {
		return Manifest{}, fmt.Errorf("create staging directory: %w", err)
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(staging)
		}
	}()

	cli := mihomoCLI(mihomoPath)
	version, err := cli.version(ctx)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		GeneratedAt:   generatedAt.UTC(),
		Repository:    cfg.Repository,
		Priority:      slices.Clone(cfg.Priority),
		Fallback:      cfg.Fallback,
		MihomoVersion: version,
		RuleSets:      built.RuleSets,
	}

	record := func(name, ruleset, behavior, format string, count int) error {
		info, err := os.Stat(filepath.Join(staging, name))
		if err != nil {
			return fmt.Errorf("stat artifact %s: %w", name, err)
		}
		manifest.Artifacts = append(manifest.Artifacts, Artifact{
			Path:      name,
			RuleSet:   ruleset,
			Behavior:  behavior,
			Format:    format,
			RuleCount: count,
			Bytes:     info.Size(),
			GitHubURL: "https://raw.githubusercontent.com/" + cfg.Repository + "/mihomo/" + name,
			JSDelivr:  "https://cdn.jsdelivr.net/gh/" + cfg.Repository + "@mihomo/" + name,
		})
		return nil
	}
	writeMRS := func(rulesetName, behavior string, lines []string, compile func(textPath, mrsPath string) error) error {
		textName := rulesetName + "." + behavior + ".list"
		textPath := filepath.Join(staging, textName)
		if err := writeLines(textPath, lines); err != nil {
			return err
		}
		if err := record(textName, rulesetName, behavior, "text", len(lines)); err != nil {
			return err
		}
		mrsName := rulesetName + "." + behavior + ".mrs"
		if err := compile(textPath, filepath.Join(staging, mrsName)); err != nil {
			return fmt.Errorf("compile %s: %w", mrsName, err)
		}
		return record(mrsName, rulesetName, behavior, "mrs", len(lines))
	}

	for _, ruleset := range built.RuleSets {
		if len(ruleset.Rules.Domains) != 0 {
			if err := writeMRS(ruleset.Name, "domain", ruleset.Rules.Domains, func(textPath, mrsPath string) error {
				return cli.convertDomain(ctx, textPath, mrsPath, ruleset.Rules.Domains)
			}); err != nil {
				return Manifest{}, err
			}
		}
		if len(ruleset.Rules.Prefixes) != 0 {
			if err := writeMRS(ruleset.Name, "ipcidr", ruleset.Rules.Prefixes.Texts(), func(textPath, mrsPath string) error {
				return cli.convertIPCIDR(ctx, textPath, mrsPath, ruleset.Rules.Prefixes)
			}); err != nil {
				return Manifest{}, err
			}
		}
		if len(ruleset.Rules.Classical) != 0 {
			name := ruleset.Name + ".classical.txt"
			path := filepath.Join(staging, name)
			if err := writeLines(path, ruleset.Rules.Classical); err != nil {
				return Manifest{}, err
			}
			if err := cli.validateClassical(ctx, path); err != nil {
				return Manifest{}, fmt.Errorf("validate %s: %w", name, err)
			}
			if err := record(name, ruleset.Name, "classical", "text", len(ruleset.Rules.Classical)); err != nil {
				return Manifest{}, err
			}
		}
	}
	slices.SortFunc(manifest.Artifacts, func(a, b Artifact) int { return cmp.Compare(a.Path, b.Path) })
	if err := writeJSON(filepath.Join(staging, "stats.json"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := writeIndex(filepath.Join(staging, "index.html"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := writeREADME(filepath.Join(staging, "README.md"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := os.WriteFile(filepath.Join(staging, ".nojekyll"), nil, 0o644); err != nil {
		return Manifest{}, fmt.Errorf("write .nojekyll: %w", err)
	}
	if err := replaceDirectory(absOutput, staging); err != nil {
		return Manifest{}, err
	}
	published = true
	return manifest, nil
}

func isListedDownload(artifact Artifact) bool {
	return artifact.Format == "mrs" || artifact.Behavior == "classical"
}

func writeLines(path string, lines []string) error {
	if len(lines) == 0 {
		return fmt.Errorf("refuse to write empty ruleset %s", path)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func replaceDirectory(target, staging string) error {
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(staging, target); err != nil {
			return fmt.Errorf("publish output directory: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("inspect output directory: %w", err)
	}

	backup, err := os.MkdirTemp(filepath.Dir(target), "."+filepath.Base(target)+"-previous-")
	if err != nil {
		return fmt.Errorf("reserve output backup: %w", err)
	}
	if err := os.Remove(backup); err != nil {
		return fmt.Errorf("prepare output backup: %w", err)
	}
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backup previous output: %w", err)
	}
	if err := os.Rename(staging, target); err != nil {
		if restoreErr := os.Rename(backup, target); restoreErr != nil {
			return fmt.Errorf("publish output: %w (also failed to restore previous output: %v)", err, restoreErr)
		}
		return fmt.Errorf("publish output: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove previous output backup %s: %w", backup, err)
	}
	return nil
}

//go:embed web/index.html web/ruleset-matcher.mjs web/ruleset-data.mjs web/explorer.mjs web/explorer.css
var webFiles embed.FS

var indexTemplate = template.Must(template.New("index").Funcs(template.FuncMap{
	"size":             humanSize,
	"rfc3339":          func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
	"time":             func(t time.Time) string { return t.UTC().Format("2006-01-02 15:04:05 UTC") },
	"subTotal":         func(before, after build.Counts) int { return before.Total() - after.Total() },
	"isListedDownload": isListedDownload,
}).ParseFS(webFiles, "web/index.html"))

func writeIndex(path string, manifest Manifest) error {
	var buf bytes.Buffer
	if err := indexTemplate.ExecuteTemplate(&buf, "index.html", manifest); err != nil {
		return fmt.Errorf("render site index: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write site index: %w", err)
	}
	entries, err := webFiles.ReadDir("web")
	if err != nil {
		return fmt.Errorf("list embedded site assets: %w", err)
	}
	for _, entry := range entries {
		if entry.Name() == "index.html" {
			continue
		}
		payload, err := webFiles.ReadFile("web/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read embedded site asset %s: %w", entry.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), entry.Name()), payload, 0o644); err != nil {
			return fmt.Errorf("write site asset %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func writeREADME(path string, manifest Manifest) error {
	var text strings.Builder
	fmt.Fprintf(&text, "# Mihomo rule sets\n\nGenerated at %s. Do not edit this branch manually.\n\n", manifest.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&text, "Evaluation order: `%s`, then fallback `%s`.\n\n", strings.Join(manifest.Priority, "` → `"), manifest.Fallback)
	text.WriteString("| Artifact | Rules | Size | GitHub | jsDelivr |\n|---|---:|---:|---|---|\n")
	for _, a := range manifest.Artifacts {
		if isListedDownload(a) {
			fmt.Fprintf(&text, "| `%s` | %d | %s | [download](%s) | [download](%s) |\n", a.Path, a.RuleCount, humanSize(a.Bytes), a.GitHubURL, a.JSDelivr)
		}
	}
	text.WriteString("\nFull statistics: [GitHub Pages](https://liblaf.github.io/route-rules/).\n")
	if err := os.WriteFile(path, []byte(text.String()), 0o644); err != nil {
		return fmt.Errorf("write generated README: %w", err)
	}
	return nil
}

func humanSize(bytes int64) string {
	switch {
	case bytes < 1024:
		return fmt.Sprintf("%d B", bytes)
	case bytes < 1024*1024:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/(1024*1024))
	}
}
