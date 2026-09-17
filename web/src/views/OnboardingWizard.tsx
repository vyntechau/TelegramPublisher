import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { 
  ONBOARDING_STEPS, 
  DEFAULT_SETTINGS, 
  API_ENDPOINTS, 
  KEYBOARD_MODES, 
  AUTO_POST_FORMATS,
  SUBSCRIPTION_GATEWAYS,
  apiUrl
} from '../constants';
import { SettingsMap } from '../types';
import { useLanguage, useTranslation } from '../context/LanguageContext';
import { ALL_LANGUAGES } from '../locales';
import { 
  Server, 
  BookOpen, 
  Terminal, 
  Send, 
  Radio, 
  Coins, 
  CheckCircle2, 
  ArrowRight, 
  ArrowLeft, 
  Sparkles, 
  Check, 
  RefreshCw, 
  ExternalLink,
  Shield,
  Layers,
  Smartphone,
  Lock,
  Clock,
  Code,
  Globe
} from 'lucide-react';

export const OnboardingWizard: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { setDefaultLanguage } = useLanguage();
  const [currentStep, setCurrentStep] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(true);
  const [saving, setSaving] = useState<boolean>(false);
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'failed' | 'success'>('idle');
  const [testMessage, setTestMessage] = useState<string>('');
  const [settings, setSettings] = useState<SettingsMap>(DEFAULT_SETTINGS);

  useEffect(() => {
    const fetchSettings = async () => {
      try {
        setLoading(true);
        const token = localStorage.getItem('tp_token');
        const headers: Record<string, string> = { 'Content-Type': 'application/json' };
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const resp = await fetch(apiUrl(API_ENDPOINTS.SETTINGS), { headers });
        if (resp.ok) {
          const data = await resp.json();
          setSettings((prev) => ({ ...prev, ...data }));
          
          if (data.onboarding_completed === 'true') {
            localStorage.setItem('tp_onboarding_completed', 'true');
          }

          const savedStep = parseInt(data.onboarding_step || localStorage.getItem('tp_onboarding_step') || '1', 10);
          if (savedStep >= 1 && savedStep <= 5) {
            setCurrentStep(savedStep);
          }
        }
      } catch (e) {
        console.warn('Could not load settings from server, using defaults:', e);
        const cachedStep = parseInt(localStorage.getItem('tp_onboarding_step') || '1', 10);
        if (cachedStep >= 1 && cachedStep <= 5) {
          setCurrentStep(cachedStep);
        }
      } finally {
        setLoading(false);
      }
    };

    fetchSettings();
  }, []);

  const saveStepData = async (nextStep: number, isFinal: boolean = false) => {
    try {
      setSaving(true);
      const token = localStorage.getItem('tp_token');
      const headers: Record<string, string> = { 'Content-Type': 'application/json' };
      if (token) headers['Authorization'] = `Bearer ${token}`;

      const isCompleted = isFinal || settings.onboarding_completed === 'true';

      const updatedPayload: SettingsMap = {
        ...settings,
        onboarding_step: nextStep.toString(),
        onboarding_completed: isCompleted ? 'true' : 'false',
      };

      localStorage.setItem('tp_onboarding_step', nextStep.toString());
      if (isCompleted) {
        localStorage.setItem('tp_onboarding_completed', 'true');
      }

      await fetch(apiUrl(API_ENDPOINTS.SETTINGS), {
        method: 'POST',
        headers,
        body: JSON.stringify(updatedPayload),
      });
    } catch (e) {
      console.warn('Failed to sync settings with server:', e);
    } finally {
      setSaving(false);
    }
  };

  const checkApiAvailability = async (urlToCheck?: string): Promise<boolean> => {
    setTestStatus('testing');
    setTestMessage(t('onboarding.test_pinging', 'Pinging REST API endpoint...'));
    try {
      const rawUrl = urlToCheck !== undefined ? urlToCheck : (settings.api_url || '');
      const targetUrl = rawUrl ? rawUrl.replace(/\/+$/, '') : window.location.origin;

      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 6000);

      const resp = await fetch(`${targetUrl}${API_ENDPOINTS.PAYMENTS_PLANS}`, {
        signal: controller.signal,
      });
      clearTimeout(timeoutId);

      if (resp.ok) {
        setTestStatus('success');
        setTestMessage(t('onboarding.test_success', 'Connected successfully to TelegramPublisher API!'));
        return true;
      } else {
        setTestStatus('failed');
        setTestMessage(`${t('onboarding.test_http_error', 'Received HTTP error')}: ${resp.status}`);
        return false;
      }
    } catch (err: any) {
      setTestStatus('failed');
      const msg = err.name === 'AbortError'
        ? t('onboarding.test_timeout', 'Connection timed out reaching API endpoint')
        : (err.message || t('onboarding.test_failed', 'Failed to reach API URL'));
      setTestMessage(msg);
      return false;
    }
  };

  const handleNext = async () => {
    if (currentStep === 1) {
      let isAvailable = testStatus === 'success';
      if (!isAvailable) {
        isAvailable = await checkApiAvailability();
      }
      if (!isAvailable) {
        return;
      }
    }

    if (currentStep < 5) {
      const next = currentStep + 1;
      await saveStepData(next);
      setCurrentStep(next);
    } else {
      await saveStepData(5, true);
      navigate('/admin/overview');
    }
  };

  const handleSkip = async () => {
    if (currentStep === 1) {
      // Step 1 cannot be skipped without working API
      let isAvailable = testStatus === 'success';
      if (!isAvailable) {
        isAvailable = await checkApiAvailability();
      }
      if (!isAvailable) {
        return;
      }
    }

    if (currentStep < 5) {
      const next = currentStep + 1;
      await saveStepData(next);
      setCurrentStep(next);
    } else {
      await saveStepData(5, true);
      navigate('/admin/overview');
    }
  };

  const handlePrev = () => {
    if (currentStep > 1) {
      const prev = currentStep - 1;
      localStorage.setItem('tp_onboarding_step', prev.toString());
      setCurrentStep(prev);
    }
  };

  const updateSetting = (key: string, value: string) => {
    setSettings((prev) => ({ ...prev, [key]: value }));
  };

  if (loading) {
    return (
      <div className="min-h-[70vh] flex items-center justify-center">
        <div className="liquid-glass p-8 rounded-3xl border border-white/10 flex flex-col items-center gap-3">
          <RefreshCw className="animate-spin text-cyan-400" size={24} />
          <div className="text-sm font-semibold text-slate-300">
            {t('onboarding.loading', 'Loading setup configuration...')}
          </div>
        </div>
      </div>
    );
  }

  const activeApiUrl = settings.api_url ? settings.api_url.replace(/\/+$/, '') : window.location.origin;
  const activeGateway = settings.subscription_gateway || 'azpays';

  return (
    <div className="max-w-4xl mx-auto space-y-6 sm:space-y-8 py-2 sm:py-6 px-3 sm:px-6">
      {/* Onboarding Header */}
      <div className="text-center space-y-2">
        <h1 className="text-2xl sm:text-3xl md:text-4xl font-black text-white tracking-tight">
          {t('onboarding.wizard_title', 'System Setup & Onboarding')}
        </h1>
        <p className="text-xs sm:text-sm text-slate-400 max-w-lg mx-auto leading-relaxed">
          {t('onboarding.wizard_desc', 'Complete end-to-end configuration for operational endpoints, multi-language localization, Telegram channels, and crypto subscriptions.')}
        </p>
      </div>

      {/* Non-Clickable Sequential Stepper Progress Bar (Users cannot jump steps directly) */}
      <div className="liquid-glass p-3 sm:p-4 rounded-2xl sm:rounded-3xl border border-white/10 shadow-lg">
        {/* Mobile Step Fraction Header */}
        <div className="flex sm:hidden items-center justify-between pb-1">
          <div className="text-xs font-bold text-cyan-400 uppercase tracking-wider">
            {t('onboarding.step_indicator', `Step ${currentStep} of ${ONBOARDING_STEPS.length}`)
              .replace('{current}', String(currentStep))
              .replace('{total}', String(ONBOARDING_STEPS.length))}
          </div>
          <div className="text-xs font-semibold text-white">
            {t(`onboarding.step${currentStep}_title`, ONBOARDING_STEPS[currentStep - 1].title)}
          </div>
        </div>

        {/* Stepper Track (Read-Only Display) */}
        <div className="grid grid-cols-5 gap-1.5 sm:gap-3">
          {ONBOARDING_STEPS.map((step) => {
            const Icon = step.icon;
            const isCompleted = step.id < currentStep;
            const isActive = step.id === currentStep;

            return (
              <div
                key={step.id}
                className={`flex flex-col sm:flex-row items-center sm:items-start gap-1.5 sm:gap-2.5 p-2 sm:p-2.5 rounded-xl sm:rounded-2xl transition-all text-start pointer-events-none cursor-default select-none ${
                  isActive
                    ? 'liquid-glass-card border-cyan-500/40 shadow-md shadow-cyan-500/10'
                    : isCompleted
                    ? 'bg-white/[0.02]'
                    : 'opacity-40'
                }`}
              >
                <div
                  className={`w-6 h-6 sm:w-8 sm:h-8 rounded-xl flex items-center justify-center shrink-0 transition-transform ${
                    isActive
                      ? 'bg-cyan-500 text-slate-950 font-bold shadow-md shadow-cyan-500/40 scale-105'
                      : isCompleted
                      ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30'
                      : 'bg-white/5 text-slate-400 border border-white/10'
                  }`}
                >
                  {isCompleted ? <Check size={14} /> : <Icon size={14} />}
                </div>

                <div className="hidden sm:block overflow-hidden">
                  <div className={`text-xs font-bold truncate ${isActive ? 'text-white' : 'text-slate-300'}`}>
                    {t(`onboarding.step${step.id}_title`, step.title)}
                  </div>
                  <div className="text-[10px] text-slate-400 truncate">
                    {t(`onboarding.step${step.id}_desc`, step.desc)}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Main Step Card Form */}
      <div className="liquid-glass-card p-5 sm:p-8 md:p-10 rounded-3xl border border-white/10 shadow-2xl relative overflow-hidden">
        <div className="ambient-orb w-64 h-64 bg-cyan-500/10 -top-20 -right-20 pointer-events-none" />

        {/* STEP 1: API & CORE SECURITY */}
        {currentStep === 1 && (
          <div className="space-y-6">
            <div className="space-y-1">
              <div className="inline-flex items-center gap-1.5 text-xs font-bold text-cyan-400 uppercase tracking-wider">
                <Server size={14} />
                <span>{t('onboarding.step1_badge', 'Step 1: API & Security Settings')}</span>
              </div>
              <h2 className="text-xl sm:text-2xl font-bold text-white">
                {t('onboarding.step1_heading', 'Configure API Endpoint & Cleanup Rules')}
              </h2>
              <p className="text-xs sm:text-sm text-slate-300">
                {t('onboarding.step1_sub', 'Set up your client REST API base URL and automatic copyright protection rules.')}
              </p>
            </div>

            {/* API URL Input with Live Test */}
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200 flex items-center justify-between">
                <span>{t('onboarding.api_url_label', 'REST API Base URL')}</span>
                <span className="text-[11px] text-slate-400">{t('onboarding.twelve_factor', 'Twelve-Factor compliant')}</span>
              </label>
              <div className="flex flex-col sm:flex-row gap-2">
                <input
                  type="text"
                  value={settings.api_url || ''}
                  onChange={(e) => {
                    updateSetting('api_url', e.target.value);
                    if (testStatus !== 'idle') {
                      setTestStatus('idle');
                      setTestMessage('');
                    }
                  }}
                  placeholder="https://api.yourdomain.com or http://localhost:8080"
                  className="liquid-input flex-1 px-4 py-3 rounded-xl text-sm font-mono text-white placeholder-slate-500 min-h-[46px]"
                />
                <button
                  type="button"
                  onClick={() => checkApiAvailability()}
                  disabled={testStatus === 'testing'}
                  className={`px-5 py-3 rounded-xl text-xs sm:text-sm font-bold transition-all flex items-center justify-center gap-2 shrink-0 min-h-[46px] border ${
                    testStatus === 'success'
                      ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
                      : 'liquid-glass text-white hover:bg-white/10 border-white/15'
                  }`}
                >
                  {testStatus === 'testing' ? (
                    <RefreshCw size={14} className="animate-spin text-cyan-400" />
                  ) : (
                    <CheckCircle2 size={14} className={testStatus === 'success' ? 'text-emerald-400' : 'text-cyan-400'} />
                  )}
                  <span>{t('onboarding.btn_test_conn', 'Test Connection')}</span>
                </button>
              </div>

              {testStatus === 'idle' ? (
                <div className="p-3 rounded-xl text-xs font-medium flex items-center gap-2 bg-amber-500/10 border border-amber-500/25 text-amber-300">
                  <Shield size={14} className="shrink-0 text-amber-400" />
                  <span>{t('onboarding.api_required_notice', 'API connection verification is required before proceeding to the next step.')}</span>
                </div>
              ) : (
                <div
                  className={`p-3 rounded-xl text-xs font-medium flex items-center gap-2 ${
                    testStatus === 'success'
                      ? 'bg-emerald-500/15 border border-emerald-500/30 text-emerald-300'
                      : testStatus === 'failed'
                      ? 'bg-red-500/15 border border-red-500/30 text-red-300'
                      : 'bg-cyan-500/15 border border-cyan-500/30 text-cyan-300'
                  }`}
                >
                  {testStatus === 'testing' && <RefreshCw size={14} className="animate-spin text-cyan-400 shrink-0" />}
                  {testStatus === 'success' && <CheckCircle2 size={14} className="text-emerald-400 shrink-0" />}
                  {testStatus === 'failed' && <Shield size={14} className="text-red-400 shrink-0" />}
                  <span>{testMessage}</span>
                </div>
              )}
            </div>

            {/* Interactive API Docs Box */}
            <div className="liquid-glass p-4 sm:p-5 rounded-2xl border border-white/10 space-y-2.5 bg-gradient-to-r from-blue-950/20 via-transparent to-purple-950/20">
              <div className="text-xs font-bold text-slate-200 flex items-center gap-2">
                <BookOpen size={15} className="text-cyan-400" />
                <span>{t('onboarding.docs_box_title', 'Interactive Documentation & Schema URLs')}</span>
              </div>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5 pt-1">
                <a
                  href={`${activeApiUrl}${API_ENDPOINTS.DOCS_SCALAR}`}
                  target="_blank"
                  rel="noreferrer"
                  className="p-3 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/5 transition-all flex items-center justify-between group"
                >
                  <div className="flex items-center gap-2">
                    <BookOpen size={15} className="text-cyan-400" />
                    <div>
                      <div className="text-xs font-bold text-white group-hover:text-cyan-300">
                        {t('onboarding.scalar_docs', 'Scalar REST Docs')}
                      </div>
                      <div className="text-[10px] text-slate-400 font-mono">{activeApiUrl}/docs</div>
                    </div>
                  </div>
                  <ExternalLink size={12} className="text-slate-400 group-hover:text-white" />
                </a>

                <a
                  href={`${activeApiUrl}${API_ENDPOINTS.GRAPHQL_PLAYGROUND}`}
                  target="_blank"
                  rel="noreferrer"
                  className="p-3 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/5 transition-all flex items-center justify-between group"
                >
                  <div className="flex items-center gap-2">
                    <Terminal size={15} className="text-purple-400" />
                    <div>
                      <div className="text-xs font-bold text-white group-hover:text-purple-300">
                        {t('onboarding.graphql_playground', 'GraphQL Playground')}
                      </div>
                      <div className="text-[10px] text-slate-400 font-mono">{activeApiUrl}/graphql</div>
                    </div>
                  </div>
                  <ExternalLink size={12} className="text-slate-400 group-hover:text-white" />
                </a>
              </div>
            </div>

            {/* System Default Language & Localization */}
            <div className="space-y-3 p-4 sm:p-5 rounded-2xl liquid-glass border border-cyan-500/20 bg-cyan-950/10">
              <div className="flex items-center justify-between">
                <label className="text-xs font-semibold text-slate-200 flex items-center gap-1.5">
                  <Globe size={14} className="text-cyan-400" />
                  <span>{t('settings.default_lang', 'Default System Language')}</span>
                </label>
                <span className="text-[10px] text-cyan-300 bg-cyan-500/10 px-2 py-0.5 rounded-full border border-cyan-500/20 font-mono">
                  FlagCDN
                </span>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
                {ALL_LANGUAGES.map((l) => {
                  const isSelected = (settings.default_language || 'en') === l.code;
                  return (
                    <button
                      key={l.code}
                      type="button"
                      onClick={() => {
                        updateSetting('default_language', l.code);
                        setDefaultLanguage(l.code);
                      }}
                      className={`p-2.5 rounded-xl text-xs font-medium flex items-center justify-between gap-2 border transition-all text-start ${
                        isSelected
                          ? 'liquid-pill-active text-white border-cyan-400 shadow-md shadow-cyan-500/20'
                          : 'bg-white/[0.03] border-white/10 text-slate-300 hover:bg-white/[0.07]'
                      }`}
                    >
                      <div className="flex items-center gap-2 truncate">
                        <img
                          src={l.flagUrl || `https://flagcdn.com/w40/${l.countryCode}.png`}
                          alt={l.name}
                          className="w-5 h-3.5 object-cover rounded-xs border border-white/20 shadow-xs shrink-0"
                          loading="lazy"
                        />
                        <div className="truncate">
                          <div className="font-bold leading-tight truncate">{l.nativeName}</div>
                          <div className="text-[10px] text-slate-400 leading-tight truncate">{l.name} {l.isRtl ? '• RTL' : ''}</div>
                        </div>
                      </div>
                      {isSelected && <Check size={14} className="text-cyan-400 shrink-0 ms-1" />}
                    </button>
                  );
                })}
              </div>

              <p className="text-[10px] text-slate-400">
                {t('settings.default_lang_help', 'Default language assigned to new visitors and Telegram bot interactions.')}
              </p>
            </div>

            {/* Auto-Delete & Copyright Protection Settings */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-200 flex items-center gap-1.5">
                  <Clock size={13} className="text-amber-400" />
                  <span>{t('onboarding.ttl_timer', 'Auto-Delete Cleanup Timer (Seconds)')}</span>
                </label>
                <input
                  type="number"
                  value={settings.auto_delete_seconds || '120'}
                  onChange={(e) => updateSetting('auto_delete_seconds', e.target.value)}
                  className="liquid-input w-full p-2.5 rounded-xl text-xs font-mono text-white"
                />
              </div>

              <div className="flex items-center justify-between p-3 rounded-xl liquid-glass border border-white/10 self-end">
                <span className="text-xs font-semibold text-slate-200">
                  {t('onboarding.force_sub', 'Force Subscription Gate')}
                </span>
                <input
                  type="checkbox"
                  checked={settings.force_sub_enabled === 'true'}
                  onChange={(e) => updateSetting('force_sub_enabled', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 accent-cyan-500 rounded"
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-200">
                {t('onboarding.copyright_warning', 'Copyright Warning Notice')}
              </label>
              <textarea
                rows={2}
                value={settings.copyright_warning_text || ''}
                onChange={(e) => updateSetting('copyright_warning_text', e.target.value)}
                className="liquid-input w-full p-3 rounded-xl text-xs text-white resize-none"
              />
            </div>
          </div>
        )}

        {/* STEP 2: BOT & CHANNEL AUTO-PUBLISHING */}
        {currentStep === 2 && (
          <div className="space-y-6">
            <div className="space-y-1">
              <div className="inline-flex items-center gap-1.5 text-xs font-bold text-cyan-400 uppercase tracking-wider">
                <Send size={14} />
                <span>{t('onboarding.step2_badge', 'Step 2: Bot & Channel Auto-Publishing')}</span>
              </div>
              <h2 className="text-xl sm:text-2xl font-bold text-white">
                {t('onboarding.step2_heading', 'Telegram Bot & Broadcast Channels')}
              </h2>
              <p className="text-xs sm:text-sm text-slate-300">
                {t('onboarding.step2_sub', 'Configure your Telegram bot token and the channels where new video releases are published automatically.')}
              </p>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-200">
                {t('onboarding.bot_token_label', 'Telegram Bot Token (Optional if set in config)')}
              </label>
              <input
                type="password"
                value={settings.bot_token || ''}
                onChange={(e) => updateSetting('bot_token', e.target.value)}
                placeholder="123456789:ABCdefGhIJKlmNoPQRsTUVwxyZ..."
                className="liquid-input w-full p-3 rounded-xl text-xs font-mono text-white placeholder-slate-500"
              />
            </div>

            <div className="liquid-glass p-4 rounded-2xl border border-white/10 flex items-center justify-between">
              <div>
                <div className="text-sm font-bold text-white">
                  {t('onboarding.auto_post_enable', 'Enable Multi-Channel Auto-Publishing')}
                </div>
                <div className="text-xs text-slate-400">
                  {t('onboarding.auto_post_desc', 'Broadcast new videos to channels when released in Author Studio')}
                </div>
              </div>
              <input
                type="checkbox"
                checked={settings.auto_post_enabled === 'true'}
                onChange={(e) => updateSetting('auto_post_enabled', e.target.checked ? 'true' : 'false')}
                className="w-5 h-5 accent-cyan-500 rounded cursor-pointer"
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-200 flex items-center justify-between">
                <span>{t('onboarding.target_channels', 'Target Channels (Comma-separated)')}</span>
                <span className="text-[11px] text-slate-400">e.g. @MyChannel, -100123456789</span>
              </label>
              <textarea
                value={settings.auto_post_channels || ''}
                onChange={(e) => updateSetting('auto_post_channels', e.target.value)}
                rows={2}
                placeholder="@cinema_hub, @vyntech_media, -100192837465"
                className="liquid-input w-full p-3.5 rounded-xl text-xs sm:text-sm font-mono text-white placeholder-slate-500 resize-none"
              />
            </div>

            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200">
                {t('onboarding.post_format', 'Publication Format in Channels')}
              </label>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {AUTO_POST_FORMATS.map((f) => (
                  <button
                    key={f.format}
                    type="button"
                    onClick={() => updateSetting('auto_post_format', f.format)}
                    className={`p-3.5 rounded-xl text-start border transition-all ${
                      settings.auto_post_format === f.format
                        ? 'liquid-pill-active text-white'
                        : 'bg-white/[0.02] border-white/10 text-slate-300 hover:bg-white/[0.05]'
                    }`}
                  >
                    <div className="text-xs font-bold mb-1 flex items-center gap-1.5">
                      <Sparkles size={13} />
                      <span>{f.format === 'teaser_with_button' ? t('onboarding.format_teaser', f.label) : t('onboarding.format_full_media', f.label)}</span>
                    </div>
                    <div className="text-[11px] opacity-80">
                      {f.format === 'teaser_with_button' ? t('onboarding.format_teaser_desc', f.desc) : t('onboarding.format_full_media_desc', f.desc)}
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* STEP 3: POST KEYBOARDS & BUTTONS */}
        {currentStep === 3 && (
          <div className="space-y-6">
            <div className="space-y-1">
              <div className="inline-flex items-center gap-1.5 text-xs font-bold text-cyan-400 uppercase tracking-wider">
                <Radio size={14} />
                <span>{t('onboarding.step3_badge', 'Step 3: Interactive Post Buttons')}</span>
              </div>
              <h2 className="text-xl sm:text-2xl font-bold text-white">
                {t('onboarding.step3_heading', 'Keyboard & Action Controls')}
              </h2>
              <p className="text-xs sm:text-sm text-slate-300">
                {t('onboarding.step3_sub', 'Customize the buttons and reply keyboards that appear beneath media sent by the bot.')}
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200">
                {t('onboarding.keyboard_mode_label', 'Telegram Keyboard Display Mode')}
              </label>
              <div className="grid grid-cols-3 gap-2">
                {KEYBOARD_MODES.map((m) => {
                  const modeLabel = m.mode === 'inline' ? t('onboarding.kb_inline', m.label) : m.mode === 'persistent' ? t('onboarding.kb_persistent', m.label) : t('onboarding.kb_both', m.label);
                  const modeDesc = m.mode === 'inline' ? t('onboarding.kb_inline_desc', m.desc) : m.mode === 'persistent' ? t('onboarding.kb_persistent_desc', m.desc) : t('onboarding.kb_both_desc', m.desc);
                  return (
                    <button
                      key={m.mode}
                      type="button"
                      onClick={() => updateSetting('keyboard_mode', m.mode)}
                      className={`p-3 rounded-xl text-center border transition-all ${
                        settings.keyboard_mode === m.mode
                          ? 'liquid-pill-active text-white'
                          : 'bg-white/[0.02] border-white/10 text-slate-300 hover:bg-white/[0.05]'
                      }`}
                    >
                      <div className="text-xs font-bold">{modeLabel}</div>
                      <div className="text-[10px] opacity-75 hidden sm:block">{modeDesc}</div>
                    </button>
                  );
                })}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <label className="liquid-glass p-3.5 rounded-xl border border-white/10 flex items-center justify-between cursor-pointer">
                <span className="text-xs font-semibold text-white">↗️ {t('onboarding.btn_forward', 'Forward / Share Button')}</span>
                <input
                  type="checkbox"
                  checked={settings.show_forward_button === 'true'}
                  onChange={(e) => updateSetting('show_forward_button', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 accent-cyan-500 rounded"
                />
              </label>

              <label className="liquid-glass p-3.5 rounded-xl border border-white/10 flex items-center justify-between cursor-pointer">
                <span className="text-xs font-semibold text-white">🚨 {t('onboarding.btn_report', 'Report Broken File Button')}</span>
                <input
                  type="checkbox"
                  checked={settings.show_report_button === 'true'}
                  onChange={(e) => updateSetting('show_report_button', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 accent-cyan-500 rounded"
                />
              </label>

              <label className="liquid-glass p-3.5 rounded-xl border border-white/10 flex items-center justify-between cursor-pointer">
                <span className="text-xs font-semibold text-white">👍/👎 {t('onboarding.btn_reactions', 'Reaction Buttons')}</span>
                <input
                  type="checkbox"
                  checked={settings.show_reactions === 'true'}
                  onChange={(e) => updateSetting('show_reactions', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 accent-cyan-500 rounded"
                />
              </label>

              <label className="liquid-glass p-3.5 rounded-xl border border-white/10 flex items-center justify-between cursor-pointer">
                <span className="text-xs font-semibold text-white">🚀 {t('onboarding.btn_miniapp', 'Mini App Web Player Button')}</span>
                <input
                  type="checkbox"
                  checked={settings.show_mini_app_button === 'true'}
                  onChange={(e) => updateSetting('show_mini_app_button', e.target.checked ? 'true' : 'false')}
                  className="w-4 h-4 accent-cyan-500 rounded"
                />
              </label>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-200 flex items-center gap-1.5">
                <Code size={13} className="text-indigo-400" />
                <span>{t('onboarding.custom_buttons_json', 'Custom Keyboards / Message Buttons (JSON format)')}</span>
              </label>
              <textarea
                rows={2}
                value={settings.custom_buttons_json || ''}
                onChange={(e) => updateSetting('custom_buttons_json', e.target.value)}
                placeholder='[[{"text":"🌐 Official Website","url":"https://vyntech.cloud"}]]'
                className="liquid-input w-full p-3 rounded-xl text-xs font-mono text-white resize-none"
              />
            </div>
          </div>
        )}

        {/* STEP 4: MULTI-GATEWAY CRYPTO SUBSCRIPTIONS */}
        {currentStep === 4 && (
          <div className="space-y-6">
            <div className="space-y-1">
              <div className="inline-flex items-center gap-1.5 text-xs font-bold text-cyan-400 uppercase tracking-wider">
                <Coins size={14} />
                <span>{t('onboarding.step4_badge', 'Step 4: Crypto Subscription Gateways')}</span>
              </div>
              <h2 className="text-xl sm:text-2xl font-bold text-white">
                {t('onboarding.step4_heading', 'VIP Monetization & Gateway Setup')}
              </h2>
              <p className="text-xs sm:text-sm text-slate-300">
                {t('onboarding.step4_sub', 'Choose your cryptocurrency gateway provider (AzPays, Coinbase Commerce, or NOWPayments.io) to accept VIP subscription payments.')}
              </p>
            </div>

            <div className="liquid-glass p-4 rounded-2xl border border-white/10 flex items-center justify-between">
              <div>
                <div className="text-sm font-bold text-white">
                  {t('onboarding.crypto_sub_enable', 'Enable Crypto VIP Subscriptions')}
                </div>
                <div className="text-xs text-slate-400">
                  {t('onboarding.crypto_sub_desc', 'Process cryptocurrency checkouts for VIP memberships')}
                </div>
              </div>
              <input
                type="checkbox"
                checked={settings.subscription_enabled === 'true'}
                onChange={(e) => updateSetting('subscription_enabled', e.target.checked ? 'true' : 'false')}
                className="w-5 h-5 accent-cyan-500 rounded cursor-pointer"
              />
            </div>

            {/* Gateway Provider Selector */}
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200">
                {t('onboarding.active_gateway', 'Active Payment Gateway Provider')}
              </label>
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                {SUBSCRIPTION_GATEWAYS.map((gw) => (
                  <button
                    key={gw.id}
                    type="button"
                    onClick={() => updateSetting('subscription_gateway', gw.id)}
                    className={`p-3.5 rounded-xl text-start border transition-all ${
                      activeGateway === gw.id
                        ? 'liquid-pill-active text-white'
                        : 'bg-white/[0.02] border-white/10 text-slate-300 hover:bg-white/[0.05]'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-xs font-bold">{gw.name.split(' ')[0]}</span>
                      <span className="text-[9px] px-1.5 py-0.5 rounded-full bg-white/10 text-cyan-300 font-extrabold">{gw.badge}</span>
                    </div>
                    <div className="text-[11px] opacity-75">{gw.desc}</div>
                  </button>
                ))}
              </div>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-200">
                  {t('onboarding.gateway_api_key', 'API Key')} ({activeGateway})
                </label>
                <input
                  type="password"
                  value={settings.subscription_api_key || settings.azpays_api_key || ''}
                  onChange={(e) => {
                    updateSetting('subscription_api_key', e.target.value);
                    updateSetting('azpays_api_key', e.target.value);
                  }}
                  placeholder={`${activeGateway}_live_key_...`}
                  className="liquid-input w-full p-3 rounded-xl text-xs sm:text-sm font-mono text-white placeholder-slate-500 min-h-[44px]"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-slate-200">
                  {t('onboarding.gateway_secret_key', 'Webhook Secret / Signature Key')}
                </label>
                <input
                  type="password"
                  value={settings.subscription_secret_key || settings.azpays_secret_key || ''}
                  onChange={(e) => {
                    updateSetting('subscription_secret_key', e.target.value);
                    updateSetting('azpays_secret_key', e.target.value);
                  }}
                  placeholder={`${activeGateway}_secret_...`}
                  className="liquid-input w-full p-3 rounded-xl text-xs sm:text-sm font-mono text-white placeholder-slate-500 min-h-[44px]"
                />
              </div>
            </div>
          </div>
        )}

        {/* STEP 5: REVIEW ALL SETTINGS & LAUNCH */}
        {currentStep === 5 && (
          <div className="space-y-6 text-center py-2">
            <div className="w-14 h-14 rounded-2xl bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center mx-auto shadow-lg shadow-cyan-500/25">
              <Sparkles size={28} className="text-white" />
            </div>

            <div className="space-y-1">
              <h2 className="text-2xl sm:text-3xl font-black text-white">
                {t('onboarding.step5_heading', 'System Ready for Production!')}
              </h2>
              <p className="text-xs sm:text-sm text-slate-300 max-w-md mx-auto">
                {t('onboarding.step5_sub', 'All settings are configured and synchronized with the database.')}
              </p>
            </div>

            {/* Configuration Checklist Summary */}
            <div className="liquid-glass p-4 sm:p-5 rounded-2xl border border-white/10 text-start space-y-2.5 text-xs max-w-xl mx-auto">
              <div className="text-[11px] font-bold text-slate-400 uppercase tracking-wider mb-1">
                {t('onboarding.checklist_title', 'Configuration Checklist:')}
              </div>
              <div className="flex items-center justify-between py-1 border-b border-white/5">
                <span className="text-slate-300">{t('onboarding.check_api_endpoint', 'API Endpoint')}</span>
                <span className="font-mono text-cyan-400">{activeApiUrl}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-white/5">
                <span className="text-slate-300">{t('onboarding.check_autodelete', 'Auto-Delete Duration')}</span>
                <span className="font-mono text-white">{settings.auto_delete_seconds || '120'} {t('common.seconds', 'seconds')}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-white/5">
                <span className="text-slate-300">{t('onboarding.check_keyboard_mode', 'Keyboard Layout Mode')}</span>
                <span className="capitalize font-bold text-white">{settings.keyboard_mode || 'both'}</span>
              </div>
              <div className="flex items-center justify-between py-1 border-b border-white/5">
                <span className="text-slate-300">{t('onboarding.check_channel_post', 'Channel Auto-Publishing')}</span>
                <span className="font-bold text-white">
                  {settings.auto_post_enabled === 'true' ? t('common.active', 'Enabled') : t('common.dismissed', 'Disabled')}
                </span>
              </div>
              <div className="flex items-center justify-between py-1">
                <span className="text-slate-300">{t('onboarding.check_crypto_gw', 'Crypto Subscription Gateway')}</span>
                <span className="capitalize font-bold text-amber-400">{activeGateway}</span>
              </div>
            </div>

            {/* Quick Destination Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-start pt-2">
              <button
                type="button"
                onClick={() => {
                  saveStepData(5, true);
                  navigate('/admin/overview');
                }}
                className="liquid-glass p-4 rounded-2xl border border-white/10 hover:border-cyan-500/30 transition-all group text-start"
              >
                <div className="w-8 h-8 rounded-lg bg-cyan-500/10 flex items-center justify-center text-cyan-400 mb-2 group-hover:scale-110 transition-transform">
                  <Shield size={16} />
                </div>
                <div className="text-xs font-bold text-white mb-1">
                  {t('admin.title', 'Admin Dashboard')}
                </div>
                <div className="text-[11px] text-slate-400">
                  {t('onboarding.card_admin_desc', 'Platform KPIs & live user moderation')}
                </div>
              </button>

              <button
                type="button"
                onClick={() => {
                  saveStepData(5, true);
                  navigate('/studio/upload');
                }}
                className="liquid-glass p-4 rounded-2xl border border-white/10 hover:border-indigo-500/30 transition-all group text-start"
              >
                <div className="w-8 h-8 rounded-lg bg-indigo-500/10 flex items-center justify-center text-indigo-400 mb-2 group-hover:scale-110 transition-transform">
                  <Layers size={16} />
                </div>
                <div className="text-xs font-bold text-white mb-1">
                  {t('studio.title', 'Author Studio')}
                </div>
                <div className="text-[11px] text-slate-400">
                  {t('onboarding.card_studio_desc', 'Upload and register video releases')}
                </div>
              </button>

              <button
                type="button"
                onClick={() => {
                  saveStepData(5, true);
                  navigate('/catalog');
                }}
                className="liquid-glass p-4 rounded-2xl border border-white/10 hover:border-emerald-500/30 transition-all group text-start"
              >
                <div className="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-400 mb-2 group-hover:scale-110 transition-transform">
                  <Smartphone size={16} />
                </div>
                <div className="text-xs font-bold text-white mb-1">
                  {t('user.player_title', 'User Mini App')}
                </div>
                <div className="text-[11px] text-slate-400">
                  {t('onboarding.card_user_desc', 'Browse media catalog & test web player')}
                </div>
              </button>
            </div>
          </div>
        )}

        {/* Action Buttons: Prev, Skip, Next */}
        <div className="flex flex-col-reverse sm:flex-row items-center justify-between gap-3 pt-6 border-t border-white/[0.08] mt-8">
          <div>
            {currentStep > 1 && (
              <button
                type="button"
                onClick={handlePrev}
                className="liquid-pill px-5 py-2.5 rounded-xl text-xs sm:text-sm font-semibold text-slate-300 hover:text-white transition-all flex items-center gap-1.5 w-full sm:w-auto justify-center"
              >
                <ArrowLeft size={14} className="rtl:rotate-180" />
                <span>{t('onboarding.btn_prev', 'Previous')}</span>
              </button>
            )}
          </div>

          <div className="flex items-center gap-2.5 w-full sm:w-auto">
            {/* Skip Option (Available on steps 2, 3, 4) */}
            {currentStep > 1 && currentStep < 5 && (
              <button
                type="button"
                onClick={handleSkip}
                className="liquid-pill px-4 py-2.5 rounded-xl text-xs sm:text-sm font-medium text-slate-400 hover:text-slate-200 transition-all flex-1 sm:flex-initial text-center"
              >
                {t('onboarding.btn_skip', 'Skip this step')}
              </button>
            )}

            <button
              type="button"
              onClick={handleNext}
              disabled={saving}
              className="liquid-button px-6 py-2.5 rounded-xl text-xs sm:text-sm font-bold text-white shadow-lg flex items-center justify-center gap-2 flex-1 sm:flex-initial min-w-[130px]"
            >
              {saving ? (
                <RefreshCw size={14} className="animate-spin" />
              ) : currentStep === 5 ? (
                <span>{t('onboarding.btn_finish', 'Finish & Launch')}</span>
              ) : (
                <>
                  <span>{t('onboarding.btn_next', 'Continue')}</span>
                  <ArrowRight size={14} className="rtl:rotate-180" />
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
