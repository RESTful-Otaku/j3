// Lightweight translation layer for the 11 supported locales. The settings
// screen stores the preference; shared labels use this dictionary.
export type Locale = 'en' | 'fr' | 'de' | 'es' | 'pt' | 'hi' | 'pa' | 'ar' | 'ja' | 'ko' | 'zh';

export const LOCALES: { id: Locale; label: string; dir?: 'rtl' }[] = [
  { id: 'en', label: 'English' }, { id: 'fr', label: 'Français' },
  { id: 'de', label: 'Deutsch' }, { id: 'es', label: 'Español' },
  { id: 'pt', label: 'Português' }, { id: 'hi', label: 'हिन्दी' },
  { id: 'pa', label: 'ਪੰਜਾਬੀ' }, { id: 'ar', label: 'العربية', dir: 'rtl' },
  { id: 'ja', label: '日本語' }, { id: 'ko', label: '한국어' }, { id: 'zh', label: '中文' },
];

type Key = 'dashboard' | 'setup' | 'scoreboard' | 'public' | 'brackets' | 'tables' |
  'pools' | 'control' | 'save' | 'cancel' | 'confirm' | 'participants' |
  'results' | 'schedule' | 'medals' | 'upcoming' | 'live';

const DICTIONARY: Record<Locale, Record<Key, string>> = {
  en: { dashboard: 'Dashboard', setup: 'Setup', scoreboard: 'Scoreboard', public: 'Public', brackets: 'Brackets', tables: 'Tables', pools: 'Pools', control: 'Control', save: 'Save', cancel: 'Cancel', confirm: 'Confirm', participants: 'Participants', results: 'Results', schedule: 'Schedule', medals: 'Medals', upcoming: 'Upcoming', live: 'Live' },
  fr: { dashboard: 'Tableau de bord', setup: 'Configuration', scoreboard: 'Tableau de score', public: 'Public', brackets: 'Tableaux', tables: 'Classements', pools: 'Poules', control: 'Contrôle', save: 'Enregistrer', cancel: 'Annuler', confirm: 'Confirmer', participants: 'Participants', results: 'Résultats', schedule: 'Programme', medals: 'Médailles', upcoming: 'À venir', live: 'En direct' },
  de: { dashboard: 'Übersicht', setup: 'Einrichtung', scoreboard: 'Anzeigetafel', public: 'Öffentlich', brackets: 'Turnierbaum', tables: 'Tabellen', pools: 'Gruppen', control: 'Steuerung', save: 'Speichern', cancel: 'Abbrechen', confirm: 'Bestätigen', participants: 'Teilnehmer', results: 'Ergebnisse', schedule: 'Zeitplan', medals: 'Medaillen', upcoming: 'Als Nächstes', live: 'Live' },
  es: { dashboard: 'Panel', setup: 'Configuración', scoreboard: 'Marcador', public: 'Público', brackets: 'Cuadros', tables: 'Tablas', pools: 'Grupos', control: 'Control', save: 'Guardar', cancel: 'Cancelar', confirm: 'Confirmar', participants: 'Participantes', results: 'Resultados', schedule: 'Horario', medals: 'Medallas', upcoming: 'Próximos', live: 'En directo' },
  pt: { dashboard: 'Painel', setup: 'Configuração', scoreboard: 'Placar', public: 'Público', brackets: 'Chaves', tables: 'Tabelas', pools: 'Grupos', control: 'Controlo', save: 'Guardar', cancel: 'Cancelar', confirm: 'Confirmar', participants: 'Participantes', results: 'Resultados', schedule: 'Calendário', medals: 'Medalhas', upcoming: 'A seguir', live: 'Ao vivo' },
  hi: { dashboard: 'डैशबोर्ड', setup: 'सेटअप', scoreboard: 'स्कोरबोर्ड', public: 'सार्वजनिक', brackets: 'ब्रैकेट', tables: 'तालिकाएँ', pools: 'पूल', control: 'नियंत्रण', save: 'सहेजें', cancel: 'रद्द करें', confirm: 'पुष्टि करें', participants: 'प्रतिभागी', results: 'परिणाम', schedule: 'कार्यक्रम', medals: 'पदक', upcoming: 'आगामी', live: 'लाइव' },
  pa: { dashboard: 'ਡੈਸ਼ਬੋਰਡ', setup: 'ਸੈੱਟਅੱਪ', scoreboard: 'ਸਕੋਰਬੋਰਡ', public: 'ਜਨਤਕ', brackets: 'ਬ੍ਰੈਕਟ', tables: 'ਟੇਬਲ', pools: 'ਪੂਲ', control: 'ਕੰਟਰੋਲ', save: 'ਸੇਵ ਕਰੋ', cancel: 'ਰੱਦ ਕਰੋ', confirm: 'ਪੁਸ਼ਟੀ ਕਰੋ', participants: 'ਭਾਗੀਦਾਰ', results: 'ਨਤੀਜੇ', schedule: 'ਸਮਾਂ-ਸਾਰਣੀ', medals: 'ਮੈਡਲ', upcoming: 'ਅਗਲਾ', live: 'ਲਾਈਵ' },
  ar: { dashboard: 'لوحة التحكم', setup: 'الإعداد', scoreboard: 'لوحة النتائج', public: 'عام', brackets: 'الجداول', tables: 'الترتيب', pools: 'المجموعات', control: 'التحكم', save: 'حفظ', cancel: 'إلغاء', confirm: 'تأكيد', participants: 'المشاركون', results: 'النتائج', schedule: 'الجدول', medals: 'الميداليات', upcoming: 'التالي', live: 'مباشر' },
  ja: { dashboard: 'ダッシュボード', setup: '設定', scoreboard: 'スコアボード', public: '公開', brackets: 'トーナメント表', tables: '順位表', pools: 'グループ', control: 'コントロール', save: '保存', cancel: 'キャンセル', confirm: '確認', participants: '出場者', results: '結果', schedule: '日程', medals: 'メダル', upcoming: '次', live: 'ライブ' },
  ko: { dashboard: '대시보드', setup: '설정', scoreboard: '점수판', public: '공개', brackets: '대진표', tables: '순위표', pools: '조', control: '제어', save: '저장', cancel: '취소', confirm: '확인', participants: '참가자', results: '결과', schedule: '일정', medals: '메달', upcoming: '다음', live: '라이브' },
  zh: { dashboard: '仪表板', setup: '设置', scoreboard: '记分牌', public: '公开', brackets: '对阵表', tables: '积分榜', pools: '小组', control: '控制', save: '保存', cancel: '取消', confirm: '确认', participants: '参赛者', results: '结果', schedule: '赛程', medals: '奖牌', upcoming: '即将开始', live: '直播' },
};

export function translate(locale: string, key: Key): string {
  const safe = (locale in DICTIONARY ? locale : 'en') as Locale;
  return DICTIONARY[safe][key];
}
