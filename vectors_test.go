package hitung

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

// The vectors in testdata/vectors.json were computed with exact fractions,
// separately from this code. The JavaScript package runs the same file.
type vectors struct {
	Rupiah []struct {
		N    int64
		Text string
	}
	ParseRupiah []struct {
		Text string
		N    int64
	}
	ParseRupiahInvalid []string
	PPN                []struct {
		Amount, RateBP, DPP, PPN, Total int64
		Mode                            string
	}
	Discount []struct {
		Amount      int64
		DiscountsBP []int64
		Result      int64
	}
	Price []struct {
		UnitPrice, Qty, Gross, Discount, Free, Net int64
		Promo                                      struct {
			Kind                        string
			PercentBP, Amount, Buy, Get int64
		}
	}
	Invoice []struct {
		Lines          []struct{ Qty, Price, DiscBP int64 }
		DiscBP, RateBP int64
		Mode           string
		PKP            bool
		Totals         struct {
			Gross, LineDisc, InvoiceDisc, Net, DPP, PPN, Total int64
		}
		LandedCosts []int64
	}
	MovingAverage []struct {
		StockQty, AvgCost, PaidQty, BonusQty, TotalCost, Result int64
	}
	FEFO []struct {
		Batches []struct {
			ID     string
			Expiry string
			Qty    int64
		}
		Qty   int64
		Today string
		Picks []struct {
			BatchID string
			Expiry  string
			Qty     int64
		}
		Short int64
	}
	ReorderPoint []struct{ Sold, Days, LeadDays, SafetyDays, Result int64 }
	OrderPacks   []struct{ Stock, Target, PackSize, Result int64 }
}

func load(t *testing.T) vectors {
	t.Helper()
	raw, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var v vectors
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func mode(t *testing.T, s string) PPNMode {
	switch s {
	case "exclusive":
		return Exclusive
	case "inclusive":
		return Inclusive
	case "none":
		return NoPPN
	}
	t.Fatalf("unknown mode %q", s)
	return NoPPN
}

func day(t *testing.T, s string) time.Time {
	d, err := time.Parse(time.DateOnly, s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestRupiah(t *testing.T) {
	for _, c := range load(t).Rupiah {
		if got := Rupiah(c.N); got != c.Text {
			t.Errorf("Rupiah(%d) = %q, want %q", c.N, got, c.Text)
		}
	}
	if got := Rupiah(math.MinInt64); got != "-Rp 9.223.372.036.854.775.808" {
		t.Errorf("Rupiah(MinInt64) = %q", got)
	}
}

func TestParseRupiah(t *testing.T) {
	v := load(t)
	for _, c := range v.ParseRupiah {
		got, err := ParseRupiah(c.Text)
		if err != nil || got != c.N {
			t.Errorf("ParseRupiah(%q) = %d, %v; want %d", c.Text, got, err, c.N)
		}
	}
	for _, s := range v.ParseRupiahInvalid {
		if got, err := ParseRupiah(s); err == nil {
			t.Errorf("ParseRupiah(%q) = %d, want error", s, got)
		}
	}
	for _, c := range v.Rupiah {
		if got, err := ParseRupiah(Rupiah(c.N)); err != nil || got != c.N {
			t.Errorf("round trip %d: got %d, %v", c.N, got, err)
		}
	}
}

func TestPPN(t *testing.T) {
	for _, c := range load(t).PPN {
		got := PPN(c.Amount, mode(t, c.Mode), c.RateBP)
		want := Tax{DPP: c.DPP, PPN: c.PPN, Total: c.Total}
		if got != want {
			t.Errorf("PPN(%d, %s, %d) = %+v, want %+v", c.Amount, c.Mode, c.RateBP, got, want)
		}
	}
}

func TestDiscount(t *testing.T) {
	for _, c := range load(t).Discount {
		if got := Discount(c.Amount, c.DiscountsBP...); got != c.Result {
			t.Errorf("Discount(%d, %v) = %d, want %d", c.Amount, c.DiscountsBP, got, c.Result)
		}
	}
}

func TestPrice(t *testing.T) {
	kinds := map[string]PromoKind{"none": NoPromo, "percent": PercentOff, "amount": AmountOff, "buyxgety": BuyXGetY}
	for _, c := range load(t).Price {
		p := Promo{Kind: kinds[c.Promo.Kind], PercentBP: c.Promo.PercentBP, Amount: c.Promo.Amount, Buy: c.Promo.Buy, Get: c.Promo.Get}
		got := Price(c.UnitPrice, c.Qty, p)
		want := Priced{Gross: c.Gross, Discount: c.Discount, Free: c.Free, Net: c.Net}
		if got != want {
			t.Errorf("Price(%d, %d, %+v) = %+v, want %+v", c.UnitPrice, c.Qty, p, got, want)
		}
	}
}

func TestInvoice(t *testing.T) {
	for i, c := range load(t).Invoice {
		inv := Invoice{DiscBP: c.DiscBP, Mode: mode(t, c.Mode), RateBP: c.RateBP}
		for _, l := range c.Lines {
			inv.Lines = append(inv.Lines, Line{Qty: l.Qty, Price: l.Price, DiscBP: l.DiscBP})
		}
		got := inv.Totals()
		want := Totals(c.Totals)
		if got != want {
			t.Errorf("invoice %d: Totals() = %+v, want %+v", i, got, want)
		}
		lc := inv.LandedCosts(c.PKP)
		if !reflect.DeepEqual(lc, c.LandedCosts) {
			t.Errorf("invoice %d: LandedCosts(%v) = %v, want %v", i, c.PKP, lc, c.LandedCosts)
		}
	}
}

func TestMovingAverage(t *testing.T) {
	for _, c := range load(t).MovingAverage {
		if got := MovingAverage(c.StockQty, c.AvgCost, c.PaidQty, c.BonusQty, c.TotalCost); got != c.Result {
			t.Errorf("MovingAverage(%d, %d, %d, %d, %d) = %d, want %d", c.StockQty, c.AvgCost, c.PaidQty, c.BonusQty, c.TotalCost, got, c.Result)
		}
	}
}

func TestPickFEFO(t *testing.T) {
	for i, c := range load(t).FEFO {
		var bs []Batch
		for _, b := range c.Batches {
			bs = append(bs, Batch{ID: b.ID, Expiry: day(t, b.Expiry), Qty: b.Qty})
		}
		before := append([]Batch(nil), bs...)
		picks, short := PickFEFO(bs, c.Qty, day(t, c.Today))
		if !reflect.DeepEqual(bs, before) {
			t.Errorf("case %d: PickFEFO changed its input", i)
		}
		var want []Pick
		for _, p := range c.Picks {
			want = append(want, Pick{BatchID: p.BatchID, Expiry: day(t, p.Expiry), Qty: p.Qty})
		}
		if !reflect.DeepEqual(picks, want) || short != c.Short {
			t.Errorf("case %d: PickFEFO(qty %d, today %s) = %v, short %d; want %v, short %d", i, c.Qty, c.Today, picks, short, want, c.Short)
		}
	}
}

func TestPickFEFOTimeOfDay(t *testing.T) {
	// A batch expiring today stays sellable all day, whatever the clock says.
	jkt := time.FixedZone("WIB", 7*3600)
	bs := []Batch{{ID: "A", Expiry: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC), Qty: 1}}
	picks, short := PickFEFO(bs, 1, time.Date(2026, 11, 18, 23, 30, 0, 0, jkt))
	if len(picks) != 1 || short != 0 {
		t.Errorf("batch expiring today not picked late in the day: %v, short %d", picks, short)
	}
}

func TestReorderPoint(t *testing.T) {
	for _, c := range load(t).ReorderPoint {
		if got := ReorderPoint(c.Sold, c.Days, c.LeadDays, c.SafetyDays); got != c.Result {
			t.Errorf("ReorderPoint(%d, %d, %d, %d) = %d, want %d", c.Sold, c.Days, c.LeadDays, c.SafetyDays, got, c.Result)
		}
	}
}

func TestOrderPacks(t *testing.T) {
	for _, c := range load(t).OrderPacks {
		if got := OrderPacks(c.Stock, c.Target, c.PackSize); got != c.Result {
			t.Errorf("OrderPacks(%d, %d, %d) = %d, want %d", c.Stock, c.Target, c.PackSize, got, c.Result)
		}
	}
}

func TestLandedCostsSumExactly(t *testing.T) {
	inv := Invoice{Lines: []Line{{7, 333, 0}, {11, 101, 150}, {3, 9999, 900}}, DiscBP: 275, Mode: Exclusive, RateBP: 1100}
	for _, pkp := range []bool{false, true} {
		var sum int64
		for _, c := range inv.LandedCosts(pkp) {
			sum += c
		}
		tt := inv.Totals()
		want := tt.Total
		if pkp {
			want = tt.DPP
		}
		if sum != want {
			t.Errorf("pkp=%v: landed costs sum to %d, want %d", pkp, sum, want)
		}
	}
}
