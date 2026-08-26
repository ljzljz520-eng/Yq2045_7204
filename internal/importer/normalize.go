package importer

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"sitepay/internal/domain"
)

type Normalizer struct {
	Aliases map[string]string
}

func NewNormalizer() Normalizer {
	return Normalizer{Aliases: map[string]string{
		"bricklayer": "mason",
		"woodworker": "carpenter",
		"ironworker": "steelworker",
		"wireman":    "electrician",
		"pipefitter": "plumber",
	}}
}

func (n Normalizer) WorkerName(value string) (string, error) {
	cleaned := collapseSpaces(value)
	if cleaned == "" {
		return "", fmt.Errorf("worker name is required")
	}
	if len([]rune(cleaned)) > 80 {
		return "", fmt.Errorf("worker name is too long")
	}
	for _, char := range cleaned {
		if unicode.IsControl(char) {
			return "", fmt.Errorf("worker name contains control character")
		}
	}
	return cleaned, nil
}

func (n Normalizer) Trade(value string) (string, error) {
	cleaned := strings.ToLower(collapseSpaces(value))
	if alias, ok := n.Aliases[cleaned]; ok {
		cleaned = alias
	}
	if cleaned == "" {
		return "", fmt.Errorf("trade is required")
	}
	return cleaned, nil
}

func (n Normalizer) Entry(entry domain.WorkEntry) (domain.WorkEntry, error) {
	name, err := n.WorkerName(entry.WorkerName)
	if err != nil {
		return domain.WorkEntry{}, err
	}
	trade, err := n.Trade(entry.Trade)
	if err != nil {
		return domain.WorkEntry{}, err
	}
	entry.WorkerName, entry.Trade = name, trade
	if err := entry.Validate(); err != nil {
		return domain.WorkEntry{}, err
	}
	return entry, nil
}

func collapseSpaces(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func DuplicateEntries(entries []domain.WorkEntry) map[string][]int {
	groups := make(map[string][]int)
	for index, entry := range entries {
		key := strings.ToLower(strings.TrimSpace(entry.WorkerName)) + "|" + strings.ToLower(strings.TrimSpace(entry.Trade)) + "|" + entry.WorkDate
		groups[key] = append(groups[key], index)
	}
	for key, indexes := range groups {
		if len(indexes) < 2 {
			delete(groups, key)
		}
	}
	return groups
}

func SortedDuplicateKeys(entries []domain.WorkEntry) []string {
	duplicates := DuplicateEntries(entries)
	keys := make([]string, 0, len(duplicates))
	for key := range duplicates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func Deduplicate(entries []domain.WorkEntry) ([]domain.WorkEntry, int) {
	seen := make(map[string]bool)
	result := make([]domain.WorkEntry, 0, len(entries))
	removed := 0
	for _, entry := range entries {
		key := strings.ToLower(entry.WorkerName) + "|" + strings.ToLower(entry.Trade) + "|" + entry.WorkDate + "|" + fmt.Sprintf("%d|%.2f|%.2f", entry.CompletedPieces, entry.UnitPrice, entry.NightAllowance)
		if seen[key] {
			removed++
			continue
		}
		seen[key] = true
		result = append(result, entry)
	}
	return result, removed
}

func PartitionByTrade(entries []domain.WorkEntry) map[string][]domain.WorkEntry {
	result := make(map[string][]domain.WorkEntry)
	for _, entry := range entries {
		key := strings.ToLower(entry.Trade)
		result[key] = append(result[key], entry)
	}
	return result
}

func WorkDateRange(entries []domain.WorkEntry) (string, string) {
	if len(entries) == 0 {
		return "", ""
	}
	min, max := entries[0].WorkDate, entries[0].WorkDate
	for _, entry := range entries[1:] {
		if entry.WorkDate < min {
			min = entry.WorkDate
		}
		if entry.WorkDate > max {
			max = entry.WorkDate
		}
	}
	return min, max
}

func IssueForNormalization(line int, column, value string, err error) domain.ImportIssue {
	return domain.ImportIssue{Line: line, Column: column, Value: value, Message: err.Error()}
}

func NormalizeResult(result domain.ImportResult) domain.ImportResult {
	normalizer := NewNormalizer()
	clean := domain.ImportResult{Total: result.Total, Issues: append([]domain.ImportIssue(nil), result.Issues...)}
	for _, entry := range result.Entries {
		normalized, err := normalizer.Entry(entry)
		if err != nil {
			clean.Issues = append(clean.Issues, IssueForNormalization(entry.ImportedLine, "worker_name", entry.WorkerName, err))
			continue
		}
		clean.Entries = append(clean.Entries, normalized)
	}
	return clean
}
