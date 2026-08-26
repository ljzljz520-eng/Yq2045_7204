package importer

import (
	"fmt"
	"sort"
	"strings"

	"sitepay/internal/domain"
)

type ErrorReport struct {
	Issues []domain.ImportIssue
}

func NewErrorReport(issues []domain.ImportIssue) ErrorReport {
	copyIssues := append([]domain.ImportIssue(nil), issues...)
	sort.SliceStable(copyIssues, func(i, j int) bool { return copyIssues[i].Line < copyIssues[j].Line })
	return ErrorReport{Issues: copyIssues}
}

func (r ErrorReport) Empty() bool { return len(r.Issues) == 0 }

func (r ErrorReport) String() string {
	if r.Empty() {
		return "no import errors"
	}
	lines := make([]string, 0, len(r.Issues))
	for _, issue := range r.Issues {
		location := fmt.Sprintf("line %d", issue.Line)
		if issue.Column != "" {
			location += " " + issue.Column
		}
		message := issue.Message
		if issue.Value != "" {
			message += fmt.Sprintf(" (value %q)", issue.Value)
		}
		lines = append(lines, location+": "+message)
	}
	return strings.Join(lines, "\n")
}

func GroupByLine(issues []domain.ImportIssue) map[int][]domain.ImportIssue {
	groups := make(map[int][]domain.ImportIssue)
	for _, issue := range issues {
		groups[issue.Line] = append(groups[issue.Line], issue)
	}
	return groups
}

func CountColumns(issues []domain.ImportIssue) map[string]int {
	counts := make(map[string]int)
	for _, issue := range issues {
		key := issue.Column
		if key == "" {
			key = "row"
		}
		counts[key]++
	}
	return counts
}
