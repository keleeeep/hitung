package hitung

import (
	"math/big"
	"sort"
)

// Line is one line of a supplier invoice (faktur): Qty units at Price each,
// with an optional line discount.
type Line struct {
	Qty    int64
	Price  int64
	DiscBP int64
}

// Invoice is a supplier invoice: lines, an optional discount on the whole
// invoice, and how PPN applies.
type Invoice struct {
	Lines  []Line
	DiscBP int64
	Mode   PPNMode
	RateBP int64
}

// Totals is the footer of an invoice. Gross - LineDisc - InvoiceDisc = Net,
// and Net goes through PPN to give DPP, PPN and Total.
type Totals struct {
	Gross       int64
	LineDisc    int64
	InvoiceDisc int64
	Net         int64
	DPP         int64
	PPN         int64
	Total       int64
}

// lineNets returns each line after its own discount, rounded per line as a
// printed invoice does.
func (inv Invoice) lineNets() []int64 {
	out := make([]int64, len(inv.Lines))
	for i, l := range inv.Lines {
		out[i] = Discount(l.Qty*l.Price, l.DiscBP)
	}
	return out
}

// Totals computes the invoice footer.
func (inv Invoice) Totals() Totals {
	var t Totals
	var afterLine int64
	for i, n := range inv.lineNets() {
		t.Gross += inv.Lines[i].Qty * inv.Lines[i].Price
		afterLine += n
	}
	t.LineDisc = t.Gross - afterLine
	t.Net = Discount(afterLine, inv.DiscBP)
	t.InvoiceDisc = afterLine - t.Net
	tax := PPN(t.Net, inv.Mode, inv.RateBP)
	t.DPP, t.PPN, t.Total = tax.DPP, tax.PPN, tax.Total
	return t
}

// LandedCosts spreads what the invoice really costs over its lines, in
// proportion to each line's net, so the result sums exactly to the cost.
// Use it to update stock cost per item.
//
// A business registered for PPN (PKP) credits the input tax, so its cost is
// the DPP. A business that is not PKP cannot, so PPN is part of its cost and
// the lines sum to Total.
func (inv Invoice) LandedCosts(pkp bool) []int64 {
	t := inv.Totals()
	target := t.Total
	if pkp {
		target = t.DPP
	}
	return allocate(target, inv.lineNets())
}

// allocate splits total over weights with the largest remainder method, so
// the parts are proportional and sum exactly to total. Zero total weight
// gives all zeros.
func allocate(total int64, weights []int64) []int64 {
	out := make([]int64, len(weights))
	var sum int64
	for _, w := range weights {
		sum += w
	}
	if sum == 0 || len(weights) == 0 {
		return out
	}
	type rem struct {
		i int
		r *big.Int
	}
	rems := make([]rem, len(weights))
	var given int64
	bigTotal, bigSum := big.NewInt(total), big.NewInt(sum)
	for i, w := range weights {
		q, r := new(big.Int).QuoRem(new(big.Int).Mul(bigTotal, big.NewInt(w)), bigSum, new(big.Int))
		out[i] = q.Int64()
		given += out[i]
		rems[i] = rem{i, r}
	}
	sort.SliceStable(rems, func(a, b int) bool { return rems[a].r.Cmp(rems[b].r) > 0 })
	for k := 0; given < total && k < len(rems); k++ {
		out[rems[k].i]++
		given++
	}
	return out
}
