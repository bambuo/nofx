import { useMemo } from 'react'
import type { FlowMarketItem } from '../../lib/api/data'

function td(lang: string, zh: string, id: string, en: string) {
  return lang === 'zh' ? zh : lang === 'id' ? id : en
}

interface FlowMarketsProps {
  items?: FlowMarketItem[]
  window?: string
  language?: string
}

function baseLabel(raw: string): string {
  return raw.toUpperCase().replace(/^XYZ:/, '').replace(/[-_]/g, '').replace(/(USDT|USDC|USD)$/, '')
}
function num(s: string): number {
  const n = parseFloat(s)
  return Number.isFinite(n) ? n : 0
}
function compact(n: number): string {
  const a = Math.abs(n)
  const sign = n < 0 ? '-' : '+'
  if (a >= 1e9) return `${sign}$${(a / 1e9).toFixed(2)}B`
  if (a >= 1e6) return `${sign}$${(a / 1e6).toFixed(2)}M`
  if (a >= 1e3) return `${sign}$${(a / 1e3).toFixed(1)}K`
  return `${sign}$${a.toFixed(0)}`
}

// Shared 5-column grid so the header, every row, and the legend line up exactly.
// symbol | net inflow | buy/sell split bar | trades | last price
const GRID = '64px 96px minmax(120px, 1fr) 80px 96px'

/**
 * FlowMarkets renders the Vergex net-flow ranking (real data from
 * GET /api/vergex/flow-markets via the trader's claw402 wallet). Each row shows
 * a market's net inflow over the window, a buy/sell split bar, trade count, and
 * latest price. Sorted by net inflow descending (the upstream ordering).
 */
export function FlowMarkets({ items, window = '1h', language = 'en' }: FlowMarketsProps) {
  const win = window.toUpperCase()
  const rows = useMemo(() => {
    if (!items || items.length === 0) return []
    const max = items.reduce((m, it) => Math.max(m, Math.abs(num(it.netFlow))), 1)
    return items.slice(0, 10).map((it) => {
      const buy = num(it.buyNotional)
      const sell = num(it.sellNotional)
      const tot = buy + sell || 1
      const net = num(it.netFlow)
      return {
        key: it.key || it.symbol,
        label: baseLabel(it.symbol),
        net,
        netStr: compact(net),
        buyPct: (buy / tot) * 100,
        widthPct: (Math.abs(net) / max) * 100,
        trades: it.trades,
        price: num(it.latestPrice),
      }
    })
  }, [items])

  if (rows.length === 0) {
    return <div className="tm-sc" style={{ padding: '12px 0' }}>{td(language, '无净流量数据（需要 claw402 付费）。', 'Tidak ada data net-flow (pembayaran claw402 diperlukan).', 'No net-flow data (claw402 payment required).')}</div>
  }

  return (
    <div className="tm-mono" style={{ fontSize: 11 }}>
      {/* column header — every number below is labeled by this row */}
      <div
        className="tm-sc"
        style={{
          display: 'grid',
          gridTemplateColumns: GRID,
          alignItems: 'end',
          gap: 12,
          paddingBottom: 4,
          borderBottom: '1px solid var(--tm-hair)',
          fontSize: 9,
        }}
      >
        <span>{td(language, '交易对', 'SIMBOL', 'SYMBOL')}</span>
        <span style={{ textAlign: 'right' }}>{win} {td(language, '净流量', 'NET', 'NET')}</span>
        <span>{td(language, '买入/卖出', 'BELI/JUAL', 'BUY/SELL')}</span>
        <span style={{ textAlign: 'right' }}>{td(language, '成交', 'TRANSAKSI', 'TRADES')}</span>
        <span style={{ textAlign: 'right' }}>{td(language, '价格', 'HARGA', 'PRICE')}</span>
      </div>

      {/* rows */}
      {rows.map((r) => (
        <div
          key={r.key}
          style={{
            display: 'grid',
            gridTemplateColumns: GRID,
            alignItems: 'center',
            gap: 12,
            height: 26,
            borderBottom: '1px solid var(--tm-hair)',
          }}
        >
          {/* symbol */}
          <span style={{ fontWeight: 600, color: 'var(--tm-ink)' }}>{r.label}</span>

          {/* net inflow figure (green = net buying / red = net selling) */}
          <span className={r.net >= 0 ? 'tm-up' : 'tm-dn'} style={{ textAlign: 'right', fontWeight: 600 }}>
            {r.netStr}
          </span>

          {/* buy/sell split bar — width encodes net-inflow magnitude (vs. the top
              market), the green/red split inside encodes buy vs. sell share.
              Green grows from the LEFT (buy), red fills the REST (sell). */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <div
              title={`buy ${r.buyPct.toFixed(0)}% / sell ${(100 - r.buyPct).toFixed(0)}%`}
              style={{
                position: 'relative',
                height: 8,
                flex: 1,
                minWidth: 40,
                background: 'var(--tm-hair)',
              }}
            >
              <div style={{ position: 'absolute', inset: 0, width: `${Math.max(4, r.widthPct)}%`, display: 'flex' }}>
                <div style={{ width: `${r.buyPct}%`, background: 'var(--tm-up)' }} />
                <div style={{ width: `${100 - r.buyPct}%`, background: 'var(--tm-dn)' }} />
              </div>
            </div>
            <span className="tm-sc" style={{ fontSize: 9, minWidth: 30, textAlign: 'right' }}>
              {r.buyPct.toFixed(0)}%
            </span>
          </div>

          {/* trade count */}
          <span style={{ textAlign: 'right', color: 'var(--tm-ink-2)' }}>
            {r.trades.toLocaleString('en-US')}
          </span>

          {/* last price */}
          <span style={{ textAlign: 'right', color: 'var(--tm-ink-2)' }}>
            ${r.price.toLocaleString('en-US', { maximumFractionDigits: 4 })}
          </span>
        </div>
      ))}

      {/* legend — explains every column */}
      <div className="tm-sc" style={{ marginTop: 8, fontSize: 9, lineHeight: 1.6 }}>
        {td(language, '净流入', 'arus masuk bersih', 'net inflow')} = {win} {td(language, '净买入', 'net pembelian', 'net buying')} · <span className="tm-up">{td(language, '绿色', 'hijau', 'green')}</span>/<span className="tm-dn">{td(language, '红色', 'merah', 'red')}</span> = {td(language, '买入/卖出比例', 'rasio beli/jual', 'buy/sell split')}
        {' · '}{td(language, '成交 = 笔数', 'transaksi = jumlah', 'trades = count')} · {td(language, '最新价 = 最后成交价', 'harga terakhir = harga transaksi terakhir', 'last price = last traded price')}
      </div>
    </div>
  )
}

export default FlowMarkets
