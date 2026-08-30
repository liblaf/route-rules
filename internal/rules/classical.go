package rules

import (
	"fmt"
	"regexp"
	"sort"
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
	case "DOMAIN-KEYWORD", "DOMAIN-REGEX", "DOMAIN-WILDCARD":
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
		panic(fmt.Sprintf("unknown classical kind %q", r.Kind))
	}
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
