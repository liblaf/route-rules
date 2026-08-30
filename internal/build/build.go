package build

import (
	"context"
	"fmt"
	"strings"

	"github.com/liblaf/route-rules/internal/config"
	"github.com/liblaf/route-rules/internal/rules"
	"github.com/liblaf/route-rules/internal/source"
)

type Counts struct {
	Domain    int `json:"domain"`
	Classical int `json:"classical"`
	IPCIDR    int `json:"ipcidr"`
}

type SourceReport struct {
	Name             string                 `json:"name"`
	Kind             string                 `json:"kind"`
	InputRules       int                    `json:"input_rules"`
	Accepted         Counts                 `json:"accepted"`
	Ignored          map[string]int         `json:"ignored,omitempty"`
	Fetches          []source.FetchMetadata `json:"fetches,omitempty"`
	OmittedAsCovered int                    `json:"omitted_as_covered,omitempty"`
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
	byName   map[string]rules.Collection
}

func Run(ctx context.Context, cfg config.Config) (Result, error) {
	loader := source.NewLoader(cfg.V2FlyURL)
	result := Result{byName: make(map[string]rules.Collection, len(cfg.RuleSets))}
	ignored := make(map[string]struct{}, len(cfg.IgnoreDomains))
	for _, domain := range cfg.IgnoreDomains {
		ignored[strings.ToLower(strings.TrimSuffix(domain, "."))] = struct{}{}
	}

	for _, configured := range cfg.RuleSets {
		built := RuleSet{Name: configured.Name, Excluded: append([]string(nil), configured.Exclude...)}
		var aggregate rules.Collection
		for _, configuredSource := range configured.Sources {
			loaded, err := loader.Load(ctx, cfg.BaseDir, configuredSource)
			if err != nil {
				return Result{}, fmt.Errorf("ruleset %q: %w", configured.Name, err)
			}
			loaded.Rules = removeIgnoredDomains(loaded.Rules, ignored)
			beforeOptional := counts(loaded.Rules)
			if configuredSource.OptionalIfCovered {
				loaded.Rules = loaded.Rules.Excluding(aggregate)
			}
			afterOptional := counts(loaded.Rules)
			built.Sources = append(built.Sources, SourceReport{
				Name:             configuredSource.Name,
				Kind:             configuredSource.Kind,
				InputRules:       loaded.InputRules,
				Accepted:         afterOptional,
				Ignored:          loaded.Ignored,
				Fetches:          loaded.Fetches,
				OmittedAsCovered: total(beforeOptional) - total(afterOptional),
			})
			aggregate.Append(loaded.Rules)
		}

		built.BeforeOptimization = counts(aggregate)
		aggregate = aggregate.Optimized()
		built.AfterOptimization = counts(aggregate)
		for _, excludedName := range configured.Exclude {
			excluded, exists := result.byName[excludedName]
			if !exists {
				panic("validated exclusion was not built")
			}
			aggregate = aggregate.Excluding(excluded)
		}
		built.Rules = aggregate
		built.Final = counts(aggregate)
		result.RuleSets = append(result.RuleSets, built)
		result.byName[configured.Name] = aggregate
	}

	if err := verifyAssertions(cfg, result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (result Result) Rules(name string) (rules.Collection, bool) {
	collection, exists := result.byName[name]
	return collection, exists
}

func verifyAssertions(cfg config.Config, built Result) error {
	for _, assertion := range cfg.Assertions {
		actual := cfg.Fallback
		for _, name := range cfg.Priority {
			collection, exists := built.byName[name]
			if !exists {
				panic("priority references an unbuilt ruleset")
			}
			if collection.MatchesDomain(assertion.Domain) {
				actual = name
				break
			}
		}
		if actual != assertion.RuleSet {
			matches := make([]string, 0)
			for _, name := range cfg.Priority {
				if built.byName[name].MatchesDomain(assertion.Domain) {
					matches = append(matches, name)
				}
			}
			return fmt.Errorf("domain assertion %q: classified as %q, expected %q (matching sets: %s)", assertion.Domain, actual, assertion.RuleSet, strings.Join(matches, ", "))
		}
	}
	return nil
}

func removeIgnoredDomains(collection rules.Collection, ignored map[string]struct{}) rules.Collection {
	filtered := make(rules.DomainSet, 0, len(collection.Domains))
	for _, rule := range collection.Domains {
		if _, exists := ignored[rule.Value]; !exists {
			filtered = append(filtered, rule)
		}
	}
	collection.Domains = filtered
	return collection
}

func counts(collection rules.Collection) Counts {
	return Counts{
		Domain:    len(collection.Domains),
		Classical: len(collection.Classical),
		IPCIDR:    len(collection.Prefixes),
	}
}

func total(counts Counts) int {
	return counts.Domain + counts.Classical + counts.IPCIDR
}
