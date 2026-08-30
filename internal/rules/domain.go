package rules

import (
	"fmt"
	"sort"
	"strings"
)

type DomainKind uint8

const (
	DomainExact DomainKind = iota
	DomainSuffix
	DomainWildcard
)

type DomainRule struct {
	Kind  DomainKind
	Value string
}

func NewDomainRule(kind DomainKind, value string) (DomainRule, error) {
	value = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(value), "."))
	value = strings.TrimPrefix(value, "+.")
	if value == "" || strings.ContainsAny(value, " \t\r\n,/") {
		return DomainRule{}, fmt.Errorf("invalid domain %q", value)
	}
	if strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") || strings.Contains(value, "..") {
		return DomainRule{}, fmt.Errorf("invalid domain %q", value)
	}
	if kind != DomainWildcard && strings.Contains(value, "*") {
		return DomainRule{}, fmt.Errorf("wildcard in non-wildcard domain %q", value)
	}
	if kind == DomainWildcard && !IsWholeLabelWildcard(value) {
		return DomainRule{}, fmt.Errorf("wildcard must occupy a complete label: %q", value)
	}
	return DomainRule{Kind: kind, Value: value}, nil
}

func ParseDomainLine(line string) (DomainRule, error) {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, "+."):
		return NewDomainRule(DomainSuffix, strings.TrimPrefix(line, "+."))
	case strings.Contains(line, "*"):
		return NewDomainRule(DomainWildcard, line)
	default:
		return NewDomainRule(DomainExact, line)
	}
}

func IsWholeLabelWildcard(value string) bool {
	if !strings.Contains(value, "*") || strings.Contains(value, "?") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if strings.ContainsAny(label, "*?") && label != "*" {
			return false
		}
	}
	return true
}

func (r DomainRule) Text() string {
	if r.Kind == DomainSuffix {
		return "+." + r.Value
	}
	return r.Value
}

func (r DomainRule) Matches(domain string) bool {
	domain = normalizeLookupDomain(domain)
	switch r.Kind {
	case DomainExact:
		return domain == r.Value
	case DomainSuffix:
		return domain == r.Value || strings.HasSuffix(domain, "."+r.Value)
	case DomainWildcard:
		return wildcardMatches(r.Value, domain)
	default:
		panic(fmt.Sprintf("unknown domain kind %d", r.Kind))
	}
}

type DomainSet []DomainRule

func (set DomainSet) Matches(domain string) bool {
	for _, rule := range set {
		if rule.Matches(domain) {
			return true
		}
	}
	return false
}

func (set DomainSet) Optimized() DomainSet {
	exacts := make(map[string]DomainRule)
	suffixes := make(map[string]DomainRule)
	wildcards := make(map[string]DomainRule)
	for _, rule := range set {
		switch rule.Kind {
		case DomainExact:
			exacts[rule.Value] = rule
		case DomainSuffix:
			suffixes[rule.Value] = rule
		case DomainWildcard:
			wildcards[rule.Value] = rule
		default:
			panic(fmt.Sprintf("unknown domain kind %d", rule.Kind))
		}
	}

	for value := range suffixes {
		if hasStrictSuffixAncestor(suffixes, value) {
			delete(suffixes, value)
		}
	}
	for value := range exacts {
		if hasSuffixAncestor(suffixes, value) || anyWildcardMatches(wildcards, value) {
			delete(exacts, value)
		}
	}
	for value := range wildcards {
		if wildcardCoveredBySuffix(suffixes, value) {
			delete(wildcards, value)
		}
	}

	result := make(DomainSet, 0, len(exacts)+len(suffixes)+len(wildcards))
	for _, rule := range exacts {
		result = append(result, rule)
	}
	for _, rule := range suffixes {
		result = append(result, rule)
	}
	for _, rule := range wildcards {
		result = append(result, rule)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Text() < result[j].Text() })
	return result
}

// Excluding drops only rules whose entire match space is covered by an earlier
// set. Partial suffix differences cannot be represented by Mihomo's domain
// rules and remain safe because consumers must evaluate sets in priority order.
func (set DomainSet) Excluding(earlier DomainSet) DomainSet {
	earlier = earlier.Optimized()
	suffixes := make(map[string]DomainRule)
	wildcards := make(map[string]DomainRule)
	exacts := make(map[string]struct{})
	for _, rule := range earlier {
		switch rule.Kind {
		case DomainExact:
			exacts[rule.Value] = struct{}{}
		case DomainSuffix:
			suffixes[rule.Value] = rule
		case DomainWildcard:
			wildcards[rule.Value] = rule
		}
	}

	result := make(DomainSet, 0, len(set))
	for _, rule := range set {
		covered := false
		switch rule.Kind {
		case DomainExact:
			_, exact := exacts[rule.Value]
			covered = exact || hasSuffixAncestor(suffixes, rule.Value) || anyWildcardMatches(wildcards, rule.Value)
		case DomainSuffix:
			covered = hasSuffixAncestor(suffixes, rule.Value)
		case DomainWildcard:
			_, same := wildcards[rule.Value]
			covered = same || wildcardCoveredBySuffix(suffixes, rule.Value)
		}
		if !covered {
			result = append(result, rule)
		}
	}
	return result.Optimized()
}

func hasSuffixAncestor(suffixes map[string]DomainRule, value string) bool {
	for current := value; ; {
		if _, exists := suffixes[current]; exists {
			return true
		}
		dot := strings.IndexByte(current, '.')
		if dot < 0 {
			return false
		}
		current = current[dot+1:]
	}
}

func hasStrictSuffixAncestor(suffixes map[string]DomainRule, value string) bool {
	dot := strings.IndexByte(value, '.')
	if dot < 0 {
		return false
	}
	return hasSuffixAncestor(suffixes, value[dot+1:])
}

func anyWildcardMatches(wildcards map[string]DomainRule, value string) bool {
	for _, rule := range wildcards {
		if rule.Matches(value) {
			return true
		}
	}
	return false
}

func wildcardCoveredBySuffix(suffixes map[string]DomainRule, value string) bool {
	labels := strings.Split(value, ".")
	lastWildcard := -1
	for index, label := range labels {
		if label == "*" {
			lastWildcard = index
		}
	}
	if lastWildcard == len(labels)-1 {
		return false
	}
	return hasSuffixAncestor(suffixes, strings.Join(labels[lastWildcard+1:], "."))
}

func wildcardMatches(pattern, domain string) bool {
	patternLabels := strings.Split(pattern, ".")
	domainLabels := strings.Split(domain, ".")
	if len(patternLabels) != len(domainLabels) {
		return false
	}
	for index := range patternLabels {
		if patternLabels[index] != "*" && patternLabels[index] != domainLabels[index] {
			return false
		}
	}
	return true
}

func normalizeLookupDomain(domain string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
}
