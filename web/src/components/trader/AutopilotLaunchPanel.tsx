import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  AlertCircle,
  ArrowRight,
  CircleDollarSign,
  Download,
  CheckCircle2,
  Copy,
  ExternalLink,
  KeyRound,
  Loader2,
  RefreshCw,
  ShieldCheck,
  Wallet,
  Zap,
} from 'lucide-react'
import { toast } from 'sonner'
import { api } from '../../lib/api'
import { buildDashboardPath, ROUTES } from '../../router/paths'
import type {
  AIModel,
  CurrentBeginnerWalletResponse,
  Exchange,
  ExchangeAccountState,
  TraderInfo,
} from '../../types'
import { HyperliquidWalletConnect } from '../common/HyperliquidWalletConnect'

type LaunchStepStatus = 'ready' | 'action' | 'blocked'

interface AutopilotLaunchPanelProps {
  models: AIModel[]
  exchanges: Exchange[]
  exchangeAccountStates: Record<string, ExchangeAccountState>
  traders?: TraderInfo[]
  isLoggedIn: boolean
  language: string
  onRefresh: () => Promise<void>
  onOpenClaw402Config?: () => void
  onOpenHyperliquidConfig?: () => void
}

const MIN_AI_FEE_USDC = 1
const MIN_TRADING_USDC = 12

function parseNumber(value?: string | number) {
  if (typeof value === 'number') return Number.isFinite(value) ? value : 0
  if (!value) return 0
  const parsed = Number(value.replace(/[,$\s]/g, ''))
  return Number.isFinite(parsed) ? parsed : 0
}

function shortAddress(address?: string) {
  if (!address) return '--'
  return `${address.slice(0, 6)}…${address.slice(-4)}`
}

function formatUSDC(value: number) {
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

async function copyText(value: string, label: string) {
  try {
    await navigator.clipboard.writeText(value)
    toast.success(`${label} copied`)
  } catch {
    toast.error('Copy failed')
  }
}

function BeginnerHyperliquidGuide({
  hasInjectedWallet,
  language,
}: {
  hasInjectedWallet: boolean
  language: string
}) {
  const isZh = language === 'zh'
  const isId = language === 'id'
  const t = (zh: string, id: string, en: string) =>
    isZh ? zh : isId ? id : en

  const steps = [
    {
      title: t('准备 EVM 钱包', 'Siapkan dompet EVM', 'Prepare an EVM wallet'),
      detail: hasInjectedWallet
        ? t(
            '检测到钱包扩展程序。解锁后请在下方连接。',
            'Ekstensi dompet terdeteksi. Buka kunci, lalu hubungkan di bawah.',
            'Wallet extension detected. Unlock it, then connect below.',
          )
        : t(
            '安装 Rabby 或 MetaMask，创建或导入钱包，然后返回这里。',
            'Instal Rabby atau MetaMask, buat atau impor dompet, lalu kembali ke sini.',
            'Install Rabby or MetaMask, create or import a wallet, then return here.',
          ),
      icon: Wallet,
    },
    {
      title: t('打开 Hyperliquid', 'Buka Hyperliquid', 'Open Hyperliquid'),
      detail: t(
        '在 Hyperliquid 上使用同一钱包。存入 USDC 作为交易保证金。',
        'Gunakan dompet yang sama di Hyperliquid. Setor USDC sebagai jaminan trading.',
        'Use the same wallet on Hyperliquid. Deposit USDC there as trading collateral.',
      ),
      icon: CircleDollarSign,
    },
    {
      title: t('授权 NOFX', 'Otorisasi NOFX', 'Authorize NOFX'),
      detail: t(
        '返回 NOFX，批准 Agent 和创建者费用。NOFX 存储 Agent 密钥，而非你的主钱包密钥。',
        'Kembali ke NOFX, setujui Agen dan biaya builder. NOFX menyimpan kunci Agen, bukan kunci dompet utama Anda.',
        'Back in NOFX, approve the Agent and builder fee. NOFX stores the Agent key, not your main wallet key.',
      ),
      icon: KeyRound,
    },
  ]

  return (
    <div className="rounded-xl border border-nofx-gold/20 bg-nofx-bg-lighter p-4">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="text-sm font-semibold text-nofx-text">
            {t('Hyperliquid 新手？', 'Baru di Hyperliquid?', 'New to Hyperliquid?')}
          </div>
          <p className="mt-1 text-xs leading-5 text-nofx-text-muted">
            {t(
              '如果你还没有交易钱包或从未使用过 Hyperliquid，从这里开始。',
              'Mulai di sini jika Anda belum memiliki dompet trading atau belum pernah menggunakan Hyperliquid.',
              'Start here if you do not have a trading wallet or have never used Hyperliquid before.',
            )}
          </p>
        </div>
        <div
          className={`w-fit rounded-full px-2.5 py-1 text-[11px] font-semibold ${
            hasInjectedWallet
              ? 'bg-nofx-success/10 text-nofx-success'
              : 'bg-nofx-gold/10 text-nofx-gold'
          }`}
        >
          {hasInjectedWallet
            ? t('已检测到钱包', 'Dompet terdeteksi', 'Wallet detected')
            : t('需要钱包', 'Dompet diperlukan', 'Wallet needed')}
        </div>
      </div>

      <div className="mt-4 grid gap-3">
        {steps.map((step, index) => {
          const Icon = step.icon
          return (
            <div key={step.title} className="flex gap-3">
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-nofx-gold/20 bg-nofx-bg-deeper text-nofx-gold">
                <Icon className="h-3.5 w-3.5" />
              </div>
              <div>
                <div className="text-sm font-semibold text-nofx-text">
                  {index + 1}. {step.title}
                </div>
                <p className="mt-0.5 text-xs leading-5 text-nofx-text-muted">
                  {step.detail}
                </p>
              </div>
            </div>
          )
        })}
      </div>

      <div className="mt-4 flex flex-wrap gap-2">
        {!hasInjectedWallet ? (
          <>
            <a
              href="https://rabby.io/"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 rounded-lg border border-nofx-gold/20 bg-nofx-bg-deeper px-3 py-2 text-xs font-semibold text-nofx-text hover:border-nofx-gold/40 hover:bg-nofx-bg"
            >
              <Download className="h-3.5 w-3.5" />
              {t('安装 Rabby', 'Instal Rabby', 'Install Rabby')}
            </a>
            <a
              href="https://metamask.io/download/"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 rounded-lg border border-nofx-gold/20 bg-nofx-bg-deeper px-3 py-2 text-xs font-semibold text-nofx-text hover:border-nofx-gold/40 hover:bg-nofx-bg"
            >
              <ExternalLink className="h-3.5 w-3.5" />
              {t('MetaMask', 'MetaMask', 'MetaMask')}
            </a>
          </>
        ) : null}
        <a
          href="https://app.hyperliquid.xyz/"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center gap-2 rounded-lg bg-nofx-gold px-3 py-2 text-xs font-bold text-white hover:bg-nofx-accent"
        >
          {t('打开 Hyperliquid', 'Buka Hyperliquid', 'Open Hyperliquid')}
          <ExternalLink className="h-3.5 w-3.5" />
        </a>
      </div>
    </div>
  )
}

export function AutopilotLaunchPanel({
  models,
  exchanges,
  exchangeAccountStates,
  traders = [],
  isLoggedIn,
  language,
  onRefresh,
  onOpenClaw402Config,
  onOpenHyperliquidConfig,
}: AutopilotLaunchPanelProps) {
  const navigate = useNavigate()
  const [wallet, setWallet] = useState<CurrentBeginnerWalletResponse | null>(
    null
  )
  const [walletLoading, setWalletLoading] = useState(false)
  const [launching, setLaunching] = useState(false)
  const [refreshing, setRefreshing] = useState(false)
  const [hasInjectedWallet, setHasInjectedWallet] = useState(false)
  const isZh = language === 'zh'
  const isId = language === 'id'
  const t = (zh: string, id: string, en: string) =>
    isZh ? zh : isId ? id : en

  useEffect(() => {
    setHasInjectedWallet(
      typeof window !== 'undefined' &&
        Boolean((window as Window & { ethereum?: unknown }).ethereum)
    )
  }, [])

  const claw402Model = useMemo(
    () =>
      models.find(
        (model) =>
          model.provider === 'claw402' &&
          model.enabled &&
          (model.has_api_key || model.apiKey || model.walletAddress)
      ) || null,
    [models]
  )

  const feeWalletAddress = claw402Model?.walletAddress || wallet?.address || ''
  const feeWalletBalance = parseNumber(
    claw402Model?.balanceUsdc || wallet?.balance_usdc
  )
  const feeReady =
    Boolean(feeWalletAddress) && feeWalletBalance >= MIN_AI_FEE_USDC

  const hyperliquidExchange = useMemo(
    () =>
      exchanges.find(
        (exchange) =>
          exchange.exchange_type === 'hyperliquid' &&
          exchange.enabled &&
          Boolean(exchange.hyperliquidWalletAddr) &&
          Boolean(exchange.hyperliquidBuilderApproved)
      ) || null,
    [exchanges]
  )

  const hyperliquidConnected = Boolean(hyperliquidExchange)
  const exchangeState = hyperliquidExchange
    ? exchangeAccountStates[hyperliquidExchange.id]
    : undefined
  const tradingBalance = parseNumber(
    exchangeState?.available_balance ?? exchangeState?.total_equity
  )
  const tradingBalanceReady =
    hyperliquidConnected &&
    exchangeState?.status === 'ok' &&
    tradingBalance >= MIN_TRADING_USDC

  const autopilotTrader = useMemo(
    () =>
      traders.find((trader) => trader.trader_name === 'NOFX Autopilot') ||
      traders.find((trader) =>
        (trader.strategy_name || '').toLowerCase().includes('claw402')
      ) ||
      null,
    [traders]
  )

  const allReady = feeReady && hyperliquidConnected && tradingBalanceReady

  const loadWallet = async () => {
    setWalletLoading(true)
    try {
      setWallet(await api.getCurrentBeginnerWallet())
    } catch {
      setWallet(null)
    } finally {
      setWalletLoading(false)
    }
  }

  useEffect(() => {
    void loadWallet()
  }, [])

  const refreshEverything = async () => {
    setRefreshing(true)
    try {
      await Promise.all([onRefresh(), loadWallet()])
    } finally {
      setRefreshing(false)
    }
  }

  const ensureClaw402Strategy = async () => {
    const strategies = await api.getStrategies()
    const existing =
      strategies.find(
        (strategy) =>
          strategy.is_active &&
          strategy.config?.ai_config?.coin_source?.source_type ===
            'vergex_signal'
      ) ||
      strategies.find((strategy) =>
        strategy.name.toLowerCase().includes('claw402')
      )

    if (existing) {
      if (!existing.is_active) {
        await api.activateStrategy(existing.id)
      }
      return existing.id
    }

    const config = await api.getDefaultStrategyConfig()
    const created = await api.createStrategy({
      name: 'NOFX Claw402 Auto Strategy',
      description:
        'Single built-in strategy: Claw402 board, per-symbol details, raw candles, then execution.',
      config,
    })
    if (created?.id) {
      await api.activateStrategy(created.id)
      return created.id
    }

    const refreshed = await api.getStrategies()
    const fallback = refreshed.find((strategy) =>
      strategy.name.toLowerCase().includes('claw402')
    )
    if (!fallback) throw new Error('Failed to create Claw402 strategy')
    await api.activateStrategy(fallback.id)
    return fallback.id
  }

  const launchAutopilot = async () => {
    if (!claw402Model || !hyperliquidExchange) return
    setLaunching(true)
    try {
      // Re-fetch the live trader list before deciding. The `traders` prop can be
      // stale (open tab serving an old snapshot), which would make us create a
      // duplicate "NOFX Autopilot" — paying the ~35s first-create cost again and
      // orphaning the dashboard onto a deleted id (the "API Not Found" 404). Match
      // the same way the autopilotTrader memo does, falling back to it on failure.
      let trader = autopilotTrader
      try {
        const fresh = await api.getTraders(true)
        trader =
          fresh.find((t) => t.trader_name === 'NOFX Autopilot') ||
          fresh.find((t) =>
            (t.strategy_name || '').toLowerCase().includes('claw402')
          ) ||
          trader
      } catch {
        // network hiccup — keep the prop-derived trader rather than risk a dup
      }
      if (!trader) {
        const strategyId = await ensureClaw402Strategy()
        trader = await api.createTrader({
          name: 'NOFX Autopilot',
          ai_model_id: claw402Model.id,
          exchange_id: hyperliquidExchange.id,
          strategy_id: strategyId,
          scan_interval_minutes: 5,
          is_cross_margin: true,
          show_in_competition: true,
          btc_eth_leverage: 10,
          altcoin_leverage: 10,
        })
      }
      if (!trader.is_running) {
        await api.startTrader(trader.trader_id)
      }
      await onRefresh()
      toast.success('NOFX Autopilot is running')
      navigate(buildDashboardPath(trader.trader_id))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Launch failed')
    } finally {
      setLaunching(false)
    }
  }

  const steps: Array<{
    title: string
    detail: string
    status: LaunchStepStatus
    meta?: string
    action?: JSX.Element
  }> = [
    {
      title: t('AI 费用钱包', 'Dompet biaya AI', 'AI fee wallet'),
      detail: t(
        '用 Base USDC 支付 Claw402.ai 的数据和模型调用费用。与交易保证金分开管理。',
        'Membayar panggilan data dan model Claw402.ai dengan Base USDC. Ini terpisah dari jaminan trading.',
        'Pays Claw402.ai data and model calls with Base USDC. This is separate from trading collateral.',
      ),
      status: feeReady ? 'ready' : 'action',
      meta: feeWalletAddress
        ? `${shortAddress(feeWalletAddress)} · ${formatUSDC(feeWalletBalance)} USDC`
        : t('需要 Base USDC 钱包', 'Dompet Base USDC diperlukan', 'Base USDC wallet required'),
      action: feeWalletAddress ? (
        <div className="flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={() => onOpenClaw402Config?.()}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-nofx-gold hover:text-nofx-accent"
          >
            <CircleDollarSign className="h-3.5 w-3.5" />
            {t('充值', 'Setor', 'Deposit')}
          </button>
          <button
            type="button"
            onClick={() => void copyText(feeWalletAddress, 'AI fee wallet')}
            className="inline-flex items-center gap-1.5 text-xs font-semibold text-nofx-gold hover:text-nofx-accent"
          >
            <Copy className="h-3.5 w-3.5" />
            {t('复制', 'Salin', 'Copy')}
          </button>
        </div>
      ) : (
        <button
          type="button"
          onClick={() => onOpenClaw402Config?.()}
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-nofx-gold hover:text-nofx-accent"
        >
          <ArrowRight className="h-3.5 w-3.5" />
          {t('打开', 'Buka', 'Open')}
        </button>
      ),
    },
    {
      title: t('Hyperliquid 交易钱包', 'Dompet trading Hyperliquid', 'Hyperliquid trading wallet'),
      detail: t(
        '连接 EVM 钱包，批准 NOFX Agent 和创建者费用，然后保存到 NOFX。',
        'Hubungkan dompet EVM, setujui Agen NOFX, setujui biaya builder, lalu simpan ke NOFX.',
        'Connect an EVM wallet, approve a NOFX Agent, approve the builder fee, then save it to NOFX.',
      ),
      status: hyperliquidConnected ? 'ready' : 'action',
      meta: hyperliquidExchange?.hyperliquidWalletAddr
        ? `${shortAddress(hyperliquidExchange.hyperliquidWalletAddr)} · ${t('已授权', 'Telah diotorisasi', 'authorized')}`
        : t('需要 Agent 和交易授权', 'Otorisasi Agen dan trading diperlukan', 'Agent and trading authorization required'),
      action: (
        <button
          type="button"
          onClick={() => onOpenHyperliquidConfig?.()}
          className="inline-flex items-center gap-1.5 text-xs font-semibold text-nofx-gold hover:text-nofx-accent"
        >
          <Wallet className="h-3.5 w-3.5" />
          {t('打开', 'Buka', 'Open')}
        </button>
      ),
    },
    {
      title: t('交易余额', 'Saldo trading', 'Trading balance'),
      detail: t(
        '将 USDC 存入 Hyperliquid。NOFX 将其用作 Claw402 自动驾驶策略的保证金。',
        'Setor USDC ke Hyperliquid. NOFX menggunakannya sebagai margin untuk strategi Claw402 Autopilot.',
        'Deposit USDC to Hyperliquid. NOFX uses it as margin for the Claw402 Autopilot strategy.',
      ),
      status: tradingBalanceReady
        ? 'ready'
        : hyperliquidConnected
          ? 'action'
          : 'blocked',
      meta: hyperliquidConnected
        ? `${formatUSDC(tradingBalance)} ${t('USDC 可用', 'USDC tersedia', 'USDC available')}`
        : t('请先连接 Hyperliquid', 'Hubungkan Hyperliquid terlebih dahulu', 'Connect Hyperliquid first'),
    },
    {
      title: t('NOFX 自动驾驶', 'Autopilot NOFX', 'NOFX Autopilot'),
      detail: t(
        '读取 Claw402 面板，获取 Signal Lab 和清算结构，用 K 线确认，仅在信号足够强时以 10 倍全仓交易。',
        'Membaca papan Claw402, mengambil Signal Lab dan struktur likuidasi, konfirmasi dengan candle, lalu trading full-size 10x hanya saat sinyal cukup kuat.',
        'Reads the Claw402 board, fetches Signal Lab and liquidation structure, confirms with candles, then trades full-size 10x only when the setup is strong enough.',
      ),
      status: allReady ? 'ready' : 'blocked',
      meta: autopilotTrader?.is_running
        ? t('运行中', 'Berjalan', 'Running')
        : autopilotTrader
          ? t('准备就绪', 'Siap dimulai', 'Ready to start')
          : t('配置完成后即可创建', 'Siap dibuat saat pengaturan selesai', 'Ready to create when setup is complete'),
    },
  ]

  const renderPrimaryAction = () => {
    if (!feeReady) {
      return (
        <button
          type="button"
          onClick={() => {
            if (onOpenClaw402Config) {
              onOpenClaw402Config()
            } else {
              navigate(ROUTES.welcome)
            }
          }}
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-nofx-gold px-4 py-3 text-sm font-bold text-white hover:bg-nofx-accent"
        >
          {t('打开 Claw402 钱包', 'Buka dompet Claw402', 'Open Claw402 wallet')}
          <ArrowRight className="h-4 w-4" />
        </button>
      )
    }

    if (!hyperliquidConnected) {
      return (
        <button
          type="button"
          onClick={() => {
            if (onOpenHyperliquidConfig) {
              onOpenHyperliquidConfig()
            } else {
              document
                .getElementById('hyperliquid-quick-connect')
                ?.scrollIntoView({ behavior: 'smooth', block: 'start' })
            }
          }}
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-nofx-gold px-4 py-3 text-sm font-bold text-white hover:bg-nofx-accent"
        >
          {t('连接 Hyperliquid', 'Hubungkan Hyperliquid', 'Connect Hyperliquid')}
          <ArrowRight className="h-4 w-4" />
        </button>
      )
    }

    if (!tradingBalanceReady) {
      return (
        <a
          href="https://app.hyperliquid.xyz/"
          target="_blank"
          rel="noreferrer"
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-nofx-gold px-4 py-3 text-sm font-bold text-white hover:bg-nofx-accent"
        >
          {t('在 Hyperliquid 上存入 USDC', 'Setor USDC di Hyperliquid', 'Deposit USDC on Hyperliquid')}
          <ExternalLink className="h-4 w-4" />
        </a>
      )
    }

    if (autopilotTrader?.is_running) {
      return (
        <button
          type="button"
          onClick={() =>
            navigate(buildDashboardPath(autopilotTrader.trader_id))
          }
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-nofx-success px-4 py-3 text-sm font-bold text-white hover:bg-nofx-success/80"
        >
          {t('打开仪表盘', 'Buka dashboard', 'Open dashboard')}
          <ArrowRight className="h-4 w-4" />
        </button>
      )
    }

    return (
      <button
        type="button"
        onClick={launchAutopilot}
        disabled={launching || !allReady}
        className="inline-flex items-center justify-center gap-2 rounded-lg bg-nofx-gold px-4 py-3 text-sm font-bold text-white hover:bg-nofx-accent disabled:cursor-not-allowed disabled:opacity-60"
      >
        {launching ? (
          <Loader2 className="h-4 w-4 animate-spin" />
        ) : (
          <Zap className="h-4 w-4" />
        )}
        {t('启动 NOFX 自动驾驶', 'Mulai Autopilot NOFX', 'Start NOFX Autopilot')}
      </button>
    )
  }

  return (
    <section className="overflow-hidden rounded-xl border border-nofx-gold/20 bg-nofx-bg-lighter">
      <div className="grid gap-0 xl:grid-cols-[1.05fr_0.95fr]">
        <div className="p-5 md:p-6">
          <div className="mb-5 flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div className="mb-2 inline-flex items-center gap-2 rounded-full border border-nofx-gold/25 bg-nofx-gold/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em] text-nofx-gold">
                <ShieldCheck className="h-3.5 w-3.5" />
                {t('引导启动', 'Panduan Peluncuran', 'Guided Launch')}
              </div>
              <h2 className="text-2xl font-bold tracking-tight text-nofx-text md:text-3xl">
                {t('几分钟内启动 NOFX 自动驾驶', 'Mulai Autopilot NOFX dalam hitungan menit', 'Start NOFX Autopilot in minutes')}
              </h2>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-nofx-text-muted">
                {t(
                  '一套策略，一条启动路径。充值 AI 费用钱包，授权 Hyperliquid，存入 USDC，然后运行 Claw402 自动驾驶。',
                  'Satu strategi, satu jalur peluncuran. Danai dompet biaya AI, otorisasi Hyperliquid, setor USDC, lalu jalankan Claw402 Autopilot.',
                  'One strategy, one launch path. Fund the AI fee wallet, authorize Hyperliquid, deposit USDC, then run the Claw402 Autopilot.',
                )}
              </p>
            </div>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => void refreshEverything()}
                disabled={refreshing || walletLoading}
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-nofx-gold/20 bg-nofx-bg-deeper px-3 py-2 text-xs font-semibold text-nofx-text-muted hover:text-nofx-text disabled:opacity-60"
              >
                <RefreshCw
                  className={`h-3.5 w-3.5 ${refreshing || walletLoading ? 'animate-spin' : ''}`}
                />
                {t('刷新', 'Segarkan', 'Refresh')}
              </button>
              {renderPrimaryAction()}
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-2">
            {steps.map((step, index) => (
              <div
                key={step.title}
                className="rounded-lg border border-nofx-gold/20 bg-nofx-bg p-4"
              >
                <div className="flex items-start gap-3">
                  <div
                    className={`mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border text-sm font-bold ${
                      step.status === 'ready'
                        ? 'border-nofx-success/30 bg-nofx-success/15 text-nofx-success'
                        : step.status === 'action'
                          ? 'border-nofx-gold/30 bg-nofx-gold/15 text-nofx-gold'
                          : 'border-nofx-gold/20 bg-nofx-bg-deeper text-nofx-text-muted'
                    }`}
                  >
                    {step.status === 'ready' ? (
                      <CheckCircle2 className="h-4 w-4" />
                    ) : step.status === 'action' ? (
                      index + 1
                    ) : (
                      <AlertCircle className="h-4 w-4" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <h3 className="font-semibold text-nofx-text">{step.title}</h3>
                      {step.action}
                    </div>
                    <p className="mt-1 text-xs leading-5 text-nofx-text-muted">
                      {step.detail}
                    </p>
                    <div className="mt-3 font-mono text-xs text-nofx-gold/90">
                      {step.meta}
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <aside className="border-t border-nofx-gold/20 bg-nofx-bg p-5 md:p-6 xl:border-l xl:border-t-0">
          <div className="mb-4 flex items-center gap-2 text-sm font-semibold text-nofx-text">
            <Wallet className="h-4 w-4 text-nofx-gold" />
            {t('Hyperliquid 设置', 'Pengaturan Hyperliquid', 'Hyperliquid setup')}
          </div>
          {hyperliquidConnected ? (
            <div className="rounded-lg border border-nofx-success/25 bg-nofx-success/10 p-4">
              <div className="flex items-center gap-2 text-sm font-semibold text-nofx-success">
                <CheckCircle2 className="h-4 w-4" />
                {t('交易授权已就绪', 'Otorisasi trading siap', 'Trading authorization is ready')}
              </div>
              <div className="mt-2 font-mono text-xs text-nofx-success/90">
                {shortAddress(hyperliquidExchange?.hyperliquidWalletAddr)}
              </div>
              <p className="mt-3 text-xs leading-5 text-nofx-text-muted">
                {t(
                  '资金保留在你的 Hyperliquid 账户中。NOFX 仅存储自动执行所需的授权 Agent 密钥。',
                  'Dana tetap di akun Hyperliquid Anda. NOFX hanya menyimpan kunci Agen yang diotorisasi yang diperlukan untuk eksekusi otomatis.',
                  'Funds stay in your Hyperliquid account. NOFX only stores the authorized Agent key required for automated execution.',
                )}
              </p>
            </div>
          ) : (
            <div className="space-y-4">
              <BeginnerHyperliquidGuide hasInjectedWallet={hasInjectedWallet} language={language} />
              <div id="hyperliquid-quick-connect">
                <HyperliquidWalletConnect
                  language={isZh ? 'zh' : 'en'}
                  isLoggedIn={isLoggedIn}
                  variant="inline"
                  onSaved={refreshEverything}
                />
              </div>
            </div>
          )}
        </aside>
      </div>
    </section>
  )
}
