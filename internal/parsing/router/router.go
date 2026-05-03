package router

import (
	"fmt"
	"path/filepath"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/parsing"
)

// Router dispatches files to correct parser based on extension.
type Router struct {
	parsers map[string]parsing.Parser
}

// New builds router from list of parsers.
func New(parsers []parsing.Parser) *Router {
	m := make(map[string]parsing.Parser)

	for _, p := range parsers {
		// We assume parser is CSV/JSON-specific and inferred externally
		// via type or registration in main
		switch t := p.(type) {
		case interface{ Extension() string }:
			m[t.Extension()] = p
		}
	}

	return &Router{parsers: m}
}

func (r *Router) Parse(file domain.RawFile) (domain.PersonData, error) {
	ext := filepath.Ext(file.Name)

	parser, ok := r.parsers[ext]
	if !ok {
		return domain.PersonData{}, fmt.Errorf("no parser for extension: %s", ext)
	}

	return parser.Parse(file)
}
