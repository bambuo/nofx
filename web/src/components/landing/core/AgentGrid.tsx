import { motion } from 'framer-motion'
import { TrendingUp, Layers, Zap, Hexagon, Crosshair } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../../contexts/AuthContext'
import type { Language } from '../../../i18n/translations'

const agentGridCopy: Record<Language, {
  eyebrow: string
  titleMain: string
  titleAccent: string
  description: string
  classLabel: string
  winLabel: string
  riskLabel: string
  initialize: string
  presets: {
    name: string
    class: string
    desc: string
    apy: string
    winRate: string
    risk: string
    color: string
    border: string
    bg_glow: string
    icon: typeof Zap
  }[]
}> = {
  en: {
    eyebrow: 'ASSET CLASS SELECT',
    titleMain: 'PROFESSIONAL',
    titleAccent: 'TRADERS',
    description:
      'CREATE TRADERS FOR US STOCKS, COMMODITIES, FX AND PRE-IPO MARKETS. DESCRIBE THE STRATEGY IN ONE SENTENCE.',
    classLabel: 'Class',
    winLabel: 'Win %',
    riskLabel: 'Risk',
    initialize: 'INITIALIZE',
  presets: [
  {
    name: 'ALPHA-1',
    class: 'US_STOCKS',
    desc: 'Large-cap momentum and breakout trading.',
    apy: '142%',
    winRate: '68%',
    risk: 'HIGH',
    color: 'text-nofx-gold',
    border: 'border-nofx-gold/50',
    bg_glow: 'shadow-sm',
    icon: Zap,
  },
  {
    name: 'BETA-X',
    class: 'MACRO_FX',
    desc: 'FX trend and macro regime allocation.',
    apy: '89%',
    winRate: '55%',
    risk: 'MED',
    color: 'text-nofx-accent',
    border: 'border-nofx-accent/30',
    bg_glow: 'shadow-sm',
    icon: TrendingUp,
  },
  {
    name: 'GAMMA-RAY',
    class: 'PRE_IPO',
    desc: 'Private-market momentum basket engine.',
    apy: '24%',
    winRate: '99%',
    risk: 'LOW',
    color: 'text-nofx-text',
    border: 'border-nofx-gold/20',
    bg_glow: 'shadow-sm',
    icon: Layers,
  },
    ],
  },
  zh: {
    eyebrow: '选择资产类别',
    titleMain: '专业',
    titleAccent: '交易员',
    description:
      '为美股、商品、外汇和 PRE-IPO 市场创建交易员。用一句话描述你的策略即可。',
    classLabel: '类别',
    winLabel: '胜率',
    riskLabel: '风险',
    initialize: '初始化',
    presets: [
      {
        name: 'ALPHA-1',
        class: '美股',
        desc: '大盘股动量与突破交易。',
        apy: '142%',
        winRate: '68%',
        risk: '高',
        color: 'text-nofx-gold',
        border: 'border-nofx-gold/50',
        bg_glow: 'shadow-sm',
        icon: Zap,
      },
      {
        name: 'BETA-X',
        class: '宏观外汇',
        desc: '外汇趋势与宏观状态配置。',
        apy: '89%',
        winRate: '55%',
        risk: '中',
        color: 'text-nofx-accent',
        border: 'border-nofx-accent/30',
        bg_glow: 'shadow-sm',
        icon: TrendingUp,
      },
      {
        name: 'GAMMA-RAY',
        class: 'PRE-IPO',
        desc: '私募市场动量篮子引擎。',
        apy: '24%',
        winRate: '99%',
        risk: '低',
        color: 'text-nofx-text',
        border: 'border-nofx-gold/20',
        bg_glow: 'shadow-sm',
        icon: Layers,
      },
    ],
  },
  id: {
    eyebrow: 'PILIH KELAS ASET',
    titleMain: 'TRADER',
    titleAccent: 'PROFESIONAL',
    description:
      'BUAT TRADER UNTUK SAHAM AS, KOMODITAS, FX, DAN PASAR PRE-IPO. JELASKAN STRATEGI DALAM SATU KALIMAT.',
    classLabel: 'Kelas',
    winLabel: 'Menang %',
    riskLabel: 'Risiko',
    initialize: 'INISIALISASI',
    presets: [
      {
        name: 'ALPHA-1',
        class: 'SAHAM_AS',
        desc: 'Trading momentum dan breakout saham besar.',
        apy: '142%',
        winRate: '68%',
        risk: 'TINGGI',
        color: 'text-nofx-gold',
        border: 'border-nofx-gold/50',
        bg_glow: 'shadow-sm',
        icon: Zap,
      },
      {
        name: 'BETA-X',
        class: 'MAKRO_FX',
        desc: 'Tren FX dan alokasi rezim makro.',
        apy: '89%',
        winRate: '55%',
        risk: 'SEDANG',
        color: 'text-nofx-accent',
        border: 'border-nofx-accent/30',
        bg_glow: 'shadow-sm',
        icon: TrendingUp,
      },
      {
        name: 'GAMMA-RAY',
        class: 'PRE_IPO',
        desc: 'Mesin basket momentum pasar privat.',
        apy: '24%',
        winRate: '99%',
        risk: 'RENDAH',
        color: 'text-nofx-text',
        border: 'border-nofx-gold/20',
        bg_glow: 'shadow-sm',
        icon: Layers,
      },
    ],
  },
}

export default function AgentGrid({ language }: { language: Language }) {
  const { user } = useAuth()
  const navigate = useNavigate()
  const copy = agentGridCopy[language]

  const handleInitialize = () => {
    if (user) {
      navigate('/strategy')
    } else {
      navigate('/login')
    }
  }

  return (
    <section
      id="market-scanner"
      className="py-16 md:py-24 bg-nofx-bg relative overflow-hidden"
    >
      {/* Background Details */}
      <div className="absolute top-0 right-0 p-10 opacity-10 pointer-events-none">
        <Hexagon className="w-64 h-64 text-nofx-text-muted" strokeWidth={0.5} />
      </div>

      <div className="max-w-7xl mx-auto px-6 relative z-10">
        <div className="flex flex-col md:flex-row justify-between items-end mb-10 md:mb-16 gap-6">
          <div>
            <div className="flex items-center gap-2 text-nofx-gold font-mono text-xs mb-2 tracking-widest uppercase">
              <Crosshair className="w-4 h-4" /> {copy.eyebrow}
            </div>
            <h2 className="text-4xl md:text-5xl font-black text-nofx-text uppercase tracking-tighter">
              {copy.titleMain}{' '}
              <span className="text-nofx-gold">
                {copy.titleAccent}
              </span>
            </h2>
          </div>
          <div className="font-mono text-right text-xs text-nofx-text-muted max-w-xs">
            {copy.description}
          </div>
        </div>

        {/* Grid Container - Removing scroll tracking for stability test */}
        <div className="flex flex-row md:grid md:grid-cols-3 gap-4 md:gap-8 overflow-x-auto md:overflow-visible pb-12 md:pb-0 snap-x snap-mandatory -mx-6 px-6 md:mx-0 md:px-0 scrollbar-hide">
          {copy.presets.map((preset, i) => {
            const Icon = preset.icon

            return (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.1 }}
                className={`group relative bg-nofx-bg-lighter backdrop-blur-xl border ${preset.border} overflow-hidden transition-all duration-300 min-w-[85vw] md:min-w-0 snap-center shrink-0 rounded-xl md:rounded-none`}
              >
                {/* Top "Hinge" decoration */}
                <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-nofx-text/10 to-transparent"></div>

                <div className="p-8 relative z-10">
                  {/* Header */}
                  <div className="flex justify-between items-start mb-6">
                    <div className="p-3 bg-nofx-bg-deeper rounded border border-[rgba(26,24,19,0.14)]">
                      <Icon className={`w-8 h-8 ${preset.color}`} />
                    </div>
                    <div className="text-right">
                      <div className="text-[10px] font-mono text-nofx-text-muted uppercase">
                        {copy.classLabel}
                      </div>
                      <div
                        className={`font-bold font-mono tracking-wider ${preset.color}`}
                      >
                        {preset.class}
                      </div>
                    </div>
                  </div>

                  {/* Name & Desc */}
                  <h3 className="text-3xl font-bold text-nofx-text mb-2 tracking-tight group-hover:text-nofx-accent transition-colors">
                    {preset.name}
                  </h3>
                  <p className="text-nofx-text-muted text-sm mb-8 leading-relaxed h-10">
                    {preset.desc}
                  </p>

                  {/* Stats Grid */}
                  <div className="grid grid-cols-3 gap-px bg-[rgba(26,24,19,0.14)] border border-[rgba(26,24,19,0.14)] rounded overflow-hidden mb-8">
                    <div className="bg-nofx-bg-deeper p-3 text-center group-hover:bg-nofx-bg transition-colors">
                      <div className="text-[10px] text-nofx-text-muted uppercase font-mono mb-1">
                        APY
                      </div>
                      <div className="text-nofx-success font-bold">
                        {preset.apy}
                      </div>
                    </div>
                    <div className="bg-nofx-bg-deeper p-3 text-center group-hover:bg-nofx-bg transition-colors">
                      <div className="text-[10px] text-nofx-text-muted uppercase font-mono mb-1">
                        {copy.winLabel}
                      </div>
                      <div className="text-nofx-text font-bold">
                        {preset.winRate}
                      </div>
                    </div>
                    <div className="bg-nofx-bg-deeper p-3 text-center group-hover:bg-nofx-bg transition-colors">
                      <div className="text-[10px] text-nofx-text-muted uppercase font-mono mb-1">
                        {copy.riskLabel}
                      </div>
                      <div className={`${preset.color} font-bold`}>
                        {preset.risk}
                      </div>
                    </div>
                  </div>

                  {/* Action Btn */}
                  <button
                    onClick={handleInitialize}
                    className={`w-full py-4 text-xs font-bold font-mono uppercase tracking-[0.2em] border border-[rgba(26,24,19,0.14)] hover:border-${preset.color === 'text-nofx-gold' ? 'nofx-gold' : 'nofx-text'} hover:bg-nofx-text/5 transition-all flex items-center justify-center gap-2 group-hover:text-nofx-text cursor-pointer text-nofx-text`}
                  >
                    <span className={preset.color}>[</span> {copy.initialize}{' '}
                    <span className={preset.color}>]</span>
                  </button>
                </div>

                {/* Decorative Background Elements */}
                <div className="absolute -right-10 -bottom-10 w-40 h-40 bg-gradient-to-br from-nofx-text/5 to-transparent rounded-full blur-2xl group-hover:opacity-50 transition-opacity opacity-20"></div>
              </motion.div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
