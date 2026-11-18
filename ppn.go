package hitung

// PPNMode says how a price relates to PPN (Pajak Pertambahan Nilai).
type PPNMode int

const (
	// Exclusive: the price excludes PPN, which is added on top.
	Exclusive PPNMode = iota
	// Inclusive: the price already includes PPN, which is split out of it.
	Inclusive
	// NoPPN: no PPN on this invoice.
	NoPPN
)

func (m PPNMode) String() string {
	switch m {
	case Exclusive:
		return "exclusive"
	case Inclusive:
		return "inclusive"
	case NoPPN:
		return "none"
	}
	return "unknown"
}

// Tax is the PPN breakdown of one amount. DPP (dasar pengenaan pajak) is the
// base the tax is charged on; Total is what the buyer pays.
type Tax struct {
	DPP   int64
	PPN   int64
	Total int64
}

// PPN splits or adds tax at rateBP basis points (1100 = 11%). The rate is a
// parameter on purpose: use the rate printed on the invoice.
//
//	PPN(100000, Exclusive, 1100) == Tax{DPP: 100000, PPN: 11000, Total: 111000}
//	PPN(111000, Inclusive, 1100) == Tax{DPP: 100000, PPN: 11000, Total: 111000}
func PPN(amount int64, mode PPNMode, rateBP int64) Tax {
	switch mode {
	case Exclusive:
		tax := mulDivRound(amount, rateBP, bpDenominator)
		return Tax{DPP: amount, PPN: tax, Total: amount + tax}
	case Inclusive:
		dpp := mulDivRound(amount, bpDenominator, bpDenominator+rateBP)
		return Tax{DPP: dpp, PPN: amount - dpp, Total: amount}
	default:
		return Tax{DPP: amount, Total: amount}
	}
}
