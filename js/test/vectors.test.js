// Runs the shared vectors in ../../testdata/vectors.json, the same file the Go tests use.
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import * as h from '../src/index.js';

const v = JSON.parse(readFileSync(new URL('../../testdata/vectors.json', import.meta.url), 'utf8'));

test('rupiah', () => {
  for (const c of v.rupiah) assert.equal(h.rupiah(c.n), c.text);
  assert.throws(() => h.rupiah(1.5), TypeError);
});

test('parseRupiah', () => {
  for (const c of v.parseRupiah) assert.equal(h.parseRupiah(c.text), c.n, c.text);
  for (const s of v.parseRupiahInvalid) assert.throws(() => h.parseRupiah(s), RangeError, s);
  for (const c of v.rupiah) assert.equal(h.parseRupiah(h.rupiah(c.n)), c.n);
});

test('ppn', () => {
  for (const c of v.ppn) assert.deepEqual(h.ppn(c.amount, c.mode, c.rateBP), { dpp: c.dpp, ppn: c.ppn, total: c.total });
});

test('discount', () => {
  for (const c of v.discount) assert.equal(h.discount(c.amount, c.discountsBP), c.result);
});

test('price', () => {
  for (const c of v.price) {
    assert.deepEqual(h.price(c.unitPrice, c.qty, c.promo), { gross: c.gross, discount: c.discount, free: c.free, net: c.net });
  }
});

test('invoiceTotals and landedCosts', () => {
  for (const c of v.invoice) {
    assert.deepEqual(h.invoiceTotals(c), c.totals);
    assert.deepEqual(h.landedCosts(c, c.pkp), c.landedCosts);
  }
});

test('movingAverage', () => {
  for (const c of v.movingAverage) {
    assert.equal(h.movingAverage(c.stockQty, c.avgCost, c.paidQty, c.bonusQty, c.totalCost), c.result);
  }
});

test('pickFEFO', () => {
  for (const c of v.fefo) {
    const before = structuredClone(c.batches);
    assert.deepEqual(h.pickFEFO(c.batches, c.qty, c.today), { picks: c.picks, short: c.short });
    assert.deepEqual(c.batches, before, 'input changed');
  }
});

test('reorderPoint and orderPacks', () => {
  for (const c of v.reorderPoint) assert.equal(h.reorderPoint(c.sold, c.days, c.leadDays, c.safetyDays), c.result);
  for (const c of v.orderPacks) assert.equal(h.orderPacks(c.stock, c.target, c.packSize), c.result);
});
