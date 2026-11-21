package output

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/liblaf/route-rules/internal/rules"
)

type mihomoCLI string

func (bin mihomoCLI) version(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, string(bin), "-v").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run %s -v: %w: %s", bin, err, strings.TrimSpace(string(out)))
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	if line == "" {
		return "", fmt.Errorf("%s -v returned no version", bin)
	}
	return line, nil
}

func (bin mihomoCLI) convertDomain(ctx context.Context, sourcePath, targetPath string, expected rules.DomainSet) error {
	return bin.convertMRS(ctx, "domain", "domain", sourcePath, targetPath, expected.Optimized(), func(lines []string) ([]string, error) {
		parsed := make(rules.DomainSet, 0, len(lines))
		for _, line := range lines {
			rule, err := rules.ParseDomain(line)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, rule)
		}
		return parsed.Optimized(), nil
	})
}

func (bin mihomoCLI) convertIPCIDR(ctx context.Context, sourcePath, targetPath string, expected rules.PrefixSet) error {
	return bin.convertMRS(ctx, "ipcidr", "IP-CIDR", sourcePath, targetPath, expected.Optimized().Texts(), func(lines []string) ([]string, error) {
		parsed := make(rules.PrefixSet, 0, len(lines))
		for _, line := range lines {
			prefix, err := netip.ParsePrefix(line)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, prefix.Masked())
		}
		return parsed.Optimized().Texts(), nil
	})
}

func (bin mihomoCLI) convertMRS(
	ctx context.Context,
	behavior, label, sourcePath, targetPath string,
	expected []string,
	normalize func([]string) ([]string, error),
) error {
	if len(expected) == 0 {
		return fmt.Errorf("refuse to compile an empty %s ruleset", label)
	}
	if err := bin.runConvert(ctx, behavior, "text", sourcePath, targetPath); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(targetPath), "."+behavior+"-dump-*.txt")
	if err != nil {
		return fmt.Errorf("create temporary file: %w", err)
	}
	dumpPath := tmp.Name()
	defer func() { _ = os.Remove(dumpPath) }()
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary file: %w", err)
	}
	if err := bin.runConvert(ctx, behavior, "mrs", targetPath, dumpPath); err != nil {
		return fmt.Errorf("round-trip %s MRS: %w", label, err)
	}
	payload, err := os.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("read round-trip %s MRS: %w", label, err)
	}
	lines := make([]string, 0, len(expected))
	for line := range strings.SplitSeq(string(payload), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	actual, err := normalize(lines)
	if err != nil {
		return fmt.Errorf("read round-trip %s MRS: %w", label, err)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("%s MRS round-trip mismatch: got %d rules, expected %d", label, len(actual), len(expected))
	}
	for i := range expected {
		if expected[i] != actual[i] {
			return fmt.Errorf("%s MRS round-trip mismatch: first difference at %d: got %q, expected %q", label, i, actual[i], expected[i])
		}
	}
	return nil
}

func (bin mihomoCLI) runConvert(ctx context.Context, behavior, format, sourcePath, targetPath string) error {
	out, err := exec.CommandContext(ctx, string(bin), "convert-ruleset", behavior, format, sourcePath, targetPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s convert-ruleset %s %s: %w: %s", bin, behavior, format, err, strings.TrimSpace(string(out)))
	}
	info, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("stat converted ruleset: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("converted ruleset %s is empty", targetPath)
	}
	return nil
}

// validateClassical starts an isolated Mihomo instance because `mihomo -t`
// validates provider declarations without loading their payloads. Validation
// succeeds only after the controller reports that every rule was loaded.
func (bin mihomoCLI) validateClassical(ctx context.Context, sourcePath string) error {
	payload, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read classical ruleset: %w", err)
	}
	expected := 0
	for line := range strings.SplitSeq(string(payload), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			expected++
		}
	}
	if expected == 0 {
		return fmt.Errorf("refuse to validate an empty classical ruleset")
	}

	// A short temporary path keeps the controller socket below Unix path limits.
	dir, err := os.MkdirTemp("", "rr-classical-")
	if err != nil {
		return fmt.Errorf("create classical validation directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()
	if err := os.WriteFile(filepath.Join(dir, "rules.txt"), payload, 0o644); err != nil {
		return fmt.Errorf("write classical validation rules: %w", err)
	}
	configPath := filepath.Join(dir, "config.yaml")
	socketPath := filepath.Join(dir, "controller.sock")
	cfg := fmt.Sprintf("external-controller-unix: %q\nrule-providers:\n  validation:\n    type: file\n    behavior: classical\n    format: text\n    path: ./rules.txt\nrules:\n  - RULE-SET,validation,DIRECT\n", socketPath)
	if err := os.WriteFile(configPath, []byte(cfg), 0o644); err != nil {
		return fmt.Errorf("write classical validation config: %w", err)
	}
	logPath := filepath.Join(dir, "mihomo.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create classical validation log: %w", err)
	}
	defer logFile.Close()

	valCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(valCtx, string(bin), "-d", dir, "-f", configPath)
	cmd.Stdout, cmd.Stderr = logFile, logFile
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start classical validation: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	stopped := false
	stop := func() {
		if !stopped {
			stopped = true
			cancel()
			<-done
		}
	}
	defer stop()

	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	fail := func(err error) error {
		stop()
		log, _ := os.ReadFile(logPath)
		return fmt.Errorf("validate classical ruleset with %s: %w: %s", bin, err, strings.TrimSpace(string(log)))
	}

	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if finished, err := pollClassicalProvider(valCtx, client, expected); finished {
			if err != nil {
				return fail(err)
			}
			return nil
		}
		select {
		case waitErr := <-done:
			stopped = true
			if waitErr == nil {
				waitErr = fmt.Errorf("validation process exited before its provider loaded")
			}
			return fail(waitErr)
		case <-valCtx.Done():
			return fail(valCtx.Err())
		case <-ticker.C:
		}
	}
}

func pollClassicalProvider(ctx context.Context, client *http.Client, expected int) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://mihomo/providers/rules", nil)
	if err != nil {
		return true, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()
	var status struct {
		Providers map[string]struct {
			RuleCount int       `json:"ruleCount"`
			UpdatedAt time.Time `json:"updatedAt"`
		} `json:"providers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); resp.StatusCode != http.StatusOK || err != nil {
		return false, nil
	}
	provider, exists := status.Providers["validation"]
	if !exists {
		return false, nil
	}
	if provider.RuleCount == expected {
		return true, nil
	}
	if !provider.UpdatedAt.IsZero() {
		return true, fmt.Errorf("loaded %d of %d classical rules", provider.RuleCount, expected)
	}
	return false, nil
}
