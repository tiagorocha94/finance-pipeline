package pipeline

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"tiagorocha94/household-finance-pipeline/internal/aggregation"
	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/ingestion"
	"tiagorocha94/household-finance-pipeline/internal/manifest"
	"tiagorocha94/household-finance-pipeline/internal/parsing"
	"tiagorocha94/household-finance-pipeline/internal/parsing/router"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

// Pipeline orchestrates the full workflow:
// ingestion → parsing → aggregation → export
type Pipeline struct {
	cfg config.Config
	log *slog.Logger

	ingestor   ingestion.Ingestor
	parser     parsing.Parser
	aggregator aggregation.Aggregator
	exporters  []presentation.Exporter
}

// New creates a fully configured pipeline.
func New(
	cfg config.Config,
	ingestor ingestion.Ingestor,
	parsers []parsing.Parser,
	aggregator aggregation.Aggregator,
	exporters []presentation.Exporter,
	logger *slog.Logger,
) (*Pipeline, error) {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	if ingestor == nil {
		return nil, fmt.Errorf("ingestor is required")
	}
	if len(parsers) == 0 {
		return nil, fmt.Errorf("at least one parser is required")
	}
	if aggregator == nil {
		return nil, fmt.Errorf("aggregator is required")
	}
	if len(exporters) == 0 {
		return nil, fmt.Errorf("at least one exporter is required")
	}

	p := &Pipeline{
		cfg:        cfg,
		log:        logger,
		ingestor:   ingestor,
		parser:     router.New(parsers),
		aggregator: aggregator,
		exporters:  exporters,
	}

	p.log.Info("pipeline initialized")
	return p, nil
}

// Run executes the pipeline for a given month.
func (p *Pipeline) Run(month string) error {
	p.log.Info("pipeline started", "month", month)

	rawFiles, err := p.ingestor.Ingest(p.cfg, month)
	if err != nil {
		return err
	}
	p.log.Info("files ingested", "count", len(rawFiles))

	var parsed []domain.PersonData
	for _, file := range rawFiles {
		p.log.Debug("parsing file", "file", file.Name)
		data, err := p.parser.Parse(file)
		if err != nil {
			return fmt.Errorf("parse %s: %w", file.Name, err)
		}
		parsed = append(parsed, data)
	}
	p.log.Info("parsing completed", "persons", len(parsed))

	household, peopleData, err := p.aggregator.Aggregate(parsed)
	if err != nil {
		return err
	}
	p.log.Info("aggregation completed")

	report := presentation.BuildReport(household, peopleData)

	for _, exporter := range p.exporters {
		name := fmt.Sprintf("%T", exporter)
		p.log.Info("exporting report", "exporter", name)

		outputs, err := exporter.Export(report)
		if err != nil {
			return err
		}
		if err := p.writeOutputs(month, outputs); err != nil {
			return err
		}
	}

	if err := manifest.Update(p.cfg.OutputDir); err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	p.log.Info("pipeline finished successfully")
	return nil
}

func (p *Pipeline) writeOutputs(month string, outputs []presentation.Output) error {
	outputDir := filepath.Join(p.cfg.OutputDir, month)

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	for _, o := range outputs {
		path := filepath.Join(outputDir, o.File.Name)
		if err := os.WriteFile(path, o.File.Content, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		p.log.Info("file written", "file", path, "bytes", len(o.File.Content))
	}

	return nil
}
