package rules

import (
	"fmt"
	"slices"
	"strings"
)

// DomainSet stores canonical Mihomo domain rules: exact ("example.com"),
// suffix ("+.example.com"), or whole-label wildcard ("*.example.com").
type DomainSet []string

func ExactDomain(value string) (string, error) {
	value, err := validateDomain(value)
	if err != nil {
		return "", err
	}
	if strings.Contains(value, "*") {
		return "", fmt.Errorf("wildcard in non-wildcard domain %q", value)
	}
	return value, nil
}

func SuffixDomain(value string) (string, error) {
	value, err := ExactDomain(value)
	if err != nil {
		return "", err
	}
	return "+." + value, nil
}

func WildcardDomain(value string) (string, error) {
	value, err := validateDomain(value)
	if err != nil {
		return "", err
	}
	if !IsWholeLabelWildcard(value) {
		return "", fmt.Errorf("wildcard must occupy a complete label: %q", value)
	}
	return value, nil
}

func ParseDomain(line string) (string, error) {
	line = strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(line, "+."):
		return SuffixDomain(line[2:])
	case strings.Contains(line, "*"):
		return WildcardDomain(line)
	default:
		return ExactDomain(line)
	}
}

func IsWholeLabelWildcard(value string) bool {
	if !strings.Contains(value, "*") || strings.Contains(value, "?") {
		return false
	}
	for label := range strings.SplitSeq(value, ".") {
		if strings.ContainsAny(label, "*?") && label != "*" {
			return false
		}
	}
	return true
}

func NormalizeDomain(domain string) string {
	return strings.TrimPrefix(normalizeLookupDomain(domain), "+.")
}

func (set DomainSet) Matches(domain string) bool {
	domain = normalizeLookupDomain(domain)
	for _, rule := range set {
		if suffix, ok := strings.CutPrefix(rule, "+."); ok {
			if domain == suffix || strings.HasSuffix(domain, "."+suffix) {
				return true
			}
		} else if strings.Contains(rule, "*") {
			if wildcardMatches(rule, domain) {
				return true
			}
		} else if domain == rule {
			return true
		}
	}
	return false
}

func (set DomainSet) Optimized() DomainSet {
	if len(set) == 0 {
		return nil
	}
	exacts, suffixes, wildcards := indexDomains(set)
	for value := range suffixes {
		if dot := strings.IndexByte(value, '.'); dot >= 0 && hasSuffixAncestor(suffixes, value[dot+1:]) {
			delete(suffixes, value)
		}
	}
	for value := range wildcards {
		if wildcardCoveredBySuffix(suffixes, value) {
			delete(wildcards, value)
		}
	}
	for value := range exacts {
		if hasSuffixAncestor(suffixes, value) || anyWildcardMatches(wildcards, value) {
			delete(exacts, value)
		}
	}

	result := make(DomainSet, 0, len(exacts)+len(suffixes)+len(wildcards))
	for value := range exacts {
		result = append(result, value)
	}
	for value := range suffixes {
		result = append(result, "+."+value)
	}
	for value := range wildcards {
		result = append(result, value)
	}
	slices.Sort(result)
	return result
}

// Excluding drops only rules whose entire match space is covered by an earlier
// set. Partial suffix differences cannot be represented by Mihomo's domain
// rules and remain safe because consumers must evaluate sets in priority order.
func (set DomainSet) Excluding(earlier DomainSet) DomainSet {
	if len(set) == 0 || len(earlier) == 0 {
		return set.Optimized()
	}
	exacts, suffixes, wildcards := indexDomains(earlier)
	result := make(DomainSet, 0, len(set))
	for _, rule := range set {
		var covered bool
		if suffix, ok := strings.CutPrefix(rule, "+."); ok {
			covered = hasSuffixAncestor(suffixes, suffix)
		} else if strings.Contains(rule, "*") {
			_, same := wildcards[rule]
			covered = same || wildcardCoveredBySuffix(suffixes, rule)
		} else {
			_, exact := exacts[rule]
			covered = exact || hasSuffixAncestor(suffixes, rule) || anyWildcardMatches(wildcards, rule)
		}
		if !covered {
			result = append(result, rule)
		}
	}
	return result.Optimized()
}

func validateDomain(value string) (string, error) {
	value = NormalizeDomain(value)
	if value == "" || strings.HasPrefix(value, "+.") || strings.ContainsAny(value, " \t\r\n,/") ||
		strings.HasPrefix(value, ".") || strings.HasSuffix(value, ".") || strings.Contains(value, "..") {
		return "", fmt.Errorf("invalid domain %q", value)
	}
	return value, nil
}

func indexDomains(set DomainSet) (exacts, suffixes, wildcards map[string]struct{}) {
	exacts = make(map[string]struct{})
	suffixes = make(map[string]struct{})
	wildcards = make(map[string]struct{})
	for _, rule := range set {
		if suffix, ok := strings.CutPrefix(rule, "+."); ok {
			suffixes[suffix] = struct{}{}
		} else if strings.Contains(rule, "*") {
			wildcards[rule] = struct{}{}
		} else {
			exacts[rule] = struct{}{}
		}
	}
	return exacts, suffixes, wildcards
}

func hasSuffixAncestor(suffixes map[string]struct{}, value string) bool {
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

func anyWildcardMatches(wildcards map[string]struct{}, value string) bool {
	for pattern := range wildcards {
		if wildcardMatches(pattern, value) {
			return true
		}
	}
	return false
}

func wildcardCoveredBySuffix(suffixes map[string]struct{}, value string) bool {
	last := strings.LastIndexByte(value, '*')
	return last >= 0 && last+2 < len(value) && value[last+1] == '.' && hasSuffixAncestor(suffixes, value[last+2:])
}

func wildcardMatches(pattern, domain string) bool {
	for {
		pDot, dDot := strings.IndexByte(pattern, '.'), strings.IndexByte(domain, '.')
		if (pDot < 0) != (dDot < 0) {
			return false
		}
		if pDot < 0 {
			return pattern == "*" || pattern == domain
		}
		if label := pattern[:pDot]; label != "*" && label != domain[:dDot] {
			return false
		}
		pattern, domain = pattern[pDot+1:], domain[dDot+1:]
	}
}

func normalizeLookupDomain(domain string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
}
