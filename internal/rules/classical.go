package rules

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// ClassicalSet stores canonical Mihomo classical rules formatted as "KIND,value".
type ClassicalSet []string

var globReplacer = strings.NewReplacer(`\\`, `\\`, `\*`, `.*`, `\?`, `.`)

func NewClassical(kind, value string) (string, error) {
	kind = strings.ToUpper(strings.TrimSpace(kind))
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("invalid %s rule %q", kind, value)
	}
	switch kind {
	case "DOMAIN-KEYWORD", "DOMAIN-WILDCARD",
		"GEOSITE", "GEOIP", "SRC-GEOIP", "IP-ASN", "SRC-IP-ASN",
		"IP-CIDR", "IP-CIDR6", "SRC-IP-CIDR", "IP-SUFFIX", "SRC-IP-SUFFIX",
		"SRC-PORT", "IN-PORT", "DSCP", "PROCESS-NAME", "PROCESS-PATH",
		"PROCESS-NAME-REGEX", "PROCESS-PATH-REGEX", "PROCESS-NAME-WILDCARD", "PROCESS-PATH-WILDCARD",
		"NETWORK", "UID", "IN-TYPE", "IN-USER", "IN-NAME", "REMATCH-NAME", "AND", "OR", "NOT":
	case "DOMAIN-REGEX":
		if _, err := regexp.Compile(value); err != nil {
			return "", fmt.Errorf("compile domain regex %q: %w", value, err)
		}
	case "DST-PORT":
		payload, _, _ := strings.Cut(value, ",")
		if _, err := parsePortRanges(payload); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported classical rule kind %q", kind)
	}
	return kind + "," + value, nil
}

func (set ClassicalSet) Matches(domain string) bool {
	domain = normalizeLookupDomain(domain)
	for _, rule := range set {
		if matchClassicalDomain(rule, domain) {
			return true
		}
	}
	return false
}

func (set ClassicalSet) MatchesDestinationPort(port uint16) bool {
	for _, rule := range set {
		if matchClassicalPort(rule, port) {
			return true
		}
	}
	return false
}

func (set ClassicalSet) Optimized() ClassicalSet {
	if len(set) == 0 {
		return nil
	}
	result := slices.Clone(set)
	slices.Sort(result)
	return slices.Compact(result)
}

func (set ClassicalSet) Excluding(earlier ClassicalSet) ClassicalSet {
	if len(set) == 0 || len(earlier) == 0 {
		return set.Optimized()
	}
	seen := make(map[string]struct{}, len(earlier))
	for _, rule := range earlier {
		seen[rule] = struct{}{}
	}
	result := make(ClassicalSet, 0, len(set))
	for _, rule := range set {
		if _, exists := seen[rule]; !exists {
			result = append(result, rule)
		}
	}
	return result.Optimized()
}

func matchClassicalDomain(rule, domain string) bool {
	kind, value, _ := strings.Cut(rule, ",")
	switch kind {
	case "DOMAIN-KEYWORD":
		return strings.Contains(domain, strings.ToLower(value))
	case "DOMAIN-REGEX":
		return regexp.MustCompile(value).MatchString(domain)
	case "DOMAIN-WILDCARD":
		expr := "^" + globReplacer.Replace(regexp.QuoteMeta(strings.ToLower(value))) + "$"
		return regexp.MustCompile(expr).MatchString(domain)
	default:
		// Contextual rules require connection metadata beyond a domain.
		return false
	}
}

func matchClassicalPort(rule string, port uint16) bool {
	value, ok := strings.CutPrefix(rule, "DST-PORT,")
	if !ok || strings.Contains(value, ",") {
		return false
	}
	ranges, err := parsePortRanges(value)
	if err != nil {
		panic(err)
	}
	for _, segment := range ranges {
		if port >= segment[0] && port <= segment[1] {
			return true
		}
	}
	return false
}

func parsePortRanges(value string) ([][2]uint16, error) {
	terms := strings.Split(value, "/")
	if len(terms) > 28 {
		return nil, fmt.Errorf("DST-PORT supports at most 28 port ranges")
	}
	result := make([][2]uint16, 0, len(terms))
	for _, term := range terms {
		startText, endText, hasEnd := strings.Cut(term, "-")
		if hasEnd && strings.Contains(endText, "-") {
			return nil, fmt.Errorf("invalid DST-PORT range %q", term)
		}
		start, err := strconv.ParseUint(strings.TrimSpace(startText), 10, 16)
		if err != nil {
			return nil, fmt.Errorf("invalid DST-PORT range %q: %w", term, err)
		}
		end := start
		if hasEnd {
			if end, err = strconv.ParseUint(strings.TrimSpace(endText), 10, 16); err != nil {
				return nil, fmt.Errorf("invalid DST-PORT range %q: %w", term, err)
			}
		}
		if start > end {
			start, end = end, start
		}
		result = append(result, [2]uint16{uint16(start), uint16(end)})
	}
	return result, nil
}
