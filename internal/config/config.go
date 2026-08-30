package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Repository    string      `json:"repository"`
	Fallback      string      `json:"fallback"`
	Priority      []string    `json:"priority"`
	V2FlyURL      string      `json:"v2fly_url"`
	IgnoreDomains []string    `json:"ignore_domains"`
	RuleSets      []RuleSet   `json:"rulesets"`
	Assertions    []Assertion `json:"assertions"`
	BaseDir       string      `json:"-"`
}

type RuleSet struct {
	Name    string   `json:"name"`
	Exclude []string `json:"exclude"`
	Sources []Source `json:"sources"`
}

type Source struct {
	Name              string   `json:"name"`
	Kind              string   `json:"kind"`
	URL               string   `json:"url"`
	URLs              []string `json:"urls"`
	Path              string   `json:"path"`
	Lists             []string `json:"lists"`
	AllLists          bool     `json:"all_lists"`
	RequireAttributes []string `json:"require_attributes"`
	OptionalIfCovered bool     `json:"optional_if_covered"`
}

type Assertion struct {
	Domain  string `json:"domain"`
	RuleSet string `json:"ruleset"`
}

func Load(path string) (Config, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Config{}, fmt.Errorf("decode config: trailing content")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return Config{}, fmt.Errorf("resolve config path: %w", err)
	}
	cfg.BaseDir = filepath.Dir(abs)
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.Repository == "" || c.Fallback == "" || c.V2FlyURL == "" {
		return fmt.Errorf("repository, fallback, and v2fly_url are required")
	}
	sets := make(map[string]int, len(c.RuleSets))
	for index, ruleset := range c.RuleSets {
		if ruleset.Name == "" || len(ruleset.Sources) == 0 {
			return fmt.Errorf("ruleset %d requires a name and sources", index)
		}
		if _, exists := sets[ruleset.Name]; exists {
			return fmt.Errorf("duplicate ruleset %q", ruleset.Name)
		}
		sets[ruleset.Name] = index
		for _, excluded := range ruleset.Exclude {
			excludedIndex, exists := sets[excluded]
			if !exists || excludedIndex >= index {
				return fmt.Errorf("ruleset %q excludes non-prior ruleset %q", ruleset.Name, excluded)
			}
		}
		sourceNames := make(map[string]struct{}, len(ruleset.Sources))
		for _, source := range ruleset.Sources {
			if source.Name == "" || source.Kind == "" {
				return fmt.Errorf("ruleset %q contains an unnamed source", ruleset.Name)
			}
			if _, exists := sourceNames[source.Name]; exists {
				return fmt.Errorf("ruleset %q contains duplicate source %q", ruleset.Name, source.Name)
			}
			sourceNames[source.Name] = struct{}{}
			if err := validateSource(source); err != nil {
				return fmt.Errorf("ruleset %q source %q: %w", ruleset.Name, source.Name, err)
			}
		}
	}
	if len(c.Priority) != len(c.RuleSets) {
		return fmt.Errorf("priority must name every ruleset exactly once")
	}
	for index, name := range c.Priority {
		if c.RuleSets[index].Name != name {
			return fmt.Errorf("priority and ruleset order differ at index %d", index)
		}
	}
	if _, exists := sets[c.Fallback]; !exists {
		return fmt.Errorf("fallback %q is not a ruleset", c.Fallback)
	}
	for _, assertion := range c.Assertions {
		if assertion.Domain == "" {
			return fmt.Errorf("assertion domain is required")
		}
		if _, exists := sets[assertion.RuleSet]; !exists {
			return fmt.Errorf("assertion for %q has unknown ruleset %q", assertion.Domain, assertion.RuleSet)
		}
	}
	return nil
}

func validateSource(source Source) error {
	hasURL := source.URL != ""
	hasURLs := len(source.URLs) != 0
	hasPath := source.Path != ""
	hasSelection := len(source.Lists) != 0 || source.AllLists
	switch source.Kind {
	case "clash", "dnsmasq", "cidr":
		if !hasURL || hasURLs || hasPath || hasSelection {
			return fmt.Errorf("kind %q requires exactly one url", source.Kind)
		}
	case "v2fly":
		if hasURL || hasURLs || hasPath || len(source.Lists) == 0 && !source.AllLists {
			return fmt.Errorf("kind v2fly requires lists or all_lists")
		}
		if source.AllLists && len(source.Lists) != 0 {
			return fmt.Errorf("kind v2fly cannot combine lists and all_lists")
		}
	case "iana-local":
		if hasURL || !hasURLs || hasPath || hasSelection {
			return fmt.Errorf("kind iana-local requires urls")
		}
	case "domain-file":
		if hasURL || hasURLs || !hasPath || hasSelection {
			return fmt.Errorf("kind domain-file requires a path")
		}
		clean := filepath.Clean(source.Path)
		if filepath.IsAbs(source.Path) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("path must stay below the configuration directory")
		}
	default:
		return fmt.Errorf("unknown kind %q", source.Kind)
	}
	return nil
}
