// Package refapp is the reference application Leitwerk governs in examples/.
// It is deliberately tiny: just enough real code to carry a spec, a runnable
// test, and a T2 migration path, so the gate demonstrates governance rather
// than only execution.
package refapp

// LineItem is one line on an order: a unit price in cents and a quantity.
type LineItem struct {
	UnitCents int
	Qty       int
}

// OrderTotalCents returns the total price of an order in cents. Callers are
// responsible for non-negative inputs (the spec states this as the contract);
// the function itself performs no validation.
func OrderTotalCents(items []LineItem) int {
	total := 0
	for _, it := range items {
		total += it.UnitCents * it.Qty
	}
	return total
}
