package mihomo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/liblaf/route-rules/internal/rules"
)

type Converter struct {
	path string
}

func New(path string) Converter {
	return Converter{path: path}
}

func (converter Converter) Version(ctx context.Context) (string, error) {
	output, err := exec.CommandContext(ctx, converter.path, "-v").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("run %s -v: %w: %s", converter.path, err, strings.TrimSpace(string(output)))
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("%s -v returned no version", converter.path)
	}
	return strings.SplitN(version, "\n", 2)[0], nil
}

func (converter Converter) ConvertDomain(ctx context.Context, sourcePath, targetPath string, expected rules.DomainSet) error {
	if len(expected) == 0 {
		return fmt.Errorf("refuse to compile an empty domain ruleset")
	}
	if err := converter.convert(ctx, "domain", "text", sourcePath, targetPath); err != nil {
		return err
	}
	dumpPath, cleanup, err := temporaryPath(filepath.Dir(targetPath), ".domain-dump-*.txt")
	if err != nil {
		return err
	}
	defer cleanup()
	if err := converter.convert(ctx, "domain", "mrs", targetPath, dumpPath); err != nil {
		return fmt.Errorf("round-trip domain MRS: %w", err)
	}
	dumped, err := readDomainSet(dumpPath)
	if err != nil {
		return fmt.Errorf("read round-trip domain MRS: %w", err)
	}
	if diff := compareStrings(domainTexts(expected.Optimized()), domainTexts(dumped.Optimized())); diff != "" {
		return fmt.Errorf("domain MRS round-trip mismatch: %s", diff)
	}
	return nil
}

func (converter Converter) ConvertIPCIDR(ctx context.Context, sourcePath, targetPath string, expected rules.PrefixSet) error {
	if len(expected) == 0 {
		return fmt.Errorf("refuse to compile an empty IP-CIDR ruleset")
	}
	if err := converter.convert(ctx, "ipcidr", "text", sourcePath, targetPath); err != nil {
		return err
	}
	dumpPath, cleanup, err := temporaryPath(filepath.Dir(targetPath), ".ipcidr-dump-*.txt")
	if err != nil {
		return err
	}
	defer cleanup()
	if err := converter.convert(ctx, "ipcidr", "mrs", targetPath, dumpPath); err != nil {
		return fmt.Errorf("round-trip IP-CIDR MRS: %w", err)
	}
	dumped, err := readPrefixSet(dumpPath)
	if err != nil {
		return fmt.Errorf("read round-trip IP-CIDR MRS: %w", err)
	}
	if diff := compareStrings(prefixTexts(expected.Optimized()), prefixTexts(dumped.Optimized())); diff != "" {
		return fmt.Errorf("IP-CIDR MRS round-trip mismatch: %s", diff)
	}
	return nil
}

// ValidateClassical starts an isolated Mihomo instance because `mihomo -t`
// validates provider declarations without loading their payloads. Validation
// succeeds only after the controller reports that every rule was loaded.
func (converter Converter) ValidateClassical(ctx context.Context, sourcePath string) error {
	payload, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read classical ruleset: %w", err)
	}
	expected := 0
	for _, line := range strings.Split(string(payload), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			expected++
		}
	}
	if expected == 0 {
		return fmt.Errorf("refuse to validate an empty classical ruleset")
	}

	// A short temporary path keeps the controller socket below Unix path limits.
	directory, err := os.MkdirTemp("", "rr-classical-")
	if err != nil {
		return fmt.Errorf("create classical validation directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(directory) }()
	rulesPath := filepath.Join(directory, "rules.txt")
	if err := os.WriteFile(rulesPath, payload, 0o644); err != nil {
		return fmt.Errorf("write classical validation rules: %w", err)
	}
	configPath := filepath.Join(directory, "config.yaml")
	socketPath := filepath.Join(directory, "controller.sock")
	config := fmt.Sprintf("external-controller-unix: %q\n", socketPath) +
		"rule-providers:\n" +
		"  validation:\n" +
		"    type: file\n" +
		"    behavior: classical\n" +
		"    format: text\n" +
		"    path: ./rules.txt\n" +
		"rules:\n" +
		"  - RULE-SET,validation,DIRECT\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write classical validation config: %w", err)
	}
	logPath := filepath.Join(directory, "mihomo.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create classical validation log: %w", err)
	}
	defer logFile.Close()
	validationContext, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	command := exec.CommandContext(validationContext, converter.path, "-d", directory, "-f", configPath)
	command.Stdout, command.Stderr = logFile, logFile
	if err := command.Start(); err != nil {
		return fmt.Errorf("start classical validation: %w", err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	stopped := false
	defer func() {
		if !stopped {
			cancel()
			<-done
		}
	}()
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: time.Second}
	failure := func(err error) error {
		log, _ := os.ReadFile(logPath)
		return fmt.Errorf("validate classical ruleset with %s: %w: %s", converter.path, err, strings.TrimSpace(string(log)))
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			stopped = true
			if err == nil {
				err = fmt.Errorf("validation process exited before its provider loaded")
			}
			return failure(err)
		case <-validationContext.Done():
			return failure(validationContext.Err())
		case <-ticker.C:
			request, err := http.NewRequestWithContext(validationContext, http.MethodGet, "http://mihomo/providers/rules", nil)
			if err != nil {
				return failure(err)
			}
			response, err := client.Do(request)
			if err != nil {
				continue
			}
			var status struct {
				Providers map[string]struct {
					RuleCount int       `json:"ruleCount"`
					UpdatedAt time.Time `json:"updatedAt"`
				} `json:"providers"`
			}
			err = json.NewDecoder(response.Body).Decode(&status)
			response.Body.Close()
			if response.StatusCode != http.StatusOK || err != nil {
				continue
			}
			provider, exists := status.Providers["validation"]
			if !exists {
				continue
			}
			if provider.RuleCount == expected {
				return nil
			}
			if !provider.UpdatedAt.IsZero() {
				return failure(fmt.Errorf("loaded %d of %d classical rules", provider.RuleCount, expected))
			}
		}
	}
}

func (converter Converter) convert(ctx context.Context, behavior, format, sourcePath, targetPath string) error {
	command := exec.CommandContext(ctx, converter.path, "convert-ruleset", behavior, format, sourcePath, targetPath)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s convert-ruleset %s %s: %w: %s", converter.path, behavior, format, err, strings.TrimSpace(string(output)))
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

func temporaryPath(directory, pattern string) (string, func(), error) {
	file, err := os.CreateTemp(directory, pattern)
	if err != nil {
		return "", nil, fmt.Errorf("create temporary file: %w", err)
	}
	path := file.Name()
	if err := file.Close(); err != nil {
		return "", nil, fmt.Errorf("close temporary file: %w", err)
	}
	return path, func() { _ = os.Remove(path) }, nil
}

func readDomainSet(path string) (rules.DomainSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result rules.DomainSet
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		rule, err := rules.ParseDomainLine(line)
		if err != nil {
			return nil, err
		}
		result = append(result, rule)
	}
	return result, scanner.Err()
}

func readPrefixSet(path string) (rules.PrefixSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var result rules.PrefixSet
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return nil, err
		}
		result = append(result, prefix.Masked())
	}
	return result, scanner.Err()
}

func domainTexts(set rules.DomainSet) []string {
	result := make([]string, 0, len(set))
	for _, rule := range set {
		result = append(result, rule.Text())
	}
	sort.Strings(result)
	return result
}

func prefixTexts(set rules.PrefixSet) []string {
	result := make([]string, 0, len(set))
	for _, prefix := range set {
		result = append(result, prefix.String())
	}
	sort.Strings(result)
	return result
}

func compareStrings(expected, actual []string) string {
	if len(expected) != len(actual) {
		return fmt.Sprintf("got %d rules, expected %d", len(actual), len(expected))
	}
	for index := range expected {
		if expected[index] != actual[index] {
			return fmt.Sprintf("first difference at %d: got %q, expected %q", index, actual[index], expected[index])
		}
	}
	return ""
}
