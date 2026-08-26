package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DatabasePath string
	InputPath    string
	ReportPath   string
	Actor        string
	Format       string
	Quiet        bool
}

func Default() Config {
	return Config{DatabasePath: "sitepay.db", Actor: "project-manager", Format: "csv"}
}

func FromArgs(args []string) (Config, error) {
	cfg := Default()
	flags := flag.NewFlagSet("sitepay", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.StringVar(&cfg.DatabasePath, "db", cfg.DatabasePath, "sqlite database path")
	flags.StringVar(&cfg.InputPath, "input", "", "CSV input path")
	flags.StringVar(&cfg.ReportPath, "report", "", "report output path")
	flags.StringVar(&cfg.Actor, "actor", cfg.Actor, "operator name")
	flags.StringVar(&cfg.Format, "format", cfg.Format, "report format: csv or json")
	flags.BoolVar(&cfg.Quiet, "quiet", false, "suppress progress output")
	if err := flags.Parse(args); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(cfg.DatabasePath) == "" {
		return Config{}, fmt.Errorf("database path cannot be empty")
	}
	if cfg.Format != "csv" && cfg.Format != "json" {
		return Config{}, fmt.Errorf("format must be csv or json")
	}
	return cfg, nil
}

func (c Config) EnsureParent() error {
	parent := filepath.Dir(c.DatabasePath)
	if parent == "." {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}

func (c Config) IsBatchMode() bool { return strings.TrimSpace(c.InputPath) != "" }

func (c Config) String() string {
	return fmt.Sprintf("database=%s input=%s report=%s actor=%s format=%s", c.DatabasePath, c.InputPath, c.ReportPath, c.Actor, c.Format)
}
