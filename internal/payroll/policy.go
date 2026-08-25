package payroll

import (
	"fmt"
	"strings"

	"sitepay/internal/domain"
)

type PolicyBook struct {
	byTrade map[string]domain.AllowancePolicy
}

func NewPolicyBook(policies []domain.AllowancePolicy) (*PolicyBook, error) {
	book := &PolicyBook{byTrade: make(map[string]domain.AllowancePolicy)}
	for _, policy := range policies {
		if err := policy.Validate(); err != nil {
			return nil, err
		}
		key := strings.ToLower(strings.TrimSpace(policy.Trade))
		if _, exists := book.byTrade[key]; exists {
			return nil, fmt.Errorf("duplicate policy for %s", key)
		}
		book.byTrade[key] = policy
	}
	return book, nil
}

func (b *PolicyBook) Resolve(trade string) (domain.AllowancePolicy, error) {
	if b == nil {
		return domain.AllowancePolicy{}, fmt.Errorf("policy book is nil")
	}
	policy, ok := b.byTrade[strings.ToLower(strings.TrimSpace(trade))]
	if !ok || !policy.Active {
		return domain.AllowancePolicy{}, fmt.Errorf("no active policy for %s", trade)
	}
	return policy, nil
}

func (b *PolicyBook) Trades() []string {
	if b == nil {
		return nil
	}
	trades := make([]string, 0, len(b.byTrade))
	for trade := range b.byTrade {
		trades = append(trades, trade)
	}
	return trades
}

func AllowanceReview(amount float64, policy domain.AllowancePolicy) (bool, string) {
	if !policy.Active {
		return false, "policy inactive"
	}
	if amount > policy.NightCap {
		if policy.RequiresReview {
			return true, fmt.Sprintf("night allowance %.2f exceeds cap %.2f", amount, policy.NightCap)
		}
		return false, "cap exceeded but review disabled"
	}
	return false, ""
}

func DefaultPolicies() []domain.AllowancePolicy {
	return []domain.AllowancePolicy{
		{Trade: "mason", NightCap: 25, RequiresReview: true, Active: true},
		{Trade: "carpenter", NightCap: 30, RequiresReview: true, Active: true},
		{Trade: "steelworker", NightCap: 35, RequiresReview: true, Active: true},
		{Trade: "electrician", NightCap: 40, RequiresReview: true, Active: true},
		{Trade: "plumber", NightCap: 30, RequiresReview: true, Active: true},
	}
}
