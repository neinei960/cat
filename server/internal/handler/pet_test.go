package handler

import "testing"

func TestParseBirthDateSupportsSingleDigitMonthAndDay(t *testing.T) {
	birthDate := parseBirthDate("2019-11-4")
	if birthDate == nil {
		t.Fatalf("expected single-digit day birth date to parse")
	}
	if got := birthDate.Format("2006-01-02"); got != "2019-11-04" {
		t.Fatalf("expected normalized birth date 2019-11-04, got %s", got)
	}
}
