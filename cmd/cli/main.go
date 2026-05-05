package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"tiagorocha94/household-finance-pipeline/internal/aggregation/household"
	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/ingestion/filesystem"
	"tiagorocha94/household-finance-pipeline/internal/parsing"
	"tiagorocha94/household-finance-pipeline/internal/parsing/csv"
	"tiagorocha94/household-finance-pipeline/internal/parsing/xlsx"
	"tiagorocha94/household-finance-pipeline/internal/pipeline"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
	csvexp "tiagorocha94/household-finance-pipeline/internal/presentation/csv"
	jsonexp "tiagorocha94/household-finance-pipeline/internal/presentation/json"
	"tiagorocha94/household-finance-pipeline/internal/presentation/markdown"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	if err := run(logger); err != nil {
		logger.Error("error running application", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	defaultMonth := time.Now().Format("2006-01")
	month := flag.String("month", defaultMonth, "month to process in YYYY-MM format (defaults to current month)")
	flag.Parse()

	cfg, err := config.Load("config.json")
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	p, err := pipeline.New(
		cfg,
		filesystem.New(),
		[]parsing.Parser{
			csv.New(),
			xlsx.New(),
		},
		household.New(cfg),
		[]presentation.Exporter{
			markdown.New(),
			jsonexp.New(),
			csvexp.New(),
		},
		logger,
	)
	if err != nil {
		return fmt.Errorf("init error: %w", err)
	}

	if err := p.Run(*month); err != nil {
		return fmt.Errorf("run error: %w", err)
	}
	return nil
}
