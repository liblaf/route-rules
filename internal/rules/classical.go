package rules

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ClassicalRule struct {
	Kind  string
	Value string
}

func NewClassicalRule(kind, value string) (ClassicalRule, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return ClassicalRule{}, fmt.Errorf("invalid %s rule %q", kind, value)
	}
	switch kind {
	case "DOMAIN-KEYWORD", "DOMAIN-REGEX", "DOMAIN-WILDCARD",
		"GEOSITE", "GEOIP", "SRC-GEOIP", "IP-ASN", "SRC-IP-ASN",
		"IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX",
		"SRC-PORT", "IN-PORT", "DSCP", "PROCESS-NAME", "PROCESS-PATH",
		"PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX", "PROCESS-NAME-WILDCARD", "PROCESS-PATH-WILDCARD",
		"NETWORK", "UID", "IN-TYPE", "IN-USER", "IN-NAME", "REMATCH-NAME", "AND", "OR", "NOT":
	case "DST-PORT":
		payload, _, _ := strings.Cut(value, ",")
		if _, err := parsePortRanges(payload); err != nil {
			return ClassicalRule{}, err
		}
	default:
		return ClassicalRule{}, fmt.Errorf("unsupported classical rule kind %q", kind)
	}
	if kind == "DOMAIN-REGEX" {
		if _, err := regexp.Compile(value); err != nil {
			return ClassicalRule{}, fmt.Errorf("compile domain regex %q: %w", value, err)
		}
	}
	return ClassicalRule{Kind: kind, Value: value}, nil
}

func (r ClassicalRule) Text() string {
	return r.Kind + "," + r.Value
}

func (r ClassicalRule) Matches(domain string) bool {
	domain = normalizeLookupDomain(domain)
	switch r.Kind {
	case "DOMAIN-KEYWORD":
		return strings.Contains(domain, strings.ToLower(r.Value))
	case "DOMAIN-REGEX":
		matched, err := regexp.MatchString(r.Value, domain)
		if err != nil {
			panic(err)
		}
		return matched
	case "DOMAIN-WILDCARD":
		return globMatches(strings.ToLower(r.Value), domain)
	default:
		// Contextual rules require connection metadata beyond a domain.
		return false
	}
}

func (r ClassicalRule) MatchesDestinationPort(port uint16) bool {
	if r.Kind != "DST-PORT" || strings.Contains(r.Value, ",") {
		return false
	}
	ranges, err := parsePortRanges(r.Value)
	if err != nil {
		panic(err)
	}
	for _, segment := range ranges {
		if port >= segment.start && port <= segment.end {
			return true
		}
	}
	return false
}

type portRange struct {
	start uint16
	end   uint16
}

func parsePortRanges(value string) ([]portRange, error) {
	terms := strings.Split(value, "/")
	if len(terms) > 28 {
		return nil, fmt.Errorf("DST-PORT supports at most 28 port ranges")
	}
	result := make([]portRange, 0, len(terms))
	for _, term := range terms {
		bounds := strings.Split(term, "-")
		if len(bounds) > 2 {
			return nil, fmt.Errorf("invalid DST-PORT range %q", term)
		}
		start, err := strconv.ParseUint(strings.TrimSpace(bounds[0]), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid DST-PORT range %q: %w", term, err)
		}
		end := start
		if len(bounds) == 2 {
			end, err = strconv.ParseUint(strings.TrimSpace(bounds[1]), 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid DST-PORT range %q: %w", term, err)
			}
		}
		if start > end {
			start, end = end, start
		}
		result = append(result, portRange{start: uint16(start), end: uint16(end)})
	}
	return result, nil
}

type ClassicalSet []ClassicalRule

func (set ClassicalSet) Matches(domain string) bool {
	for _, rule := range set {
		if rule.Matches(domain) {
			return true
		}
	}
	return false
}

func (set ClassicalSet) MatchesDestinationPort(port uint16) bool {
	for _, rule := range set {
		if rule.MatchesDestinationPort(port) {
			return true
		}
	}
	return false
}

func (set ClassicalSet) Optimized() ClassicalSet {
	unique := make(map[string]ClassicalRule, len(set))
	for _, rule := range set {
		unique[rule.Text()] = rule
	}
	result := make(ClassicalSet, 0, len(unique))
	for _, rule := range unique {
		result = append(result, rule)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Text() < result[j].Text() })
	return result
}

func (set ClassicalSet) Excluding(earlier ClassicalSet) ClassicalSet {
	seen := make(map[string]struct{}, len(earlier))
	for _, rule := range earlier {
		seen[rule.Text()] = struct{}{}
	}
	result := make(ClassicalSet, 0, len(set))
	for _, rule := range set {
		if _, exists := seen[rule.Text()]; !exists {
			result = append(result, rule)
		}
	}
	return result.Optimized()
}

func globMatches(pattern, value string) bool {
	var expression strings.Builder
	expression.WriteByte('^')
	for _, char := range pattern {
		switch char {
		case '*':
			expression.WriteString(".*")
		case '?':
			expression.WriteByte('.')
		default:
			expression.WriteString(regexp.QuoteMeta(string(char)))
		}
	}
	expression.WriteByte('$')
	return regexp.MustCompile(expression.String()).MatchString(value)
}
