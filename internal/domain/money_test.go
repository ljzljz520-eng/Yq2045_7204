package domain

import "testing"

func TestMoneyFormatting(t *testing.T) {
	for _, test := range []struct {
		value int64
		want  string
	}{
		{0, "0.00"}, {7, "0.07"}, {1250, "12.50"}, {-125, "-1.25"},
	} {
		if got := FormatCents(test.value); got != test.want {
			t.Fatalf("FormatCents(%d)=%q, want %q", test.value, got, test.want)
		}
	}
}

func TestParseMoney(t *testing.T) {
	value, err := ParseMoney("18.375")
	if err != nil || value != 1838 {
		t.Fatalf("ParseMoney returned %d %v", value, err)
	}
	if _, err := ParseMoney("1.234"); err == nil {
		t.Fatal("expected invalid precision")
	}
}
