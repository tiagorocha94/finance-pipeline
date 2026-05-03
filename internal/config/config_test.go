package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"tiagorocha94/household-finance-pipeline/internal/config"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "config-*.json")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidMinimal(t *testing.T) {
	// Neither data_dir nor output_dir are required — both have defaults.
	path := writeConfig(t, `{"people":["Alice"]}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDir != "data" {
		t.Errorf("default DataDir: want data, got %q", cfg.DataDir)
	}
	if cfg.OutputDir != "output" {
		t.Errorf("default OutputDir: want output, got %q", cfg.OutputDir)
	}
}

func TestLoad_DataDirDefault(t *testing.T) {
	path := writeConfig(t, `{"people":["Alice"]}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DataDir != "data" {
		t.Errorf("default DataDir: want data, got %q", cfg.DataDir)
	}
}

func TestLoad_OutputDirDefault(t *testing.T) {
	path := writeConfig(t, `{"people":["Alice"]}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != "output" {
		t.Errorf("want output, got %q", cfg.OutputDir)
	}
}

func TestLoad_OutputDirExplicit(t *testing.T) {
	path := writeConfig(t, `{"data_dir":"./data","output_dir":"reports","people":["Alice"]}`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputDir != "reports" {
		t.Errorf("want reports, got %q", cfg.OutputDir)
	}
}

func TestLoad_EmptyPeople(t *testing.T) {
	path := writeConfig(t, `{"data_dir":"./data","people":[]}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for empty people list")
	}
}

func TestLoad_DuplicatePerson(t *testing.T) {
	path := writeConfig(t, `{"data_dir":"./data","people":["Alice","Alice"]}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for duplicate person name")
	}
}

func TestLoad_DuplicateCategoryGroup(t *testing.T) {
	path := writeConfig(t, `{
		"data_dir":"./data",
		"people":["Alice"],
		"category_groups":[
			{"name":"Essentials","percent":"50%","categories":["Food"]},
			{"name":"Essentials","percent":"20%","categories":["Rent"]}
		]
	}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for duplicate category group name")
	}
}

func TestLoad_MultipleCatchAlls(t *testing.T) {
	path := writeConfig(t, `{
		"data_dir":"./data",
		"people":["Alice"],
		"category_groups":[
			{"name":"A","percent":"50%","categories":["*"]},
			{"name":"B","percent":"50%","categories":["*"]}
		]
	}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for multiple catch-all groups")
	}
}

func TestLoad_InvalidPercent(t *testing.T) {
	path := writeConfig(t, `{
		"data_dir":"./data",
		"people":["Alice"],
		"category_groups":[
			{"name":"A","percent":"not-a-percent","categories":["Food"]}
		]
	}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid percent")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := writeConfig(t, `{not valid json}`)
	_, err := config.Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
