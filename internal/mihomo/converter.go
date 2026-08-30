package mihomo

import (
	"bufio"
	"context"
	"fmt"
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
// validates provider declarations without loading their payloads. Provider
// parse failures are logged during startup; a clean timeout is the expected
// result after Mihomo has remained running with the provider loaded.
func (converter Converter) ValidateClassical(ctx context.Context, sourcePath string) error {
	payload, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read classical ruleset: %w", err)
	}
	if len(payload) == 0 {
		return fmt.Errorf("refuse to validate an empty classical ruleset")
	}

	directory, err := os.MkdirTemp(filepath.Dir(sourcePath), ".classical-check-")
	if err != nil {
		return fmt.Errorf("create classical validation directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(directory) }()
	rulesPath := filepath.Join(directory, "rules.list")
	if err := os.WriteFile(rulesPath, payload, 0o644); err != nil {
		return fmt.Errorf("write classical validation rules: %w", err)
	}
	configPath := filepath.Join(directory, "config.yaml")
	config := "rule-providers:\n" +
		"  validation:\n" +
		"    type: file\n" +
		"    behavior: classical\n" +
		"    format: text\n" +
		"    path: ./rules.list\n" +
		"rules:\n" +
		"  - RULE-SET,validation,DIRECT\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		return fmt.Errorf("write classical validation config: %w", err)
	}
	validationContext, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
	defer cancel()
	command := exec.CommandContext(validationContext, converter.path, "-d", directory, "-f", configPath)
	output, err := command.CombinedOutput()
	if validationContext.Err() == context.DeadlineExceeded {
		if strings.Contains(strings.ToLower(string(output)), "error") {
			return fmt.Errorf("validate classical ruleset with %s: %s", converter.path, strings.TrimSpace(string(output)))
		}
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("validate classical ruleset with %s: %w: %s", converter.path, err, strings.TrimSpace(string(output)))
	}
	return fmt.Errorf("classical validation process exited unexpectedly: %s", strings.TrimSpace(string(output)))
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
