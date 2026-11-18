// hitung: Indonesian retail and pharmacy arithmetic, the same rules as the Go package.
// Amounts are whole rupiah (safe integers). Percentages are basis points (1100 = 11%).
// Every result is rounded once, half away from zero.

const BP = 10000n;

function int(name, n) {
  if (!Number.isSafeInteger(n)) throw new TypeError(`hitung: ${name} must be a safe integer, got ${n}`);
  return BigInt(n);
}

// num / den rounded half away from zero, for BigInt inputs
function divRound(num, den) {
  let q = num / den;
  const r = num % den;
  const absR = r < 0n ? -r : r;
  const absD = den < 0n ? -den : den;
  if (absR * 2n >= absD) q += (num < 0n) !== (den < 0n) ? -1n : 1n;
  return Number(q);
}

/** Format whole rupiah: rupiah(1234567) === 'Rp 1.234.567', rupiah(-5000) === '-Rp 5.000'. */
export function rupiah(n) {
  const v = int('n', n);
  const digits = (v < 0n ? -v : v).toString();
  const grouped = digits.replace(/\B(?=(\d{3})+(?!\d))/g, '.');
  return (v < 0n ? '-' : '') + 'Rp ' + grouped;
}

/** Read 'Rp 1.234.567', 'Rp1.000', '1.234', '1234' or a negative of these. Throws on anything else. */
export function parseRupiah(text) {
  let s = String(text).trim();
  let neg = false;
  if (s.startsWith('-') || s.startsWith('−')) { neg = true; s = s.slice(1).trim(); }
  if (s.slice(0, 2).toLowerCase() === 'rp') s = s.slice(2).trim();
  if (!/^(\d{1,3}(\.\d{3})+|\d+)$/.test(s)) throw new RangeError(`hitung: not a rupiah amount: ${text}`);
  const n = Number(s.replace(/\./g, ''));
  if (!Number.isSafeInteger(n)) throw new RangeError(`hitung: amount too large: ${text}`);
  return neg ? -n : n;
}

/**
 * PPN on an amount. mode: 'exclusive' (added on top), 'inclusive' (split out), 'none'.
 * ppn(100000, 'exclusive', 1100) → { dpp: 100000, ppn: 11000, total: 111000 }
 */
export function ppn(amount, mode, rateBP) {
  const a = int('amount', amount);
  const r = int('rateBP', rateBP);
  if (mode === 'exclusive') {
    const tax = divRound(a * r, BP);
    return { dpp: amount, ppn: tax, total: amount + tax };
  }
  if (mode === 'inclusive') {
    const dpp = divRound(a * BP, BP + r);
    return { dpp, ppn: amount - dpp, total: amount };
  }
  return { dpp: amount, ppn: 0, total: amount };
}

/** Successive percentage discounts, rounded once: discount(100000, [500, 300]) === 92150. */
export function discount(amount, discountsBP = []) {
  let num = int('amount', amount);
  let den = 1n;
  for (const d of discountsBP) { num *= BP - int('discountBP', d); den *= BP; }
  return divRound(num, den);
}

/**
 * One cashier line after its promo. promo.kind: 'none' | 'percent' (percentBP) | 'amount' (amount per unit) | 'buyxgety' (buy, get).
 * price(5000, 2, { kind: 'buyxgety', buy: 1, get: 1 }) → { gross: 10000, discount: 5000, free: 1, net: 5000 }
 */
export function price(unitPrice, qty, promo = { kind: 'none' }) {
  int('unitPrice', unitPrice); int('qty', qty);
  const gross = unitPrice * qty;
  let disc = 0, free = 0;
  if (promo.kind === 'percent') disc = divRound(BigInt(gross) * int('percentBP', promo.percentBP), BP);
  if (promo.kind === 'amount') disc = Math.min(gross, promo.amount * qty);
  if (promo.kind === 'buyxgety' && promo.buy + promo.get > 0) {
    free = Math.floor(qty / (promo.buy + promo.get)) * promo.get;
    disc = free * unitPrice;
  }
  return { gross, discount: disc, free, net: gross - disc };
}

function lineNets(inv) {
  return inv.lines.map((l) => discount(l.qty * l.price, [l.discBP || 0]));
}

/**
 * Footer of a supplier invoice.
 * inv = { lines: [{ qty, price, discBP }], discBP, mode, rateBP }
 */
export function invoiceTotals(inv) {
  const nets = lineNets(inv);
  const gross = inv.lines.reduce((a, l) => a + l.qty * l.price, 0);
  const afterLine = nets.reduce((a, n) => a + n, 0);
  const net = discount(afterLine, [inv.discBP || 0]);
  const t = ppn(net, inv.mode, inv.rateBP || 0);
  return { gross, lineDisc: gross - afterLine, invoiceDisc: afterLine - net, net, dpp: t.dpp, ppn: t.ppn, total: t.total };
}

/**
 * What each invoice line really costs, summing exactly to the invoice cost:
 * the DPP when the business is PKP (it credits input PPN), else the total.
 */
export function landedCosts(inv, pkp) {
  const t = invoiceTotals(inv);
  return allocate(pkp ? t.dpp : t.total, lineNets(inv));
}

// largest remainder: proportional parts that sum exactly to total
function allocate(total, weights) {
  const sum = weights.reduce((a, w) => a + w, 0);
  if (sum === 0) return weights.map(() => 0);
  const T = BigInt(total), S = BigInt(sum);
  const parts = weights.map((w, i) => {
    const x = T * BigInt(w);
    return { i, q: x / S, r: x % S };
  });
  let given = parts.reduce((a, p) => a + p.q, 0n);
  const order = [...parts].sort((a, b) => (a.r === b.r ? a.i - b.i : a.r > b.r ? -1 : 1));
  for (let k = 0; given < T && k < order.length; k++, given++) order[k].q += 1n;
  return parts.map((p) => Number(p.q));
}

/**
 * New cost per unit after goods arrive: stock at avgCost, plus paidQty and free bonusQty costing totalCost.
 * movingAverage(100, 300, 100, 10, 27000) === 271
 */
export function movingAverage(stockQty, avgCost, paidQty, bonusQty, totalCost) {
  const s = Math.max(0, stockQty);
  const units = s + paidQty + bonusQty;
  if (units <= 0) return avgCost;
  return divRound(int('stockQty', s) * int('avgCost', avgCost) + int('totalCost', totalCost), BigInt(units));
}

/**
 * First-expiry-first-out. batches: [{ id, expiry: 'YYYY-MM-DD', qty }], today: 'YYYY-MM-DD'.
 * Batches that expired before today are skipped; one expiring today can still be sold.
 * Returns { picks: [{ batchID, expiry, qty }], short }. The input is not changed.
 */
export function pickFEFO(batches, qty, today) {
  const sellable = batches
    .map((b, i) => ({ b, i }))
    .filter(({ b }) => b.qty > 0 && b.expiry >= today)
    .sort((x, y) => (x.b.expiry < y.b.expiry ? -1 : x.b.expiry > y.b.expiry ? 1 : x.i - y.i));
  const picks = [];
  let left = qty;
  for (const { b } of sellable) {
    if (left <= 0) break;
    const take = Math.min(b.qty, left);
    picks.push({ batchID: b.id, expiry: b.expiry, qty: take });
    left -= take;
  }
  return { picks, short: Math.max(left, 0) };
}

/** Stock level to reorder at: ceil(sold / days × (leadDays + safetyDays)). reorderPoint(600, 30, 2, 3) === 100 */
export function reorderPoint(sold, days, leadDays, safetyDays) {
  if (days <= 0 || sold <= 0) return 0;
  return Math.ceil((sold * (leadDays + safetyDays)) / days);
}

/** Packs of packSize needed to bring stock up to target, rounded up; 0 when stock reaches target. */
export function orderPacks(stock, target, packSize) {
  if (stock >= target || packSize <= 0) return 0;
  return Math.ceil((target - stock) / packSize);
}
