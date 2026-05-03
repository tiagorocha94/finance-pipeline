package router

import (
	"fmt"
	"path/filepath"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/parsing"
)

// Router dispatches RawFiles to the correct Parser based on file extension.
// It is built from a list of parsers that each advertise their extension via
// the Extension() string method.
type Router struct {
	parsers map[string]parsing.Parser
}

// New builds a Router from a list of parsers. Each parser must implement
// interface{ Extension() string } to register itself for a file extension.
func New(parsers []parsing.Parser) *Router {
	m := make(map[string]parsing.Parser, len(parsers))
	for _, p := range parsers {
		if ep, ok := p.(interface{ Extension() string }); ok {
			m[ep.Extension()] = p
		}
	}
	return &Router{parsers: m}
}

func (r *Router) Parse(file domain.RawFile) (domain.PersonData, error) {
	ext := filepath.Ext(file.Name)
	parser, ok := r.parsers[ext]
	if !ok {
		return domain.PersonData{}, fmt.Errorf("no parser for extension %q", ext)
	}
	return parser.Parse(file)
}
