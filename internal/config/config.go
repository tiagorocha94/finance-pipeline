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

	// People lists the expected person names; input filenames must match these.
	People []string `json:"people"`

	// CategoryGroups defines the budget groups, their target percentages of total
	// income, and the expense categories that belong to each group.
	// If empty, budget group calculations are skipped.
	CategoryGroups []CategoryGroup `json:"category_groups"`
}

// Load reads and parses a JSON config file.
func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
