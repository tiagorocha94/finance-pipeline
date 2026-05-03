package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
)

// Ingestor reads files from a local directory structure.
type Ingestor struct{}

func New() *Ingestor {
	return &Ingestor{}
}

func (i *Ingestor) Ingest(cfg config.Config, month string) ([]domain.RawFile, error) {
	basePath := filepath.Join(cfg.DataDir, month)

	var files []domain.RawFile
	for _, person := range cfg.People {
		file, err := i.loadPersonFile(basePath, person)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (i *Ingestor) loadPersonFile(basePath, person string) (domain.RawFile, error) {
	fullPath, err := findFile(basePath, person)
	if err != nil {
		return domain.RawFile{}, err
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return domain.RawFile{}, fmt.Errorf("reading %s: %w", fullPath, err)
	}

	return domain.RawFile{
		Name:    filepath.Base(fullPath),
		Content: content,
	}, nil
}

func findFile(basePath, person string) (string, error) {
	matches, err := filepath.Glob(filepath.Join(basePath, person+".*"))
	if err != nil {
		return "", err
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no file found for %s", person)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple files found for %s", person)
	}
}
