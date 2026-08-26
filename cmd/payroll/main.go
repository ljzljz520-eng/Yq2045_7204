package main

import (
	"context"
	"fmt"
	"os"

	"sitepay/internal/config"
	"sitepay/internal/domain"
	"sitepay/internal/payroll"
	"sitepay/internal/service"
	"sitepay/internal/store"
)

func main() {
	cfg, err := config.FromArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := cfg.EnsureParent(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	db, err := store.Open(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	ctx := context.Background()
	payrollService, err := service.NewPayrollService(db, loadPolicies(ctx, db))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := payrollService.SeedPolicies(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !cfg.IsBatchMode() {
		if !cfg.Quiet {
			fmt.Println("sitepay工地计件工资核算器 ready")
			fmt.Println("use -input file.csv -report report.csv to process a batch")
		}
		return
	}
	input, err := os.Open(cfg.InputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer input.Close()
	batch := service.NewBatchService(payrollService)
	summary, issues, err := batch.Process(ctx, input, cfg.Actor, "cli-batch")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if !issues.Empty() && !cfg.Quiet {
		fmt.Fprintln(os.Stderr, issues.String())
	}
	if cfg.ReportPath != "" {
		output, err := os.Create(cfg.ReportPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		reportService := service.NewReportService(payrollService)
		_, exportErr := reportService.Export(ctx, output, cfg.Format, "")
		closeErr := output.Close()
		if exportErr != nil {
			fmt.Fprintln(os.Stderr, exportErr)
			os.Exit(1)
		}
		if closeErr != nil {
			fmt.Fprintln(os.Stderr, closeErr)
			os.Exit(1)
		}
	}
	if !cfg.Quiet {
		fmt.Printf("batch=%s imported=%d rejected=%d reviews=%d net=%s\n", summary.BatchID, summary.Imported, summary.Rejected, summary.ReviewCount, formatCents(summary.TotalNetCents))
	}
}

func loadPolicies(ctx context.Context, db *store.Store) []domain.AllowancePolicy {
	policies, err := db.ListPolicies(ctx)
	if err == nil && len(policies) > 0 {
		return policies
	}
	return payroll.DefaultPolicies()
}

func formatCents(value int64) string { return domain.FormatCents(value) }
