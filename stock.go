package hitung

import (
	"math/big"
	"sort"
	"time"
)

// MovingAverage returns the new cost per unit after goods arrive.
// stockQty units already cost avgCost each. The delivery has paidQty units
// and bonusQty free units, and costs totalCost in all (use LandedCosts for
// one invoice line). Free units lower the average. Negative stock counts as
// zero; an empty result keeps avgCost.
//
//	MovingAverage(100, 300, 100, 10, 27000) == 271 // (30000 + 27000) / 210
func MovingAverage(stockQty, avgCost, paidQty, bonusQty, totalCost int64) int64 {
	if stockQty < 0 {
		stockQty = 0
	}
	units := stockQty + paidQty + bonusQty
	if units <= 0 {
		return avgCost
	}
	value := new(big.Int).Mul(big.NewInt(stockQty), big.NewInt(avgCost))
	value.Add(value, big.NewInt(totalCost))
	return divRound(value, big.NewInt(units))
}

// Batch is stock of one item with one expiry date.
type Batch struct {
	ID     string
	Expiry time.Time
	Qty    int64
}

// Pick is how much to take from one batch.
type Pick struct {
	BatchID string
	Expiry  time.Time
	Qty     int64
}

// PickFEFO takes qty units first-expiry-first-out: the batch that expires
// soonest goes first, and batches with the same date keep their given order.
// Batches that expired before today are never picked; a batch whose expiry
// date is today can still be sold. short is what the sellable batches could
// not cover. The input is not changed.
func PickFEFO(batches []Batch, qty int64, today time.Time) (picks []Pick, short int64) {
	today = dateIn(today, today.Location())
	bs := make([]Batch, 0, len(batches))
	for _, b := range batches {
		if b.Qty > 0 && !dateIn(b.Expiry, today.Location()).Before(today) {
			bs = append(bs, b)
		}
	}
	sort.SliceStable(bs, func(i, j int) bool { return bs[i].Expiry.Before(bs[j].Expiry) })
	for _, b := range bs {
		if qty <= 0 {
			break
		}
		take := min(b.Qty, qty)
		picks = append(picks, Pick{BatchID: b.ID, Expiry: b.Expiry, Qty: take})
		qty -= take
	}
	return picks, max(qty, 0)
}

// ReorderPoint is the stock level at which to order again: average daily
// sales (sold over days) times the days until goods arrive plus safety days,
// rounded up. Integer inputs keep it exact.
//
//	ReorderPoint(600, 30, 2, 3) == 100 // 20 a day × 5 days
func ReorderPoint(sold int64, days, leadDays, safetyDays int64) int64 {
	if days <= 0 || sold <= 0 {
		return 0
	}
	need := sold * (leadDays + safetyDays)
	return (need + days - 1) / days
}

// OrderPacks is how many packs of packSize bring stock up to target,
// rounded up. It is 0 when stock already reaches target.
func OrderPacks(stock, target, packSize int64) int64 {
	if stock >= target || packSize <= 0 {
		return 0
	}
	gap := target - stock
	return (gap + packSize - 1) / packSize
}

// dateIn keeps t's calendar date (as written, in t's own location) and puts
// it at midnight in loc, so two dates compare by day only.
func dateIn(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}
