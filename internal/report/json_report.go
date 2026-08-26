package report

import (
	"encoding/json"
	"fmt"
	"io"

	"sitepay/internal/domain"
)

type JSONReport struct {
	Statements []JSONStatement `json:"statements"`
	Totals     JSONTotals      `json:"totals"`
}

type JSONStatement struct {
	StatementNo string `json:"statement_no"`
	WorkerName  string `json:"worker_name"`
	Trade       string `json:"trade"`
	WorkDate    string `json:"work_date"`
	Gross       string `json:"gross"`
	Allowance   string `json:"night_allowance"`
	Deduction   string `json:"quality_deduction"`
	Net         string `json:"net"`
	Status      string `json:"status"`
}

type JSONTotals struct {
	Statements int    `json:"statements"`
	Gross      string `json:"gross"`
	Allowance  string `json:"night_allowance"`
	Deduction  string `json:"quality_deduction"`
	Net        string `json:"net"`
	Reviews    int    `json:"reviews"`
}

func WriteJSON(writer io.Writer, statements []domain.PayrollStatement, totals domain.ReportTotals) error {
	if writer == nil {
		return fmt.Errorf("report writer is nil")
	}
	report := JSONReport{Statements: make([]JSONStatement, 0, len(statements)), Totals: JSONTotals{Statements: totals.Statements, Gross: domain.FormatCents(totals.GrossCents), Allowance: domain.FormatCents(totals.NightCents), Deduction: domain.FormatCents(totals.DeductCents), Net: domain.FormatCents(totals.NetCents), Reviews: totals.Reviews}}
	for _, statement := range statements {
		report.Statements = append(report.Statements, JSONStatement{StatementNo: statement.StatementNo, WorkerName: statement.WorkerName, Trade: statement.Trade, WorkDate: statement.WorkDate, Gross: domain.FormatCents(statement.GrossCents), Allowance: domain.FormatCents(statement.AllowanceCents), Deduction: domain.FormatCents(statement.DeductionCents), Net: domain.FormatCents(statement.NetCents), Status: statement.Status})
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func DecodeJSON(data []byte) (JSONReport, error) {
	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		return JSONReport{}, err
	}
	return report, nil
}
