package hitung

// PromoKind is the type of a selling promo.
type PromoKind int

const (
	// NoPromo: full price.
	NoPromo PromoKind = iota
	// PercentOff: PercentBP basis points off the line.
	PercentOff
	// AmountOff: Amount rupiah off each unit, never below zero.
	AmountOff
	// BuyXGetY: for every Buy+Get units, Get are free ("beli 2 gratis 1").
	BuyXGetY
)

// Promo is a selling promo for one item and unit.
type Promo struct {
	Kind      PromoKind
	PercentBP int64
	Amount    int64
	Buy, Get  int64
}

// Priced is one cashier line after its promo.
type Priced struct {
	Gross    int64
	Discount int64
	Free     int64 // free units, for BuyXGetY
	Net      int64
}

// Price applies a promo to qty units at unitPrice.
//
//	Price(5000, 2, Promo{Kind: BuyXGetY, Buy: 1, Get: 1}) == Priced{Gross: 10000, Discount: 5000, Free: 1, Net: 5000}
func Price(unitPrice, qty int64, p Promo) Priced {
	gross := unitPrice * qty
	var disc, free int64
	switch p.Kind {
	case PercentOff:
		disc = mulDivRound(gross, p.PercentBP, bpDenominator)
	case AmountOff:
		disc = min(gross, p.Amount*qty)
	case BuyXGetY:
		if p.Buy+p.Get > 0 {
			free = qty / (p.Buy + p.Get) * p.Get
			disc = free * unitPrice
		}
	}
	return Priced{Gross: gross, Discount: disc, Free: free, Net: gross - disc}
}
