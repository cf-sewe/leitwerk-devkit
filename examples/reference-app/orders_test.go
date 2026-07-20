package refapp

import "testing"

// The executable oracle for the reference-app's contract (see
// leitwerk/specs/orders.md). Written before orders.go; it fails on a regression
// to OrderTotalCents and is what the gate's `tests` check runs.

func TestOrderTotalCents(t *testing.T) {
	got := OrderTotalCents([]LineItem{{UnitCents: 250, Qty: 2}, {UnitCents: 100, Qty: 3}})
	if want := 800; got != want {
		t.Fatalf("OrderTotalCents = %d, want %d", got, want)
	}
}

func TestOrderTotalCentsEmpty(t *testing.T) {
	if got := OrderTotalCents(nil); got != 0 {
		t.Fatalf("empty order total = %d, want 0", got)
	}
}
