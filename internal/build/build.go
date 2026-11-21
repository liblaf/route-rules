package build

import (
	"context"
	"fmt"
	"net/netip"
	"slices"
	"strings"

	"github.com/liblaf/route-rules/internal/rules"
)

type Counts struct {
	Domain    int `json:"domain"`
	Classical int `json:"classical"`
	IPCIDR    int `json:"ipcidr"`
}

func (c Counts) Total() int {
	return c.Domain + c.Classical + c.IPCIDR
}

type SourceReport struct {
	Name             string          `json:"name"`
	Kind             string          `json:"kind"`
	InputRules       int             `json:"input_rules"`
	Accepted         Counts          `json:"accepted"`
	Ignored          map[string]int  `json:"ignored,omitempty"`
	Fetches          []FetchMetadata `json:"fetches,omitempty"`
	OmittedAsCovered int             `json:"omitted_as_covered,omitempty"`
}

type RuleSet struct {
	Name               string           `json:"name"`
	Rules              rules.Collection `json:"-"`
	Sources            []SourceReport   `json:"sources"`
	BeforeOptimization Counts           `json:"before_optimization"`
	AfterOptimization  Counts           `json:"after_optimization"`
	Final              Counts           `json:"final"`
	Excluded           []string         `json:"excluded,omitempty"`
}

type Result struct {
	RuleSets []RuleSet
}

func Run(ctx context.Context, cfg Config) (Result, error) {
	loader := newSourceLoader(cfg.V2FlyURL)
	result := Result{RuleSets: make([]RuleSet, 0, len(cfg.RuleSets))}
	byName := make(map[string]rules.Collection, len(cfg.RuleSets))
	ignored := make(map[string]struct{}, len(cfg.IgnoreDomains))
	for _, domain := range cfg.IgnoreDomains {
		ignored[rules.NormalizeDomain(domain)] = struct{}{}
	}

	for _, configured := range cfg.RuleSets {
		built := RuleSet{Name: configured.Name, Excluded: slices.Clone(configured.Exclude)}
		var aggregate rules.Collection
		for _, src := range configured.Sources {
			loaded, report, err := loader.load(ctx, cfg.BaseDir, src)
			if err != nil {
				return Result{}, fmt.Errorf("ruleset %q: %w", configured.Name, err)
			}
			if len(ignored) != 0 {
				loaded.Domains = slices.DeleteFunc(loaded.Domains, func(rule string) bool {
					_, exists := ignored[strings.TrimPrefix(rule, "+.")]
					return exists
				})
			}
			beforeOptional := counts(loaded)
			if src.OptionalIfCovered {
				loaded = loaded.Excluding(aggregate)
			}
			report.Accepted = counts(loaded)
			report.OmittedAsCovered = beforeOptional.Total() - report.Accepted.Total()
			built.Sources = append(built.Sources, report)
			aggregate.Append(loaded)
		}

		built.BeforeOptimization = counts(aggregate)
		aggregate = aggregate.Optimized()
		built.AfterOptimization = counts(aggregate)
		for _, excludedName := range configured.Exclude {
			excluded, exists := byName[excludedName]
			if !exists {
				return Result{}, fmt.Errorf("ruleset %q excludes unbuilt ruleset %q", configured.Name, excludedName)
			}
			aggregate = aggregate.Excluding(excluded)
		}
		built.Rules = aggregate
		built.Final = counts(aggregate)
		result.RuleSets = append(result.RuleSets, built)
		byName[configured.Name] = aggregate
	}

	if err := verifyAssertions(cfg, result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func verifyAssertions(cfg Config, built Result) error {
	byName := make(map[string]rules.Collection, len(built.RuleSets))
	for _, ruleset := range built.RuleSets {
		byName[ruleset.Name] = ruleset.Rules
	}
	for _, a := range cfg.Assertions {
		kind, target := "domain", a.Domain
		var addr netip.Addr
		if a.IP != "" {
			kind, target = "IP", a.IP
			var err error
			if addr, err = netip.ParseAddr(a.IP); err != nil {
				return fmt.Errorf("assertion IP %q: %w", a.IP, err)
			}
		}
		if a.DestinationPort != 0 {
			target = fmt.Sprintf("%s port %d", target, a.DestinationPort)
		}

		var matches []string
		for _, name := range cfg.Priority {
			collection, exists := byName[name]
			if !exists {
				return fmt.Errorf("priority references unbuilt ruleset %q", name)
			}
			if collection.Matches(a.Domain, addr, a.DestinationPort) {
				matches = append(matches, name)
			}
		}
		actual := cfg.Fallback
		if len(matches) != 0 {
			actual = matches[0]
		}
		if actual != a.RuleSet {
			return fmt.Errorf("%s assertion %q: classified as %q, expected %q (matching sets: %s)", kind, target, actual, a.RuleSet, strings.Join(matches, ", "))
		}
		if a.Exclusive && len(matches) != 1 {
			return fmt.Errorf("%s assertion %q: expected exclusive membership in %q (matching sets: %s)", kind, target, a.RuleSet, strings.Join(matches, ", "))
		}
	}
	return nil
}

func counts(collection rules.Collection) Counts {
	return Counts{
		Domain:    len(collection.Domains),
		Classical: len(collection.Classical),
		IPCIDR:    len(collection.Prefixes),
	}
}
