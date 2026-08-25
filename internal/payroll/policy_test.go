package payroll

import (
	"testing"

	"sitepay/internal/domain"
)

func TestAllowanceReview(t *testing.T) {
	policy := domain.AllowancePolicy{Trade: "mason", NightCap: 20, RequiresReview: true, Active: true}
	needed, reason := AllowanceReview(21.5, policy)
	if !needed || reason == "" {
		t.Fatalf("expected review, got %v %q", needed, reason)
	}
	needed, _ = AllowanceReview(19.5, policy)
	if needed {
		t.Fatal("did not expect review under cap")
	}
}

func TestPolicyBook(t *testing.T) {
	book, err := NewPolicyBook(DefaultPolicies())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := book.Resolve("mason"); err != nil {
		t.Fatal(err)
	}
	if _, err := book.Resolve("unknown"); err == nil {
		t.Fatal("expected missing policy")
	}
}
