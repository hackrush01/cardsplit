package parsers

import (
	"strings"
	"testing"
	"time"
)

func TestParseInfiniaCSV_DuplicateTimestamps(t *testing.T) {
	raw := `Domestic / International Transactions
Transaction type~|~Label~|~Date~|~Description~|~Amount~|~Dr/Cr~|~Rewards
Purchase~|~ABC~|~01/01/2024 10:00:00~|~Coffee~|~100.00~|~Dr~|~
Purchase~|~DEF~|~01/01/2024 10:00:00~|~Groceries~|~150.00~|~Dr~|~`

	stmt, err := ParseInfiniaCSV(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(stmt.Warnings) == 0 {
		t.Fatal("expected duplicate timestamp warning, got none")
	}

	if !strings.Contains(stmt.Warnings[0], "duplicate timestamp found") {
		t.Fatalf("unexpected warning message: %v", stmt.Warnings[0])
	}

	if len(stmt.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(stmt.Transactions))
	}

	firstActual := stmt.Transactions[0].ActualTimestamp
	secondActual := stmt.Transactions[1].ActualTimestamp
	if !secondActual.Equal(firstActual) {
		t.Fatalf("expected actual timestamps to remain equal; got %v and %v", firstActual, secondActual)
	}

	firstShifted := stmt.Transactions[0].ShiftedTimestamp
	secondShifted := stmt.Transactions[1].ShiftedTimestamp
	if !secondShifted.Equal(firstShifted.Add(time.Second)) {
		t.Fatalf("expected second shifted timestamp to be shifted by one second; got %v and %v", firstShifted, secondShifted)
	}
}
