package importer

import (
	"strings"
	"testing"
)

func TestCSVImportReportsBadRows(t *testing.T) {
	data := "worker_name,trade,completed_pieces,unit_price,night_allowance,quality_deduction,work_date\nLin,mason,5,3.20,12.75,0.50,2026-08-25\nBad,mason,nope,3,2,0,2026-08-25\n"
	result, err := NewCSVImporter().Read(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if result.Accepted() != 1 || result.Rejected() != 1 || result.Total != 2 {
		t.Fatalf("unexpected result %+v", result)
	}
	if result.Entries[0].NightAllowance != 12.75 {
		t.Fatalf("allowance=%v", result.Entries[0].NightAllowance)
	}
}
