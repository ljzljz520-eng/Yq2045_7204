package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"sitepay/internal/domain"
)

type CSVImporter struct {
	Delimiter rune
	Strict    bool
}

func NewCSVImporter() CSVImporter { return CSVImporter{Delimiter: ',', Strict: false} }

func (i CSVImporter) Read(reader io.Reader) (domain.ImportResult, error) {
	if reader == nil {
		return domain.ImportResult{}, fmt.Errorf("input reader is nil")
	}
	csvReader := csv.NewReader(reader)
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1
	if i.Delimiter != 0 {
		csvReader.Comma = i.Delimiter
	}
	result := domain.ImportResult{}
	lineNumber := 0
	first := true
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		lineNumber++
		if err != nil {
			result.Issues = append(result.Issues, domain.ImportIssue{Line: lineNumber, Message: err.Error()})
			continue
		}
		if first && looksLikeHeader(record) {
			first = false
			continue
		}
		first = false
		result.Total++
		entry, issues := parseRecord(record, lineNumber)
		if len(issues) > 0 {
			result.Issues = append(result.Issues, issues...)
			if i.Strict {
				return result, fmt.Errorf("strict import rejected line %d", lineNumber)
			}
			continue
		}
		result.Entries = append(result.Entries, entry)
	}
	return NormalizeResult(result), nil
}

func looksLikeHeader(record []string) bool {
	if len(record) == 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(record[0]), "worker_name") || strings.EqualFold(strings.TrimSpace(record[0]), "name")
}

func parseRecord(record []string, line int) (domain.WorkEntry, []domain.ImportIssue) {
	entry := domain.WorkEntry{ImportedLine: line}
	issues := make([]domain.ImportIssue, 0)
	fields := []struct {
		index string
		set   func(string) error
	}{
		{"worker_name", func(value string) error { entry.WorkerName = strings.TrimSpace(value); return nil }},
		{"trade", func(value string) error { entry.Trade = strings.TrimSpace(value); return nil }},
		{"completed_pieces", func(value string) error {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			entry.CompletedPieces = parsed
			return err
		}},
		{"unit_price", func(value string) error {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			entry.UnitPrice = parsed
			return err
		}},
		{"night_allowance", func(value string) error {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			entry.NightAllowance = parsed
			return err
		}},
		{"quality_deduction", func(value string) error {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			entry.QualityDeduction = parsed
			return err
		}},
		{"work_date", func(value string) error { entry.WorkDate = strings.TrimSpace(value); return nil }},
	}
	for index, field := range fields {
		if index >= len(record) {
			issues = append(issues, domain.ImportIssue{Line: line, Column: field.index, Message: "missing value"})
			continue
		}
		if err := field.set(record[index]); err != nil {
			issues = append(issues, domain.ImportIssue{Line: line, Column: field.index, Value: record[index], Message: "invalid number"})
		}
	}
	if err := entry.Validate(); err != nil {
		issues = append(issues, domain.ImportIssue{Line: line, Message: err.Error()})
	}
	return entry, issues
}

func EncodeHeader() []string {
	return []string{"worker_name", "trade", "completed_pieces", "unit_price", "night_allowance", "quality_deduction", "work_date"}
}

func EncodeEntry(entry domain.WorkEntry) []string {
	return []string{entry.WorkerName, entry.Trade, strconv.Itoa(entry.CompletedPieces), formatFloat(entry.UnitPrice), formatFloat(entry.NightAllowance), formatFloat(entry.QualityDeduction), entry.WorkDate}
}

func formatFloat(value float64) string { return strconv.FormatFloat(value, 'f', 2, 64) }
