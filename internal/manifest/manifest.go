package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Manifest is the top-level structure written to manifest.json.
type Manifest struct {
	Months []string `json:"months"` // sorted newest-first, e.g. ["2026-02", "2026-01"]
}

// Update scans outputDir for subdirectories containing a data.json file and
// rewrites <outputDir>/manifest.json with the sorted list of months.
func Update(outputDir string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("read output dir: %w", err)
	}

	var months []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Only include dirs that actually have a data.json so the SPA never
		// fetches a month with no data.
		dataPath := filepath.Join(outputDir, e.Name(), "data.json")
		if _, err := os.Stat(dataPath); err != nil {
			continue
		}
		// Expect YYYY-MM format — skip anything that doesn't look like it.
		if !isMonthDir(e.Name()) {
			continue
		}
		months = append(months, e.Name())
	}

	// Sort newest-first.
	sort.Sort(sort.Reverse(sort.StringSlice(months)))

	m := Manifest{Months: months}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	dest := filepath.Join(outputDir, "manifest.json")
	if err := os.WriteFile(dest, b, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	return nil
}

// isMonthDir returns true when s matches YYYY-MM.
func isMonthDir(s string) bool {
	parts := strings.Split(s, "-")
	if len(parts) != 2 || len(parts[0]) != 4 || len(parts[1]) != 2 {
		return false
	}
	for _, c := range parts[0] + parts[1] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}