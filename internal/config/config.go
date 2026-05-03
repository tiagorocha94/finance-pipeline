package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CategoryGroup defines a budget group, its target allocation, and the expense
// category names that belong to it.
type CategoryGroup struct {
	Name       string   `json:"name"`
	Percent    string   `json:"percent"`    // e.g. "50%"
	Categories []string `json:"categories"` // expense category names mapped to this group
}

// ParsedPercent returns the percent value as a float64 fraction (e.g. 0.50).
func (g CategoryGroup) ParsedPercent() (float64, error) {
	s := strings.TrimSuffix(strings.TrimSpace(g.Percent), "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("category group %q: invalid percent %q", g.Name, g.Percent)
	}
	return v / 100, nil
}

// Config is the top-level pipeline configuration.
type Config struct {
	// DataDir is the root directory from which input files are read.
	DataDir string `json:"data_dir"`

	// OutputDir is the directory where reports and the manifest are written.
	// Defaults to "output" if not set.
	OutputDir string `json:"output_dir"`

	// People lists the expected person names; input filenames must match these.
	People []string `json:"people"`

	// CategoryGroups defines the budget groups, their target percentages of total
	// income, and the expense categories that belong to each group.
	// If empty, budget group calculations are skipped.
	CategoryGroups []CategoryGroup `json:"category_groups"`
}

// Load reads, parses, and validates a JSON config file.
func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	if err := cfg.setDefaults().validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) setDefaults() *Config {
	if c.DataDir == "" {
		c.DataDir = "data"
	}
	if c.OutputDir == "" {
		c.OutputDir = "output"
	}
	return c
}

// validate returns an error for any configuration that would cause a silent
// failure downstream.
func (c *Config) validate() error {
	if len(c.People) == 0 {
		return fmt.Errorf("config: people list is empty")
	}

	seen := make(map[string]bool, len(c.People))
	for _, p := range c.People {
		if p == "" {
			return fmt.Errorf("config: people list contains an empty name")
		}
		if seen[p] {
			return fmt.Errorf("config: duplicate person name %q", p)
		}
		seen[p] = true
	}

	return c.validateCategoryGroups()
}

func (c *Config) validateCategoryGroups() error {
	if len(c.CategoryGroups) == 0 {
		return nil
	}

	seenGroups := make(map[string]bool, len(c.CategoryGroups))
	catchAlls := 0

	for _, g := range c.CategoryGroups {
		if g.Name == "" {
			return fmt.Errorf("config: category group has an empty name")
		}
		if seenGroups[g.Name] {
			return fmt.Errorf("config: duplicate category group name %q", g.Name)
		}
		seenGroups[g.Name] = true

		if _, err := g.ParsedPercent(); err != nil {
			return err
		}

		for _, cat := range g.Categories {
			if cat == "*" {
				catchAlls++
			}
		}
	}

	if catchAlls > 1 {
		return fmt.Errorf("config: only one category group may use \"*\" as a catch-all; found %d", catchAlls)
	}

	return nil
}
