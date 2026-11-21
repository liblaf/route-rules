package build

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"iter"
	"maps"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/liblaf/route-rules/internal/rules"
	"gopkg.in/yaml.v3"
)

const maxSourceBytes = 64 << 20

type FetchMetadata struct {
	URL          string `json:"url"`
	Bytes        int64  `json:"bytes"`
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
}

type fetchedDoc struct {
	payload  []byte
	metadata FetchMetadata
}

type v2flyList struct {
	Name   string   `yaml:"name"`
	Length int      `yaml:"length"`
	Rules  []string `yaml:"rules"`
}

type sourceLoader struct {
	client      *http.Client
	cache       map[string]fetchedDoc
	v2flyURL    string
	v2flyLists  []v2flyList
	v2flyByName map[string][]string
	v2flyFetch  FetchMetadata
}

func newSourceLoader(v2flyURL string) *sourceLoader {
	return &sourceLoader{
		client:   &http.Client{Timeout: 45 * time.Second},
		cache:    make(map[string]fetchedDoc),
		v2flyURL: v2flyURL,
	}
}

func (l *sourceLoader) fetch(ctx context.Context, url string) ([]byte, FetchMetadata, error) {
	if cached, exists := l.cache[url]; exists {
		return cached.payload, cached.metadata, nil
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		payload, meta, err := l.fetchOnce(ctx, url)
		if err == nil {
			l.cache[url] = fetchedDoc{payload: payload, metadata: meta}
			return payload, meta, nil
		}
		lastErr = err
		if ctx.Err() == nil && attempt < 3 {
			timer := time.NewTimer(time.Duration(attempt) * 500 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
			case <-timer.C:
			}
		}
		if ctx.Err() != nil {
			return nil, FetchMetadata{}, fmt.Errorf("fetch %s: %w", url, ctx.Err())
		}
	}
	return nil, FetchMetadata{}, fmt.Errorf("fetch %s after 3 attempts: %w", url, lastErr)
}

func (l *sourceLoader) fetchOnce(ctx context.Context, url string) ([]byte, FetchMetadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, FetchMetadata{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "liblaf/route-rules")
	req.Header.Set("Accept", "text/plain, application/yaml, text/csv, */*")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, FetchMetadata{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, FetchMetadata{}, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}
	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxSourceBytes+1))
	if err != nil {
		return nil, FetchMetadata{}, fmt.Errorf("read response: %w", err)
	}
	if len(payload) == 0 {
		return nil, FetchMetadata{}, fmt.Errorf("empty response")
	}
	if len(payload) > maxSourceBytes {
		return nil, FetchMetadata{}, fmt.Errorf("response exceeds %d bytes", maxSourceBytes)
	}
	return payload, FetchMetadata{
		URL:          url,
		Bytes:        int64(len(payload)),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}, nil
}

func (l *sourceLoader) load(ctx context.Context, baseDir string, src SourceConfig) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	var err error
	switch src.Kind {
	case "clash", "clash-file":
		loaded, report, err = l.loadText(ctx, baseDir, src, parseClash)
	case "dnsmasq":
		loaded, report, err = l.loadText(ctx, baseDir, src, parseDNSMasq)
	case "cidr", "cidr-file":
		loaded, report, err = l.loadText(ctx, baseDir, src, parseCIDR)
	case "domain-file":
		loaded, report, err = l.loadText(ctx, baseDir, src, parseDomainText)
	case "v2fly":
		loaded, report, err = l.loadV2Fly(ctx, src)
	case "iana-local":
		loaded, report, err = l.loadIANALocal(ctx, src.URLs)
	default:
		return rules.Collection{}, SourceReport{}, fmt.Errorf("load %s: unknown source kind %q", src.Name, src.Kind)
	}
	if err != nil {
		return rules.Collection{}, SourceReport{}, fmt.Errorf("load %s: %w", src.Name, err)
	}
	report.Name, report.Kind = src.Name, src.Kind
	return loaded.Optimized(), report, nil
}

func (l *sourceLoader) loadText(
	ctx context.Context,
	baseDir string,
	src SourceConfig,
	parse func(string) (rules.Collection, SourceReport, error),
) (rules.Collection, SourceReport, error) {
	var payload []byte
	var fetches []FetchMetadata
	var err error
	if strings.HasSuffix(src.Kind, "-file") {
		payload, err = os.ReadFile(filepath.Join(baseDir, src.Path))
	} else {
		var meta FetchMetadata
		payload, meta, err = l.fetch(ctx, src.URL)
		fetches = []FetchMetadata{meta}
	}
	if err != nil {
		return rules.Collection{}, SourceReport{}, err
	}
	loaded, report, err := parse(string(payload))
	if err != nil {
		return rules.Collection{}, SourceReport{}, err
	}
	report.Fetches = fetches
	return loaded, report, nil
}

func eachRuleLine(payload string) iter.Seq2[int, string] {
	return func(yield func(int, string) bool) {
		number := 0
		for raw := range strings.SplitSeq(payload, "\n") {
			number++
			line := strings.TrimSpace(raw)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !yield(number, line) {
				return
			}
		}
	}
}

func parseClash(payload string) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	for number, raw := range eachRuleLine(payload) {
		line := normalizeClashLine(raw)
		if line == "" || line == "payload:" {
			continue
		}
		rawKind, rawRemainder, ok := strings.Cut(line, ",")
		if !ok {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: malformed Clash rule %q", number, line)
		}
		kind := strings.ToUpper(strings.TrimSpace(rawKind))
		remainder := strings.TrimSpace(rawRemainder)
		value, options, _ := strings.Cut(remainder, ",")
		value = strings.TrimSpace(value)
		report.InputRules++

		switch kind {
		case "DOMAIN", "DOMAIN-SUFFIX":
			var rule string
			var err error
			if kind == "DOMAIN-SUFFIX" {
				rule, err = rules.SuffixDomain(value)
			} else {
				rule, err = rules.ExactDomain(value)
			}
			if err != nil {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %w", number, err)
			}
			loaded.Domains = append(loaded.Domains, rule)
			continue
		case "DOMAIN-WILDCARD":
			if rules.IsWholeLabelWildcard(value) {
				rule, err := rules.WildcardDomain(value)
				if err != nil {
					return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %w", number, err)
				}
				loaded.Domains = append(loaded.Domains, rule)
				continue
			}
		case "IP-CIDR", "IP-CIDR6":
			prefix, err := netip.ParsePrefix(value)
			if err != nil {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: parse prefix %q: %w", number, value, err)
			}
			if kind == "IP-CIDR" && !prefix.Addr().Is4() || kind == "IP-CIDR6" && !prefix.Addr().Is6() {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %s has wrong address family", number, kind)
			}
			if !hasClashOption(options, "src") {
				loaded.Prefixes = append(loaded.Prefixes, prefix.Masked())
				continue
			}
		case "FINAL", "MATCH", "RULE-SET", "SUB-RULE", "URL-REGEX", "USER-AGENT":
			if report.Ignored == nil {
				report.Ignored = make(map[string]int)
			}
			report.Ignored[kind]++
			continue
		}

		rule, err := rules.NewClassical(kind, remainder)
		if err != nil {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %w", number, err)
		}
		loaded.Classical = append(loaded.Classical, rule)
	}
	return loaded, report, nil
}

func hasClashOption(options, target string) bool {
	for option := range strings.SplitSeq(options, ",") {
		if strings.EqualFold(strings.TrimSpace(option), target) {
			return true
		}
	}
	return false
}

func normalizeClashLine(line string) string {
	if trimmed, ok := strings.CutPrefix(line, "-"); ok {
		line = strings.TrimSpace(trimmed)
	}
	if len(line) >= 2 && ((line[0] == '\'' && line[len(line)-1] == '\'') || (line[0] == '"' && line[len(line)-1] == '"')) {
		line = strings.TrimSpace(line[1 : len(line)-1])
	}
	return line
}

func parseDNSMasq(payload string) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	for number, line := range eachRuleLine(payload) {
		rest, ok := strings.CutPrefix(line, "server=/")
		if !ok {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: unsupported dnsmasq directive %q", number, line)
		}
		domain, _, ok := strings.Cut(rest, "/")
		if !ok || domain == "" {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: malformed dnsmasq server directive", number)
		}
		rule, err := rules.SuffixDomain(domain)
		if err != nil {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %w", number, err)
		}
		loaded.Domains = append(loaded.Domains, rule)
		report.InputRules++
	}
	return loaded, report, nil
}

func parseCIDR(payload string) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	for number, line := range eachRuleLine(payload) {
		prefix, err := netip.ParsePrefix(line)
		if err != nil {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: parse prefix %q: %w", number, line, err)
		}
		loaded.Prefixes = append(loaded.Prefixes, prefix.Masked())
		report.InputRules++
	}
	return loaded, report, nil
}

func parseDomainText(payload string) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	for number, line := range eachRuleLine(payload) {
		rule, err := rules.ParseDomain(line)
		if err != nil {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("line %d: %w", number, err)
		}
		loaded.Domains = append(loaded.Domains, rule)
		report.InputRules++
	}
	return loaded, report, nil
}

func (l *sourceLoader) loadV2Fly(ctx context.Context, src SourceConfig) (rules.Collection, SourceReport, error) {
	if l.v2flyLists == nil {
		payload, meta, err := l.fetch(ctx, l.v2flyURL)
		if err != nil {
			return rules.Collection{}, SourceReport{}, err
		}
		lists, byName, err := parseV2FlyDocument(payload)
		if err != nil {
			return rules.Collection{}, SourceReport{}, err
		}
		l.v2flyLists, l.v2flyByName, l.v2flyFetch = lists, byName, meta
	}
	loaded, report, err := selectV2Fly(l.v2flyLists, l.v2flyByName, src)
	if err != nil {
		return rules.Collection{}, SourceReport{}, err
	}
	report.Fetches = []FetchMetadata{l.v2flyFetch}
	return loaded, report, nil
}

func parseV2FlyDocument(payload []byte) ([]v2flyList, map[string][]string, error) {
	var doc struct {
		Lists []v2flyList `yaml:"lists"`
	}
	if err := yaml.Unmarshal(payload, &doc); err != nil {
		return nil, nil, fmt.Errorf("decode V2Fly YAML: %w", err)
	}
	if len(doc.Lists) == 0 {
		return nil, nil, fmt.Errorf("V2Fly document has no lists")
	}
	byName := make(map[string][]string, len(doc.Lists))
	for _, list := range doc.Lists {
		if list.Name == "" || list.Length != len(list.Rules) {
			return nil, nil, fmt.Errorf("V2Fly list %q length is %d, expected %d", list.Name, len(list.Rules), list.Length)
		}
		if _, exists := byName[list.Name]; exists {
			return nil, nil, fmt.Errorf("V2Fly document contains duplicate list %q", list.Name)
		}
		byName[list.Name] = list.Rules
	}
	return doc.Lists, byName, nil
}

func selectV2Fly(allLists []v2flyList, byName map[string][]string, src SourceConfig) (rules.Collection, SourceReport, error) {
	lists := allLists
	if !src.AllLists {
		lists = make([]v2flyList, 0, len(src.Lists))
		for i, name := range src.Lists {
			if slices.Contains(src.Lists[:i], name) {
				continue
			}
			encoded, exists := byName[name]
			if !exists {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("V2Fly list %q not found", name)
			}
			lists = append(lists, v2flyList{Name: name, Rules: encoded})
		}
	}
	var loaded rules.Collection
	var report SourceReport
	for _, list := range lists {
		for _, encoded := range list.Rules {
			kind, value, attrs, err := parseV2FlyRule(encoded)
			if err != nil {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("list %q: %w", list.Name, err)
			}
			if !hasV2FlyAttributes(attrs, src.RequireAttributes) {
				continue
			}
			report.InputRules++
			var rule string
			switch kind {
			case "domain":
				rule, err = rules.SuffixDomain(value)
				loaded.Domains = append(loaded.Domains, rule)
			case "full":
				rule, err = rules.ExactDomain(value)
				loaded.Domains = append(loaded.Domains, rule)
			case "keyword":
				rule, err = rules.NewClassical("DOMAIN-KEYWORD", value)
				loaded.Classical = append(loaded.Classical, rule)
			case "regexp":
				rule, err = rules.NewClassical("DOMAIN-REGEX", value)
				loaded.Classical = append(loaded.Classical, rule)
			default:
				return rules.Collection{}, SourceReport{}, fmt.Errorf("list %q: unsupported V2Fly rule kind %q", list.Name, kind)
			}
			if err != nil {
				return rules.Collection{}, SourceReport{}, fmt.Errorf("list %q rule %q: %w", list.Name, encoded, err)
			}
		}
	}
	return loaded, report, nil
}

func parseV2FlyRule(encoded string) (kind, value, attrs string, err error) {
	kind, rest, ok := strings.Cut(encoded, ":")
	value, attrs, hasAttrs := strings.Cut(rest, ":@")
	if !ok || kind == "" || value == "" {
		return "", "", "", fmt.Errorf("malformed V2Fly rule %q", encoded)
	}
	if hasAttrs {
		for attr := range strings.SplitSeq(attrs, ",@") {
			if attr == "" || strings.ContainsAny(attr, ",@:") {
				return "", "", "", fmt.Errorf("invalid attribute in V2Fly rule %q", encoded)
			}
		}
	}
	return kind, value, attrs, nil
}

func hasV2FlyAttributes(attrs string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if attrs == "" {
		return false
	}
next:
	for _, want := range required {
		for attr := range strings.SplitSeq(attrs, ",@") {
			if attr == want {
				continue next
			}
		}
		return false
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

func (l *sourceLoader) loadIANALocal(ctx context.Context, urls []string) (rules.Collection, SourceReport, error) {
	var loaded rules.Collection
	var report SourceReport
	remaining := maps.Clone(ianaLocalPrefixes)
	for _, url := range urls {
		payload, meta, err := l.fetch(ctx, url)
		if err != nil {
			return rules.Collection{}, SourceReport{}, err
		}
		report.Fetches = append(report.Fetches, meta)
		prefixes, err := parseIANALocalCSV(payload)
		if err != nil {
			return rules.Collection{}, SourceReport{}, fmt.Errorf("IANA CSV %s: %w", url, err)
		}
		for _, prefix := range prefixes {
			delete(remaining, prefix.String())
		}
		loaded.Prefixes = append(loaded.Prefixes, prefixes...)
		report.InputRules += len(prefixes)
	}
	if len(remaining) != 0 {
		return rules.Collection{}, SourceReport{}, fmt.Errorf("IANA registries are missing selected prefixes: %s", strings.Join(slices.Sorted(maps.Keys(remaining)), ", "))
	}
	return loaded, report, nil
}

func parseIANALocalCSV(payload []byte) ([]netip.Prefix, error) {
	records, err := csv.NewReader(bytes.NewReader(payload)).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("decode CSV: %w", err)
	}
	if len(records) < 2 || len(records[0]) == 0 || records[0][0] != "Address Block" {
		return nil, fmt.Errorf("unexpected CSV header")
	}
	var prefixes []netip.Prefix
	for _, record := range records[1:] {
		if len(record) == 0 {
			continue
		}
		fields := strings.Fields(record[0])
		if len(fields) == 0 {
			return nil, fmt.Errorf("empty address block")
		}
		if _, selected := ianaLocalPrefixes[fields[0]]; !selected {
			continue
		}
		prefix, err := netip.ParsePrefix(fields[0])
		if err != nil {
			return nil, fmt.Errorf("parse selected prefix %q: %w", fields[0], err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, nil
}
