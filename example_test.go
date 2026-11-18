package hitung_test

import (
	"fmt"
	"time"

	"github.com/keleeeep/hitung"
)

func Example() {
	// A supplier invoice: 5% off the first line, 3% off the whole invoice,
	// prices exclude PPN at 11%.
	inv := hitung.Invoice{
		Lines: []hitung.Line{
			{Qty: 24, Price: 9600, DiscBP: 500},
			{Qty: 10, Price: 3800},
		},
		DiscBP: 300,
		Mode:   hitung.Exclusive,
		RateBP: 1100,
	}
	t := inv.Totals()
	fmt.Println("Net:  ", hitung.Rupiah(t.Net))
	fmt.Println("PPN:  ", hitung.Rupiah(t.PPN))
	fmt.Println("Total:", hitung.Rupiah(t.Total))

	// Not registered for PPN, so the tax is part of the cost of each line.
	fmt.Println("Cost per line:", inv.LandedCosts(false))
	// Output:
	// Net:   Rp 249.174
	// PPN:   Rp 27.409
	// Total: Rp 276.583
	// Cost per line: [235668 40915]
}

func ExamplePrice() {
	p := hitung.Price(5000, 3, hitung.Promo{Kind: hitung.BuyXGetY, Buy: 1, Get: 1})
	fmt.Println(hitung.Rupiah(p.Net), "with", p.Free, "free")
	// Output: Rp 10.000 with 1 free
}

func ExamplePickFEFO() {
	day := func(s string) time.Time { d, _ := time.Parse(time.DateOnly, s); return d }
	batches := []hitung.Batch{
		{ID: "AMX2602A", Expiry: day("2027-02-01"), Qty: 100},
		{ID: "AMX2411B", Expiry: day("2026-12-20"), Qty: 6},
		{ID: "AMX2409C", Expiry: day("2026-09-30"), Qty: 40}, // expired, never sold
	}
	picks, short := hitung.PickFEFO(batches, 10, day("2026-11-18"))
	for _, p := range picks {
		fmt.Println(p.BatchID, p.Qty)
	}
	fmt.Println("short:", short)
	// Output:
	// AMX2411B 6
	// AMX2602A 4
	// short: 0
}
