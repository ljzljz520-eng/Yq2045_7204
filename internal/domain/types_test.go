package domain

import "testing"

func TestEntityValidation(t *testing.T) {
	worker := Worker{Name: "Lin", Trade: "mason"}
	if err := worker.Validate(); err != nil {
		t.Fatal(err)
	}
	entry := WorkEntry{WorkerName: "Lin", Trade: "mason", CompletedPieces: 3, UnitPrice: 2.5, WorkDate: "2026-08-25"}
	if err := entry.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (WorkEntry{}).Validate(); err == nil {
		t.Fatal("expected invalid entry")
	}
}
