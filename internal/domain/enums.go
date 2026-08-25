package domain

import "fmt"

const (
	StatementDraft    = "draft"
	StatementReview   = "review"
	StatementApproved = "approved"
	StatementExported = "exported"
)

const (
	AuditCreated  = "created"
	AuditImported = "imported"
	AuditApproved = "approved"
	AuditExported = "exported"
)

func ValidStatementStatus(status string) bool {
	switch status {
	case StatementDraft, StatementReview, StatementApproved, StatementExported:
		return true
	default:
		return false
	}
}

func NextStatementStatus(current string, reviewRequired bool) (string, error) {
	switch current {
	case StatementDraft:
		if reviewRequired {
			return StatementReview, nil
		}
		return StatementApproved, nil
	case StatementReview:
		return StatementApproved, nil
	case StatementApproved:
		return StatementExported, nil
	case StatementExported:
		return StatementExported, nil
	default:
		return "", fmt.Errorf("unknown statement status %q", current)
	}
}

func IsTerminalStatus(status string) bool { return status == StatementExported }

func CanTransition(current, next string) bool {
	if !ValidStatementStatus(current) || !ValidStatementStatus(next) {
		return false
	}
	if current == next {
		return current == StatementExported
	}
	valid := map[string]map[string]bool{
		StatementDraft:    {StatementReview: true, StatementApproved: true},
		StatementReview:   {StatementApproved: true},
		StatementApproved: {StatementExported: true},
		StatementExported: {},
	}
	return valid[current][next]
}
