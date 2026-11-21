package build

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Repository    string          `json:"repository"`
	Fallback      string          `json:"fallback"`
	Priority      []string        `json:"priority"`
	V2FlyURL      string          `json:"v2fly_url"`
	IgnoreDomains []string        `json:"ignore_domains"`
	RuleSets      []RuleSetConfig `json:"rulesets"`
	Assertions    []Assertion     `json:"assertions"`
	BaseDir       string          `json:"-"`
}

type RuleSetConfig struct {
	Name    string         `json:"name"`
	Exclude []string       `json:"exclude"`
	Sources []SourceConfig `json:"sources"`
}

type SourceConfig struct {
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
	Domain          string `json:"domain"`
	IP              string `json:"ip"`
	DestinationPort uint16 `json:"destination_port,omitempty"`
	RuleSet         string `json:"ruleset"`
	Exclusive       bool   `json:"exclusive"`
}

func LoadConfig(path string) (Config, error) {
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
	sets := make(map[string]struct{}, len(c.RuleSets))
	for i, ruleset := range c.RuleSets {
		if ruleset.Name == "" || len(ruleset.Sources) == 0 {
			return fmt.Errorf("ruleset %d requires a name and sources", i)
		}
		if _, exists := sets[ruleset.Name]; exists {
			return fmt.Errorf("duplicate ruleset %q", ruleset.Name)
		}
		for _, excluded := range ruleset.Exclude {
			if _, exists := sets[excluded]; !exists {
				return fmt.Errorf("ruleset %q excludes non-prior ruleset %q", ruleset.Name, excluded)
			}
		}
		sets[ruleset.Name] = struct{}{}
		sourceNames := make(map[string]struct{}, len(ruleset.Sources))
		for _, src := range ruleset.Sources {
			if src.Name == "" || src.Kind == "" {
				return fmt.Errorf("ruleset %q contains an unnamed source", ruleset.Name)
			}
			if _, exists := sourceNames[src.Name]; exists {
				return fmt.Errorf("ruleset %q contains duplicate source %q", ruleset.Name, src.Name)
			}
			sourceNames[src.Name] = struct{}{}
			if err := validateSource(src); err != nil {
				return fmt.Errorf("ruleset %q source %q: %w", ruleset.Name, src.Name, err)
			}
		}
	}
	if len(c.Priority) != len(c.RuleSets) {
		return fmt.Errorf("priority must name every ruleset exactly once")
	}
	for i, name := range c.Priority {
		if c.RuleSets[i].Name != name {
			return fmt.Errorf("priority and ruleset order differ at index %d", i)
		}
	}
	if _, exists := sets[c.Fallback]; !exists {
		return fmt.Errorf("fallback %q is not a ruleset", c.Fallback)
	}
	for _, a := range c.Assertions {
		if (a.Domain == "") == (a.IP == "") {
			return fmt.Errorf("assertion requires exactly one of domain or ip")
		}
		target := a.Domain
		if a.IP != "" {
			target = a.IP
			if _, err := netip.ParseAddr(a.IP); err != nil {
				return fmt.Errorf("assertion IP %q: %w", a.IP, err)
			}
		}
		if _, exists := sets[a.RuleSet]; !exists {
			return fmt.Errorf("assertion for %q has unknown ruleset %q", target, a.RuleSet)
		}
	}
	return nil
}

func validateSource(src SourceConfig) error {
	hasURL, hasURLs, hasPath := src.URL != "", len(src.URLs) != 0, src.Path != ""
	hasLists, hasAttrs := len(src.Lists) != 0, len(src.RequireAttributes) != 0
	hasSelection := hasLists || src.AllLists
	switch src.Kind {
	case "clash", "dnsmasq", "cidr":
		if !hasURL || hasURLs || hasPath || hasSelection || hasAttrs {
			return fmt.Errorf("kind %q requires exactly one url", src.Kind)
		}
	case "v2fly":
		if hasURL || hasURLs || hasPath || !hasSelection {
			return fmt.Errorf("kind v2fly requires lists or all_lists")
		}
		if src.AllLists && hasLists {
			return fmt.Errorf("kind v2fly cannot combine lists and all_lists")
		}
		for _, attr := range src.RequireAttributes {
			if attr == "" || strings.TrimSpace(attr) != attr || strings.ContainsAny(attr, ",@:") {
				return fmt.Errorf("kind v2fly has invalid required attribute %q", attr)
			}
		}
	case "iana-local":
		if hasURL || !hasURLs || hasPath || hasSelection || hasAttrs {
			return fmt.Errorf("kind iana-local requires urls")
		}
	case "domain-file", "cidr-file", "clash-file":
		if hasURL || hasURLs || !hasPath || hasSelection || hasAttrs {
			return fmt.Errorf("kind %s requires a path", src.Kind)
		}
		if !filepath.IsLocal(src.Path) {
			return fmt.Errorf("path must stay below the configuration directory")
		}
	default:
		return fmt.Errorf("unknown kind %q", src.Kind)
	}
	return nil
}
