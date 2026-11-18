# hitung

[![ci](https://github.com/keleeeep/hitung/actions/workflows/ci.yml/badge.svg)](https://github.com/keleeeep/hitung/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/keleeeep/hitung.svg)](https://pkg.go.dev/github.com/keleeeep/hitung)

Money and stock arithmetic for Indonesian retail and pharmacy software, in Go and JavaScript. *Hitung* means "calculate".

- **Rupiah**: format and parse `Rp 1.234.567`.
- **PPN**: add tax on top of a price, or split it out of a price that includes it.
- **Discounts**: stacked percentages (5% then 3%), rounded once.
- **Supplier invoices (faktur)**: line discounts, invoice discount, PPN, and the landed cost of each line, which sums exactly to what you paid.
- **Promos**: percent off, rupiah off per unit, buy X get Y free.
- **Stock cost**: moving-average cost, with free bonus goods lowering the average.
- **FEFO**: pick stock from the batch that expires first, and never from an expired batch.
- **Reordering**: reorder point from sales history, and how many packs to order.

Amounts are whole rupiah in integers, and percentages are basis points (`1100` = 11%), so no float touches money. Each result is rounded once, half away from zero, at the step a printed invoice rounds it. The Go and JavaScript versions pass the same [test vectors](testdata/vectors.json), which were computed separately with exact fractions.

## Go

```sh
go get github.com/keleeeep/hitung
```

```go
inv := hitung.Invoice{
	Lines: []hitung.Line{
		{Qty: 24, Price: 9600, DiscBP: 500}, // 5% off this line
		{Qty: 10, Price: 3800},
	},
	DiscBP: 300,              // 3% off the whole invoice
	Mode:   hitung.Exclusive, // prices exclude PPN
	RateBP: 1100,             // 11%
}
t := inv.Totals()
fmt.Println(hitung.Rupiah(t.Total)) // Rp 276.583

// Not registered for PPN (non-PKP): the tax is part of what each line cost.
costs := inv.LandedCosts(false) // [235668 40915]

// Selling: buy 1 get 1 free.
p := hitung.Price(5000, 3, hitung.Promo{Kind: hitung.BuyXGetY, Buy: 1, Get: 1})
// p.Net == 10000, p.Free == 1

// FEFO: the expired batch is skipped.
picks, short := hitung.PickFEFO(batches, 10, time.Now())
```

## JavaScript

No dependencies, ES modules, Node 18+ and browsers.

```js
import { invoiceTotals, landedCosts, price, pickFEFO, rupiah } from '@keleeeep/hitung';

const inv = {
  lines: [{ qty: 24, price: 9600, discBP: 500 }, { qty: 10, price: 3800, discBP: 0 }],
  discBP: 300, mode: 'exclusive', rateBP: 1100,
};
rupiah(invoiceTotals(inv).total);            // 'Rp 276.583'
landedCosts(inv, false);                     // [235668, 40915]
price(5000, 3, { kind: 'buyxgety', buy: 1, get: 1 }); // { gross: 15000, discount: 5000, free: 1, net: 10000 }
pickFEFO([{ id: 'A', expiry: '2026-09-30', qty: 40 }, { id: 'B', expiry: '2027-02-01', qty: 100 }], 10, '2026-11-18');
// { picks: [{ batchID: 'B', expiry: '2027-02-01', qty: 10 }], short: 0 }
```

The JS package is in [`js/`](js/) and is not on npm yet; copy `js/src/index.js` or install from the repo.

## Notes on the rules

- **PPN rate is a parameter.** Pass the rate printed on the invoice. The library does not encode tax law.
- **PKP or not.** A business registered for PPN (PKP) credits the input tax, so its cost is the DPP: `LandedCosts(true)`. A business that is not PKP pays the tax as part of its cost: `LandedCosts(false)`.
- **Rounding.** Line discounts round per line, the invoice discount and PPN round on the invoice total, as a printed faktur does. Landed costs use the largest remainder method, so the parts always add up to the total.
- **Expiry.** A batch whose expiry date is today can still be sold. One that expired yesterday cannot.

## Development

```sh
go test -race ./...         # Go
cd js && npm test           # JavaScript, same vectors
```

To add a case, add it to `testdata/vectors.json` with a value you worked out by hand or with exact fractions, then run both suites.

## License

MIT
