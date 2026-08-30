package source

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/liblaf/route-rules/internal/rules"
)

var ignoredClashKinds = map[string]struct{}{
	"AND":          {},
	"DST-PORT":     {},
	"FINAL":        {},
	"GEOIP":        {},
	"GEOSITE":      {},
	"IP-ASN":       {},
	"IN-PORT":      {},
	"MATCH":        {},
	"NETWORK":      {},
	"NOT":          {},
	"OR":           {},
	"PROCESS-NAME": {},
	"PROCESS-PATH": {},
	"RULE-SET":     {},
	"SRC-IP-CIDR":  {},
	"SRC-PORT":     {},
	"URL-REGEX":    {},
	"USER-AGENT":   {},
}

func parseClash(payload string) (Result, error) {
	result := Result{Ignored: make(map[string]int)}
	for number, raw := range strings.Split(payload, "\n") {
		line := normalizeClashLine(raw)
		if line == "" || line == "payload:" {
			continue
		}
		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 2 {
			return Result{}, fmt.Errorf("line %d: malformed Clash rule %q", number+1, line)
		}
		kind := strings.ToUpper(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		result.InputRules++
		switch kind {
		case "DOMAIN":
			rule, err := rules.NewDomainRule(rules.DomainExact, value)
			if err != nil {
				return Result{}, fmt.Errorf("line %d: %w", number+1, err)
			}
			result.Rules.Domains = append(result.Rules.Domains, rule)
		case "DOMAIN-SUFFIX":
			rule, err := rules.NewDomainRule(rules.DomainSuffix, value)
			if err != nil {
				return Result{}, fmt.Errorf("line %d: %w", number+1, err)
			}
			result.Rules.Domains = append(result.Rules.Domains, rule)
		case "DOMAIN-WILDCARD":
			if rules.IsWholeLabelWildcard(value) {
				rule, err := rules.NewDomainRule(rules.DomainWildcard, value)
				if err != nil {
					return Result{}, fmt.Errorf("line %d: %w", number+1, err)
				}
				result.Rules.Domains = append(result.Rules.Domains, rule)
			} else {
				rule, err := rules.NewClassicalRule(kind, value)
				if err != nil {
					return Result{}, fmt.Errorf("line %d: %w", number+1, err)
				}
				result.Rules.Classical = append(result.Rules.Classical, rule)
			}
		case "DOMAIN-KEYWORD", "DOMAIN-REGEX":
			rule, err := rules.NewClassicalRule(kind, value)
			if err != nil {
				return Result{}, fmt.Errorf("line %d: %w", number+1, err)
			}
			result.Rules.Classical = append(result.Rules.Classical, rule)
		case "IP-CIDR", "IP-CIDR6":
			prefix, err := netip.ParsePrefix(value)
			if err != nil {
				return Result{}, fmt.Errorf("line %d: parse prefix %q: %w", number+1, value, err)
			}
			if kind == "IP-CIDR" && !prefix.Addr().Is4() || kind == "IP-CIDR6" && !prefix.Addr().Is6() {
				return Result{}, fmt.Errorf("line %d: %s has wrong address family", number+1, kind)
			}
			result.Rules.Prefixes = append(result.Rules.Prefixes, prefix.Masked())
		default:
			if _, ignored := ignoredClashKinds[kind]; !ignored {
				return Result{}, fmt.Errorf("line %d: unknown Clash rule kind %q", number+1, kind)
			}
			result.Ignored[kind]++
		}
	}
	return result, nil
}

func normalizeClashLine(raw string) string {
	line := strings.TrimSpace(raw)
	if line == "" || strings.HasPrefix(line, "#") {
		return ""
	}
	if strings.HasPrefix(line, "-") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
	}
	if len(line) >= 2 {
		if line[0] == '\'' && line[len(line)-1] == '\'' || line[0] == '"' && line[len(line)-1] == '"' {
			line = line[1 : len(line)-1]
		}
	}
	return strings.TrimSpace(line)
}
