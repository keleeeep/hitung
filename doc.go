// Package hitung does the money and stock arithmetic that Indonesian retail
// and pharmacy software needs: rupiah formatting, PPN, stacked discounts,
// supplier invoice totals, landed cost per line, promos, moving-average cost,
// first-expiry-first-out picking and reorder points.
//
// Every amount is an int64 number of rupiah. Every percentage is in basis
// points (1100 = 11%, 250 = 2.5%), so no float ever touches money. Each
// result is rounded once, half away from zero, at the step a printed invoice
// would round it.
//
// The JavaScript package in js/ has the same functions and passes the same
// test vectors (testdata/vectors.json).
package hitung
