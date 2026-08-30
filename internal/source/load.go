package source

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/rules"
	"gopkg.in/yaml.v3"
)

type Result struct {
	Rules      rules.Collection
	InputRules int
	Ignored    map[string]int
	Fetches    []FetchMetadata
}

type Loader struct {
	fetcher    *Fetcher
	v2flyURL   string
	v2fly      *v2flyDocument
	v2flyFetch FetchMetadata
}

func NewLoader(v2flyURL string) *Loader {
	return &Loader{fetcher: NewFetcher(), v2flyURL: v2flyURL}
}

func (loader *Loader) Load(ctx context.Context, baseDir string, source config.Source) (Result, error) {
	var result Result
	var err error
	switch source.Kind {
	case "clash":
		result, err = loader.loadClash(ctx, source.URL)
	case "dnsmasq":
		result, err = loader.loadDNSMasq(ctx, source.URL)
	case "cidr":
		result, err = loader.loadCIDR(ctx, source.URL)
	case "cidr-file":
		result, err = loadCIDRFile(baseDir, source.Path)
	case "v2fly":
		result, err = loader.loadV2Fly(ctx, source)
	case "iana-local":
		result, err = loader.loadIANALocal(ctx, source.URLs)
	case "domain-file":
		result, err = loadDomainFile(baseDir, source.Path)
	default:
		panic(fmt.Sprintf("validated source has unknown kind %q", source.Kind))
	}
	if err != nil {
		return Result{}, fmt.Errorf("load %s: %w", source.Name, err)
	}
	result.Rules = result.Rules.Optimized()
	return result, nil
}

func (loader *Loader) loadClash(ctx context.Context, url string) (Result, error) {
	payload, metadata, err := loader.fetcher.Fetch(ctx, url)
	if err != nil {
		return Result{}, err
	}
	result, err := parseClash(string(payload))
	result.Fetches = []FetchMetadata{metadata}
	return result, err
}

func (loader *Loader) loadDNSMasq(ctx context.Context, url string) (Result, error) {
	payload, metadata, err := loader.fetcher.Fetch(ctx, url)
	if err != nil {
		return Result{}, err
	}
	result := Result{Ignored: make(map[string]int), Fetches: []FetchMetadata{metadata}}
	for number, raw := range strings.Split(string(payload), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "server=/") {
			return Result{}, fmt.Errorf("line %d: unsupported dnsmasq directive %q", number+1, line)
		}
		remainder := strings.TrimPrefix(line, "server=/")
		end := strings.IndexByte(remainder, '/')
		if end <= 0 {
			return Result{}, fmt.Errorf("line %d: malformed dnsmasq server directive", number+1)
		}
		domain, err := rules.NewDomainRule(rules.DomainSuffix, remainder[:end])
		if err != nil {
			return Result{}, fmt.Errorf("line %d: %w", number+1, err)
		}
		result.Rules.Domains = append(result.Rules.Domains, domain)
		result.InputRules++
	}
	return result, nil
}

func (loader *Loader) loadCIDR(ctx context.Context, url string) (Result, error) {
	payload, metadata, err := loader.fetcher.Fetch(ctx, url)
	if err != nil {
		return Result{}, err
	}
	result, err := parseCIDR(string(payload))
	result.Fetches = []FetchMetadata{metadata}
	return result, err
}

func loadCIDRFile(baseDir, path string) (Result, error) {
	payload, err := os.ReadFile(filepath.Join(baseDir, filepath.Clean(path)))
	if err != nil {
		return Result{}, err
	}
	return parseCIDR(string(payload))
}

func parseCIDR(payload string) (Result, error) {
	result := Result{Ignored: make(map[string]int)}
	for number, raw := range strings.Split(payload, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return Result{}, fmt.Errorf("line %d: parse prefix %q: %w", number+1, line, err)
		}
		result.Rules.Prefixes = append(result.Rules.Prefixes, prefix.Masked())
		result.InputRules++
	}
	return result, nil
}

func loadDomainFile(baseDir, path string) (Result, error) {
	payload, err := os.ReadFile(filepath.Join(baseDir, filepath.Clean(path)))
	if err != nil {
		return Result{}, err
	}
	result := Result{Ignored: make(map[string]int)}
	for number, raw := range strings.Split(string(payload), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rule, err := rules.ParseDomainLine(line)
		if err != nil {
			return Result{}, fmt.Errorf("line %d: %w", number+1, err)
		}
		result.Rules.Domains = append(result.Rules.Domains, rule)
		result.InputRules++
	}
	return result, nil
}

type v2flyDocument struct {
	Lists []v2flyList `yaml:"lists"`
}

type v2flyList struct {
	Name   string   `yaml:"name"`
	Length int      `yaml:"length"`
	Rules  []string `yaml:"rules"`
}

func (loader *Loader) loadV2Fly(ctx context.Context, source config.Source) (Result, error) {
	if loader.v2fly == nil {
		payload, metadata, err := loader.fetcher.Fetch(ctx, loader.v2flyURL)
		if err != nil {
			return Result{}, err
		}
		var document v2flyDocument
		if err := yaml.Unmarshal(payload, &document); err != nil {
			return Result{}, fmt.Errorf("decode V2Fly YAML: %w", err)
		}
		if len(document.Lists) == 0 {
			return Result{}, fmt.Errorf("V2Fly document has no lists")
		}
		listNames := make(map[string]struct{}, len(document.Lists))
		for _, list := range document.Lists {
			if list.Name == "" || list.Length != len(list.Rules) {
				return Result{}, fmt.Errorf("V2Fly list %q length is %d, expected %d", list.Name, len(list.Rules), list.Length)
			}
			if _, exists := listNames[list.Name]; exists {
				return Result{}, fmt.Errorf("V2Fly document contains duplicate list %q", list.Name)
			}
			listNames[list.Name] = struct{}{}
		}
		loader.v2fly = &document
		loader.v2flyFetch = metadata
	}

	requested := make(map[string]struct{}, len(source.Lists))
	for _, name := range source.Lists {
		requested[name] = struct{}{}
	}
	found := make(map[string]struct{}, len(requested))
	result := Result{Ignored: make(map[string]int), Fetches: []FetchMetadata{loader.v2flyFetch}}
	for _, list := range loader.v2fly.Lists {
		if !source.AllLists {
			if _, selected := requested[list.Name]; !selected {
				continue
			}
			found[list.Name] = struct{}{}
		}
		for _, encoded := range list.Rules {
			kind, value, attributes, err := parseV2FlyRule(encoded)
			if err != nil {
				return Result{}, fmt.Errorf("list %q: %w", list.Name, err)
			}
			if !hasAllAttributes(attributes, source.RequireAttributes) {
				continue
			}
			result.InputRules++
			switch kind {
			case "domain":
				rule, err := rules.NewDomainRule(rules.DomainSuffix, value)
				if err != nil {
					return Result{}, fmt.Errorf("list %q rule %q: %w", list.Name, encoded, err)
				}
				result.Rules.Domains = append(result.Rules.Domains, rule)
			case "full":
				rule, err := rules.NewDomainRule(rules.DomainExact, value)
				if err != nil {
					return Result{}, fmt.Errorf("list %q rule %q: %w", list.Name, encoded, err)
				}
				result.Rules.Domains = append(result.Rules.Domains, rule)
			case "keyword":
				rule, err := rules.NewClassicalRule("DOMAIN-KEYWORD", value)
				if err != nil {
					return Result{}, fmt.Errorf("list %q rule %q: %w", list.Name, encoded, err)
				}
				result.Rules.Classical = append(result.Rules.Classical, rule)
			case "regexp":
				rule, err := rules.NewClassicalRule("DOMAIN-REGEX", value)
				if err != nil {
					return Result{}, fmt.Errorf("list %q rule %q: %w", list.Name, encoded, err)
				}
				result.Rules.Classical = append(result.Rules.Classical, rule)
			default:
				return Result{}, fmt.Errorf("list %q: unsupported V2Fly rule kind %q", list.Name, kind)
			}
		}
	}
	for name := range requested {
		if _, exists := found[name]; !exists {
			return Result{}, fmt.Errorf("V2Fly list %q not found", name)
		}
	}
	return result, nil
}

func parseV2FlyRule(encoded string) (string, string, map[string]struct{}, error) {
	separator := strings.IndexByte(encoded, ':')
	if separator <= 0 || separator == len(encoded)-1 {
		return "", "", nil, fmt.Errorf("malformed V2Fly rule %q", encoded)
	}
	kind := encoded[:separator]
	parts := strings.Split(encoded[separator+1:], ":@")
	value := parts[0]
	attributes := make(map[string]struct{}, len(parts)-1)
	for _, attribute := range parts[1:] {
		if attribute == "" {
			return "", "", nil, fmt.Errorf("empty attribute in V2Fly rule %q", encoded)
		}
		attributes[attribute] = struct{}{}
	}
	return kind, value, attributes, nil
}

func hasAllAttributes(attributes map[string]struct{}, required []string) bool {
	for _, attribute := range required {
		if _, exists := attributes[attribute]; !exists {
			return false
		}
	}
	return true
}

var ianaLocalPrefixes = map[string]struct{}{
	"10.0.0.0/8":     {},
	"127.0.0.0/8":    {},
	"169.254.0.0/16": {},
	"172.16.0.0/12":  {},
	"192.168.0.0/16": {},
	"::1/128":        {},
	"fc00::/7":       {},
	"fe80::/10":      {},
}

func (loader *Loader) loadIANALocal(ctx context.Context, urls []string) (Result, error) {
	result := Result{Ignored: make(map[string]int)}
	found := make(map[string]struct{}, len(ianaLocalPrefixes))
	for _, url := range urls {
		payload, metadata, err := loader.fetcher.Fetch(ctx, url)
		if err != nil {
			return Result{}, err
		}
		result.Fetches = append(result.Fetches, metadata)
		reader := csv.NewReader(strings.NewReader(string(payload)))
		records, err := reader.ReadAll()
		if err != nil {
			return Result{}, fmt.Errorf("decode IANA CSV %s: %w", url, err)
		}
		if len(records) < 2 || len(records[0]) == 0 || records[0][0] != "Address Block" {
			return Result{}, fmt.Errorf("unexpected IANA CSV header in %s", url)
		}
		for _, record := range records[1:] {
			if len(record) == 0 {
				continue
			}
			fields := strings.Fields(record[0])
			if len(fields) == 0 {
				return Result{}, fmt.Errorf("IANA CSV %s contains an empty address block", url)
			}
			block := fields[0]
			if _, selected := ianaLocalPrefixes[block]; !selected {
				continue
			}
			prefix, err := netip.ParsePrefix(block)
			if err != nil {
				return Result{}, fmt.Errorf("parse selected IANA prefix %q: %w", block, err)
			}
			result.Rules.Prefixes = append(result.Rules.Prefixes, prefix.Masked())
			result.InputRules++
			found[block] = struct{}{}
		}
	}
	missing := make([]string, 0)
	for prefix := range ianaLocalPrefixes {
		if _, exists := found[prefix]; !exists {
			missing = append(missing, prefix)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return Result{}, fmt.Errorf("IANA registries are missing selected prefixes: %s", strings.Join(missing, ", "))
	}
	return result, nil
}
