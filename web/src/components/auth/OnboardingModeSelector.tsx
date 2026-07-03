import type { UserMode } from '../../lib/onboarding'
import type { Language } from '../../i18n/translations'

interface OnboardingModeSelectorProps {
  language: string
  mode: UserMode
  onChange: (mode: UserMode) => void
}

type OptionEntry = {
  title: Record<Language, string>
  badge?: Record<Language, string>
  description: Record<Language, string>
}

const optionDefs: Record<UserMode, OptionEntry> = {
  beginner: {
    title: {
      en: 'Beginner Mode',
      zh: '新手模式',
      id: 'Mode Pemula',
    },
    badge: {
      en: 'Recommended',
      zh: '推荐',
      id: 'Direkomendasikan',
    },
    description: {
      en: 'Generate a Base wallet automatically and start with Claw402 + GLM by default.',
      zh: '自动生成 Base 钱包，默认使用 Claw402 + GLM 起步。',
      id: 'Buat dompet Base secara otomatis dan mulai dengan Claw402 + GLM secara default.',
    },
  },
  advanced: {
    title: {
      en: 'Advanced Mode',
      zh: '高级模式',
      id: 'Mode Lanjutan',
    },
    description: {
      en: 'Keep the full manual flow and configure models, wallets, and exchanges yourself.',
      zh: '保留完整的手动流程，自行配置模型、钱包和交易所。',
      id: 'Pertahankan alur manual penuh dan konfigurasi model, dompet, serta bursa sendiri.',
    },
  },
}

const label: Record<Language, string> = {
  en: 'Experience',
  zh: '体验模式',
  id: 'Pengalaman',
}

export function OnboardingModeSelector({
  language,
  mode,
  onChange,
}: OnboardingModeSelectorProps) {
  const lang = (language === 'zh' || language === 'id' ? language : 'en') as Language

  return (
    <div className="space-y-2">
      <div className="text-xs font-medium text-nofx-text-muted">
        {label[lang]}
      </div>
      <div className="grid grid-cols-1 gap-2">
        {(['beginner', 'advanced'] as UserMode[]).map((id) => {
          const def = optionDefs[id]
          const selected = id === mode
          return (
            <button
              key={id}
              type="button"
              onClick={() => onChange(id)}
              className={`w-full rounded-xl border px-4 py-3 text-left transition-all ${
                selected
                  ? 'border-nofx-gold/60 bg-nofx-gold/10'
                  : 'border-[rgba(26,24,19,0.14)] bg-nofx-bg-lighter hover:border-nofx-gold/40'
              }`}
            >
              <div className="flex items-center gap-2 text-sm font-semibold text-nofx-text">
                <span>{def.title[lang]}</span>
                {def.badge && (
                  <span className="rounded-full bg-nofx-gold px-2 py-0.5 text-[10px] font-bold uppercase tracking-wide text-nofx-bg">
                    {def.badge[lang]}
                  </span>
                )}
              </div>
              <p className="mt-1 text-xs leading-5 text-nofx-text-muted">
                {def.description[lang]}
              </p>
            </button>
          )
        })}
      </div>
    </div>
  )
}
