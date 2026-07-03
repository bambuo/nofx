import { Link } from 'react-router-dom'
import {
  ArrowRight,
  CheckCircle2,
  CircleDollarSign,
  Download,
  ExternalLink,
  KeyRound,
  ShieldCheck,
  Wallet,
  Zap,
} from 'lucide-react'
import { ROUTES } from '../../router/paths'
import { useLanguage } from '../../contexts/LanguageContext'

function t3(lang: string, zh: string, id: string, en: string) {
  return lang === 'zh' ? zh : lang === 'id' ? id : en
}

const setupStepsZh = [
  {
    title: '创建 NOFX 账户',
    detail:
      '你的账户将自动驾驶策略、钱包授权状态和交易仪表盘集中在一处。',
    icon: KeyRound,
    action: '创建账户',
    to: ROUTES.register,
  },
  {
    title: '充值 AI 费用钱包',
    detail:
      'NOFX 会为 Claw402.ai 的数据和模型调用准备一个 Base USDC 钱包。该钱包与交易保证金分开管理。',
    icon: CircleDollarSign,
    action: '扫码充值',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=claw402`,
  },
  {
    title: '授权 Hyperliquid',
    detail:
      '连接你的交易钱包，批准 NOFX Agent 和创建者费用。资金保留在你的 Hyperliquid 账户中。',
    icon: Wallet,
    action: '连接交易所',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=hyperliquid`,
  },
  {
    title: '存入交易 USDC',
    detail:
      '在 Hyperliquid 上添加 USDC，然后启动 NOFX 自动驾驶。策略会自动创建并启动。',
    icon: Zap,
    action: '打开 Hyperliquid',
    href: 'https://app.hyperliquid.xyz/',
  },
]

const setupStepsId = [
  {
    title: 'Buat akun NOFX',
    detail:
      'Akun Anda menyimpan konfigurasi Autopilot, status otorisasi dompet, dan dashboard trading di satu tempat.',
    icon: KeyRound,
    action: 'Buat akun',
    to: ROUTES.register,
  },
  {
    title: 'Danai dompet biaya AI',
    detail:
      'NOFX menyiapkan dompet Base USDC untuk panggilan data dan model Claw402.ai. Dompet ini terpisah dari jaminan trading.',
    icon: CircleDollarSign,
    action: 'Buka QR setoran',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=claw402`,
  },
  {
    title: 'Otorisasi Hyperliquid',
    detail:
      'Hubungkan dompet trading Anda, setujui Agen NOFX, dan setujui biaya builder. Dana tetap di akun Hyperliquid Anda.',
    icon: Wallet,
    action: 'Hubungkan bursa',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=hyperliquid`,
  },
  {
    title: 'Setor USDC trading',
    detail:
      'Tambahkan USDC di Hyperliquid, lalu mulai Autopilot NOFX. Strategi dibuat dan diluncurkan secara otomatis.',
    icon: Zap,
    action: 'Buka Hyperliquid',
    href: 'https://app.hyperliquid.xyz/',
  },
]

const setupStepsEn = [
  {
    title: 'Create your NOFX account',
    detail:
      'Your account keeps the Autopilot configuration, wallet authorization state, and trading dashboard in one place.',
    icon: KeyRound,
    action: 'Create account',
    to: ROUTES.register,
  },
  {
    title: 'Fund the AI fee wallet',
    detail:
      'NOFX prepares a Base USDC wallet for Claw402.ai data and model calls. This wallet is separate from trading collateral.',
    icon: CircleDollarSign,
    action: 'Open deposit QR',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=claw402`,
  },
  {
    title: 'Authorize Hyperliquid',
    detail:
      'Connect your trading wallet, approve the NOFX Agent, and approve the builder fee. Funds remain in your Hyperliquid account.',
    icon: Wallet,
    action: 'Connect exchange',
    to: ROUTES.login,
    returnUrl: `${ROUTES.traders}?setup=hyperliquid`,
  },
  {
    title: 'Deposit trading USDC',
    detail:
      'Add USDC on Hyperliquid, then start NOFX Autopilot. The strategy is created and launched automatically.',
    icon: Zap,
    action: 'Open Hyperliquid',
    href: 'https://app.hyperliquid.xyz/',
  },
]

const pipeline = [
  'Read the live Claw402.ai board, with US stocks prioritized before crypto.',
  'Fetch Signal Lab and cost/liquidation heatmap details for each candidate.',
  'Confirm with raw OHLCV candles, then trade full-size 10x only when the setup is strong enough.',
]

const pipelineZh = [
  '读取实时 Claw402.ai 面板，美股优先于加密货币。',
  '获取每个候选品种的 Signal Lab 和成本/清算热力图详情。',
  '用原始 OHLCV K 线确认，仅在信号足够强时以 10 倍全仓交易。',
]

const pipelineId = [
  'Baca papan Claw402.ai langsung, dengan saham AS diprioritaskan sebelum kripto.',
  'Ambil Signal Lab dan detail peta panas biaya/likuidasi untuk setiap kandidat.',
  'Konfirmasi dengan candle OHLCV mentah, lalu trading full-size 10x hanya saat sinyal cukup kuat.',
]

export function TraderLaunchGuestPage() {
  const { language } = useLanguage()
  const setupSteps = language === 'zh' ? setupStepsZh : language === 'id' ? setupStepsId : setupStepsEn
  const steps = language === 'zh' ? pipelineZh : language === 'id' ? pipelineId : pipeline

  return (
    <div className="min-h-[calc(100vh-4rem)] overflow-hidden bg-nofx-bg px-4 py-10 md:px-8">
      <div className="mx-auto flex w-full max-w-7xl flex-col gap-8">
        <section className="grid gap-8 rounded-2xl border border-nofx-gold/20 bg-nofx-bg-lighter p-6 md:p-8 xl:grid-cols-[1.02fr_0.98fr]">
          <div className="flex flex-col justify-center">
            <div className="mb-5 inline-flex w-fit items-center gap-2 rounded-full border border-nofx-gold/25 bg-nofx-gold/10 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.2em] text-nofx-gold">
              <ShieldCheck className="h-3.5 w-3.5" />
              NOFX Autopilot
            </div>
            <h1 className="max-w-3xl text-4xl font-bold tracking-tight text-nofx-text md:text-5xl">
              {t3(language, '一套策略，四个步骤，然后开始交易。', 'Satu strategi. Empat langkah. Lalu mulai trading.', 'One strategy. Four setup steps. Then it trades.')}
            </h1>
            <p className="mt-5 max-w-2xl text-base leading-7 text-nofx-text-muted">
              {t3(language, 'NOFX 运行单一的 Claw402 驱动策略：面板、每个市场的详情、清算结构、K 线、执行。无需策略选择器，无需手动选择交易品种。', 'NOFX menjalankan satu strategi berbasis Claw402: board, detail per pasar, struktur likuidasi, candle, eksekusi. Tidak perlu pemilih strategi, tidak perlu memilih simbol manual.', 'NOFX runs a single Claw402-driven strategy: board, per-market details, liquidation structure, candles, execution. No strategy picker, no manual symbol picking required.')}
            </p>
            <div className="mt-7 flex flex-col gap-3 sm:flex-row">
              <Link
                to={ROUTES.login}
                onClick={() =>
                  sessionStorage.setItem(
                    'returnUrl',
                    `${ROUTES.traders}?setup=claw402`
                  )
                }
                className="inline-flex items-center justify-center gap-2 rounded-xl bg-nofx-gold px-5 py-3 text-sm font-bold text-white transition hover:bg-nofx-gold/90"
              >
                {t3(language, '开始设置', 'Mulai pengaturan', 'Start setup')}
                <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                to={ROUTES.register}
                className="inline-flex items-center justify-center rounded-xl border border-nofx-gold/20 bg-nofx-bg-deeper px-5 py-3 text-sm font-semibold text-nofx-text transition hover:border-nofx-gold/40 hover:bg-nofx-bg-deeper"
              >
                {t3(language, '创建账户', 'Buat akun', 'Create account')}
              </Link>
            </div>
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            {setupSteps.map((step, index) => {
              const Icon = step.icon
              const cardClass =
                'group rounded-xl border border-nofx-gold/20 bg-nofx-bg-deeper p-4 text-left transition hover:border-nofx-gold/35 hover:bg-nofx-gold/[0.06]'
              const content = (
                <>
                  <div className="mb-4 flex items-center justify-between">
                    <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-nofx-gold/20 bg-nofx-gold/10 text-nofx-gold">
                      <Icon className="h-4 w-4" />
                    </div>
                    <span className="font-mono text-xs text-nofx-text-muted">
                      0{index + 1}
                    </span>
                  </div>
                  <h2 className="text-base font-semibold text-nofx-text">
                    {step.title}
                  </h2>
                  <p className="mt-2 text-sm leading-6 text-nofx-text-muted">
                    {step.detail}
                  </p>
                  <div className="mt-4 inline-flex items-center gap-2 text-xs font-bold text-nofx-gold transition group-hover:text-nofx-gold/80">
                    {step.action}
                    {step.href ? (
                      <ExternalLink className="h-3.5 w-3.5" />
                    ) : (
                      <ArrowRight className="h-3.5 w-3.5" />
                    )}
                  </div>
                </>
              )

              if (step.href) {
                return (
                  <a
                    key={step.title}
                    href={step.href}
                    target="_blank"
                    rel="noreferrer"
                    className={cardClass}
                  >
                    {content}
                  </a>
                )
              }

              return (
                <Link
                  key={step.title}
                  to={step.to || ROUTES.login}
                  onClick={() => {
                    if (step.returnUrl) {
                      sessionStorage.setItem('returnUrl', step.returnUrl)
                    }
                  }}
                  className={cardClass}
                >
                  {content}
                </Link>
              )
            })}
          </div>
        </section>

        <section className="grid gap-5 rounded-2xl border border-nofx-gold/20 bg-nofx-bg-lighter p-5 md:grid-cols-[0.78fr_1.22fr] md:p-6">
          <div>
            <div className="text-sm font-semibold uppercase tracking-[0.18em] text-nofx-gold">
              {t3(language, '还没有交易钱包？', 'Belum punya dompet trading?', 'No trading wallet yet?')}
            </div>
            <p className="mt-3 text-sm leading-6 text-nofx-text-muted">
              {t3(language, 'NOFX 不需要你的主钱包私钥。安装或解锁 EVM 钱包，用 USDC 充值 Hyperliquid，然后在登录后授权 NOFX Agent。', 'NOFX tidak memerlukan kunci privat dompet utama Anda. Instal atau buka kunci dompet EVM, danai Hyperliquid dengan USDC, lalu otorisasi Agen NOFX setelah masuk.', 'NOFX does not need your main-wallet private key. Install or unlock an EVM wallet, fund Hyperliquid with USDC, then authorize the NOFX Agent after sign-in.')}
            </p>
          </div>
          <div className="grid gap-3 lg:grid-cols-3">
            <a
              href="https://rabby.io/"
              target="_blank"
              rel="noreferrer"
              className="group rounded-xl border border-nofx-gold/20 bg-nofx-bg-deeper p-4 transition hover:border-nofx-gold/30 hover:bg-nofx-gold/[0.06]"
            >
              <Download className="mb-3 h-4 w-4 text-nofx-gold" />
              <div className="font-semibold text-nofx-text">{t3(language, '安装 Rabby', 'Instal Rabby', 'Install Rabby')}</div>
              <p className="mt-2 text-sm leading-6 text-nofx-text-muted">
                {t3(language, '在连接 Hyperliquid 前创建或导入 EVM 钱包。', 'Buat atau impor dompet EVM sebelum menghubungkan ke Hyperliquid.', 'Create or import an EVM wallet before connecting to Hyperliquid.')}
              </p>
            </a>
            <a
              href="https://metamask.io/download/"
              target="_blank"
              rel="noreferrer"
              className="group rounded-xl border border-nofx-gold/20 bg-nofx-bg-deeper p-4 transition hover:border-nofx-gold/30 hover:bg-nofx-gold/[0.06]"
            >
              <ExternalLink className="mb-3 h-4 w-4 text-nofx-gold" />
              <div className="font-semibold text-nofx-text">MetaMask</div>
              <p className="mt-2 text-sm leading-6 text-nofx-text-muted">
                {t3(language, '已经在用 MetaMask？解锁后继续在 NOFX 中设置。', 'Sudah menggunakan MetaMask? Buka kunci, lalu lanjutkan pengaturan di NOFX.', 'Already use MetaMask? Unlock it, then continue setup inside NOFX.')}
              </p>
            </a>
            <a
              href="https://app.hyperliquid.xyz/"
              target="_blank"
              rel="noreferrer"
              className="group rounded-xl border border-nofx-gold/20 bg-nofx-gold/10 p-4 transition hover:bg-nofx-gold/15"
            >
              <ExternalLink className="mb-3 h-4 w-4 text-nofx-gold" />
              <div className="font-semibold text-nofx-text">{t3(language, '打开 Hyperliquid', 'Buka Hyperliquid', 'Open Hyperliquid')}</div>
              <p className="mt-2 text-sm leading-6 text-nofx-text-muted">
                {t3(language, '在那里存入 USDC。交易资金保留在你的 Hyperliquid 账户中。', 'Setor USDC di sana. Dana trading tetap di akun Hyperliquid Anda.', 'Deposit USDC there. Trading funds stay in your Hyperliquid account.')}
              </p>
            </a>
          </div>
        </section>

        <section className="grid gap-4 rounded-2xl border border-nofx-gold/20 bg-nofx-bg-lighter p-5 md:grid-cols-[0.72fr_1.28fr] md:p-6">
          <div>
            <div className="text-sm font-semibold uppercase tracking-[0.18em] text-nofx-gold">
              {t3(language, '启动后运行什么', 'Apa yang berjalan setelah peluncuran', 'What runs after launch')}
            </div>
            <p className="mt-3 text-sm leading-6 text-nofx-text-muted">
              {t3(language, '每个周期运行相同的生产环境路径。你只需要充值、授权和启动。', 'Jalur produksi yang sama berjalan setiap siklus. Antarmuka hanya meminta Anda mendanai, mengotorisasi, dan memulai.', 'The same production path runs every cycle. The interface only asks you to fund, authorize, and start.')}
            </p>
          </div>
          <div className="grid gap-3 lg:grid-cols-3">
            {steps.map((item) => (
              <div
                key={item}
                className="flex gap-3 rounded-xl border border-nofx-gold/20 bg-nofx-bg-deeper p-4"
              >
                <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-nofx-success" />
                <p className="text-sm leading-6 text-nofx-text">{item}</p>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  )
}
