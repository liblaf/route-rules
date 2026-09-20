package output

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/liblaf/route-rules/internal/build"
	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/mihomo"
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

func Write(ctx context.Context, outputDir string, cfg config.Config, built build.Result, converter mihomo.Converter, generatedAt time.Time) (Manifest, error) {
	cleanOutput := filepath.Clean(outputDir)
	if cleanOutput == "." || cleanOutput == string(filepath.Separator) {
		return Manifest{}, fmt.Errorf("output directory must not be the current directory or a filesystem root")
	}
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return Manifest{}, fmt.Errorf("resolve output directory: %w", err)
	}
	base := filepath.Base(absOutput)
	if base == "." || base == string(filepath.Separator) {
		return Manifest{}, fmt.Errorf("output directory must not be a filesystem root")
	}
	if _, err := os.Stat(filepath.Join(absOutput, ".git")); err == nil {
		return Manifest{}, fmt.Errorf("output directory %s is a Git working tree", absOutput)
	} else if !os.IsNotExist(err) {
		return Manifest{}, fmt.Errorf("inspect output directory: %w", err)
	}
	parent := filepath.Dir(absOutput)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return Manifest{}, fmt.Errorf("create output parent: %w", err)
	}
	staging, err := os.MkdirTemp(parent, "."+base+"-build-")
	if err != nil {
		return Manifest{}, fmt.Errorf("create staging directory: %w", err)
	}
	keepStaging := false
	defer func() {
		if !keepStaging {
			_ = os.RemoveAll(staging)
		}
	}()

	version, err := converter.Version(ctx)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		GeneratedAt:   generatedAt.UTC(),
		Repository:    cfg.Repository,
		Priority:      append([]string(nil), cfg.Priority...),
		Fallback:      cfg.Fallback,
		MihomoVersion: version,
		RuleSets:      built.RuleSets,
	}

	for _, ruleset := range built.RuleSets {
		if len(ruleset.Rules.Domains) != 0 {
			textName := ruleset.Name + ".domain.list"
			textPath := filepath.Join(staging, textName)
			lines := make([]string, 0, len(ruleset.Rules.Domains))
			for _, rule := range ruleset.Rules.Domains {
				lines = append(lines, rule.Text())
			}
			if err := writeLines(textPath, lines); err != nil {
				return Manifest{}, err
			}
			manifest.Artifacts = append(manifest.Artifacts, artifactFor(staging, cfg.Repository, textName, ruleset.Name, "domain", "text", len(lines)))

			mrsName := ruleset.Name + ".domain.mrs"
			mrsPath := filepath.Join(staging, mrsName)
			if err := converter.ConvertDomain(ctx, textPath, mrsPath, ruleset.Rules.Domains); err != nil {
				return Manifest{}, fmt.Errorf("compile %s: %w", mrsName, err)
			}
			manifest.Artifacts = append(manifest.Artifacts, artifactFor(staging, cfg.Repository, mrsName, ruleset.Name, "domain", "mrs", len(lines)))
		}

		if len(ruleset.Rules.Prefixes) != 0 {
			textName := ruleset.Name + ".ipcidr.list"
			textPath := filepath.Join(staging, textName)
			lines := make([]string, 0, len(ruleset.Rules.Prefixes))
			for _, prefix := range ruleset.Rules.Prefixes {
				lines = append(lines, prefix.String())
			}
			if err := writeLines(textPath, lines); err != nil {
				return Manifest{}, err
			}
			manifest.Artifacts = append(manifest.Artifacts, artifactFor(staging, cfg.Repository, textName, ruleset.Name, "ipcidr", "text", len(lines)))

			mrsName := ruleset.Name + ".ipcidr.mrs"
			mrsPath := filepath.Join(staging, mrsName)
			if err := converter.ConvertIPCIDR(ctx, textPath, mrsPath, ruleset.Rules.Prefixes); err != nil {
				return Manifest{}, fmt.Errorf("compile %s: %w", mrsName, err)
			}
			manifest.Artifacts = append(manifest.Artifacts, artifactFor(staging, cfg.Repository, mrsName, ruleset.Name, "ipcidr", "mrs", len(lines)))
		}

		if len(ruleset.Rules.Classical) != 0 {
			name := ruleset.Name + ".classical.list"
			path := filepath.Join(staging, name)
			lines := make([]string, 0, len(ruleset.Rules.Classical))
			for _, rule := range ruleset.Rules.Classical {
				lines = append(lines, rule.Text())
			}
			if err := writeLines(path, lines); err != nil {
				return Manifest{}, err
			}
			if err := converter.ValidateClassical(ctx, path); err != nil {
				return Manifest{}, fmt.Errorf("validate %s: %w", name, err)
			}
			manifest.Artifacts = append(manifest.Artifacts, artifactFor(staging, cfg.Repository, name, ruleset.Name, "classical", "text", len(lines)))
		}
	}
	sort.Slice(manifest.Artifacts, func(i, j int) bool { return manifest.Artifacts[i].Path < manifest.Artifacts[j].Path })
	if err := writeJSON(filepath.Join(staging, "stats.json"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := writeIndex(filepath.Join(staging, "index.html"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := writeGeneratedREADME(filepath.Join(staging, "README.md"), manifest); err != nil {
		return Manifest{}, err
	}
	if err := os.WriteFile(filepath.Join(staging, ".nojekyll"), nil, 0o644); err != nil {
		return Manifest{}, fmt.Errorf("write .nojekyll: %w", err)
	}
	if err := replaceDirectory(absOutput, staging); err != nil {
		return Manifest{}, err
	}
	keepStaging = true
	return manifest, nil
}

func artifactFor(directory, repository, name, ruleset, behavior, format string, count int) Artifact {
	payload, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		panic(fmt.Sprintf("read just-written artifact %s: %v", name, err))
	}
	return Artifact{
		Path:      name,
		RuleSet:   ruleset,
		Behavior:  behavior,
		Format:    format,
		RuleCount: count,
		Bytes:     int64(len(payload)),
		GitHubURL: "https://raw.githubusercontent.com/" + repository + "/mihomo/" + name,
		JSDelivr:  "https://cdn.jsdelivr.net/gh/" + repository + "@mihomo/" + name,
	}
}

func writeLines(path string, lines []string) error {
	if len(lines) == 0 {
		return fmt.Errorf("refuse to write empty ruleset %s", path)
	}
	payload := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeJSON(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	return nil
}

func replaceDirectory(target, staging string) error {
	if _, err := os.Stat(target); os.IsNotExist(err) {
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

func writeIndex(path string, manifest Manifest) error {
	tmpl := template.Must(template.New("index").Funcs(template.FuncMap{
		"size": func(value any) string {
			switch typed := value.(type) {
			case int:
				return humanSize(int64(typed))
			case int64:
				return humanSize(typed)
			default:
				panic(fmt.Sprintf("unsupported byte count type %T", value))
			}
		},
		"rfc3339": func(value time.Time) string { return value.UTC().Format(time.RFC3339) },
		"time":    func(value time.Time) string { return value.UTC().Format("2006-01-02 15:04:05 UTC") },
		"subTotal": func(before, after build.Counts) int {
			return before.Domain + before.Classical + before.IPCIDR - after.Domain - after.Classical - after.IPCIDR
		},
	}).ParseFS(webFiles, "web/index.html"))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create site index: %w", err)
	}
	if err := tmpl.ExecuteTemplate(file, "index.html", manifest); err != nil {
		_ = file.Close()
		return fmt.Errorf("render site index: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close site index: %w", err)
	}
	for _, name := range webAssetNames {
		payload, err := webFiles.ReadFile("web/" + name)
		if err != nil {
			return fmt.Errorf("read embedded site asset %s: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(path), name), payload, 0o644); err != nil {
			return fmt.Errorf("write site asset %s: %w", name, err)
		}
	}
	return nil
}

//go:embed web/index.html web/ruleset-matcher.mjs web/ruleset-data.mjs web/explorer.mjs web/explorer.css
var webFiles embed.FS

var webAssetNames = []string{
	"ruleset-matcher.mjs",
	"ruleset-data.mjs",
	"explorer.mjs",
	"explorer.css",
}

func writeGeneratedREADME(path string, manifest Manifest) error {
	var text strings.Builder
	fmt.Fprintf(&text, "# Mihomo rule sets\n\nGenerated at %s. Do not edit this branch manually.\n\n", manifest.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&text, "Evaluation order: `%s`, then fallback `%s`.\n\n", strings.Join(manifest.Priority, "` → `"), manifest.Fallback)
	text.WriteString("| Artifact | Rules | Size | GitHub | jsDelivr |\n|---|---:|---:|---|---|\n")
	for _, artifact := range manifest.Artifacts {
		if artifact.Format != "mrs" {
			continue
		}
		fmt.Fprintf(&text, "| `%s` | %d | %s | [download](%s) | [download](%s) |\n", artifact.Path, artifact.RuleCount, humanSize(artifact.Bytes), artifact.GitHubURL, artifact.JSDelivr)
	}
	text.WriteString("\nFull statistics: [GitHub Pages](https://liblaf.github.io/route-rules/).\n")
	if err := os.WriteFile(path, []byte(text.String()), 0o644); err != nil {
		return fmt.Errorf("write generated README: %w", err)
	}
	return nil
}

func humanSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KiB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MiB", float64(bytes)/(1024*1024))
}
