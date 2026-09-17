import React, { useState, useEffect } from 'react';
import { SettingsMap, KeyboardMode, AutoPostFormat, SubscriptionGateway } from '../../types';
import { KEYBOARD_MODES, AUTO_POST_FORMATS, SUBSCRIPTION_GATEWAYS, API_ENDPOINTS, apiUrl } from '../../constants';
import { useLanguage, useTranslation } from '../../context/LanguageContext';
import { ALL_LANGUAGES } from '../../locales';
import { Settings, Save, CheckCircle2, Shield, Radio, Coins, RefreshCw, Send, Layers, ExternalLink, Globe, Check } from 'lucide-react';

interface AdminSettingsTabProps {
  settings: SettingsMap;
  onRefresh: () => void;
}

export const AdminSettingsTab: React.FC<AdminSettingsTabProps> = ({ settings: initialSettings, onRefresh }) => {
  const [formData, setFormData] = useState<SettingsMap>(initialSettings);
  const [saving, setSaving] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const { language, setLanguage, defaultLanguage, setDefaultLanguage, setSupportedLanguagesCSV } = useLanguage();
  const { t } = useTranslation();

  useEffect(() => {
    if (initialSettings && Object.keys(initialSettings).length > 0) {
      setFormData(initialSettings);
    }
  }, [initialSettings]);

  const handleChange = (key: string, value: string) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
    setSavedSuccess(false);
  };

  const handleLanguageToggle = (code: string) => {
    const currentCSV = formData.supported_languages || 'en,fa,ar,ru,es,de,zh';
    const parts = new Set(currentCSV.split(',').map((s) => s.trim().toLowerCase()).filter(Boolean));
    if (parts.has(code)) {
      if (parts.size > 1) { // Prevent disabling all languages
        parts.delete(code);
      }
    } else {
      parts.add(code);
    }
    const newCSV = Array.from(parts).join(',');
    handleChange('supported_languages', newCSV);
    setSupportedLanguagesCSV(newCSV);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setSaving(true);
      await fetch(apiUrl(API_ENDPOINTS.SETTINGS), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify(formData),
      });
      if (formData.default_language) {
        setDefaultLanguage(formData.default_language);
      }
      if (formData.supported_languages) {
        setSupportedLanguagesCSV(formData.supported_languages);
      }
      setSavedSuccess(true);
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  const activeGateway = formData.subscription_gateway || 'azpays';
  const supportedSet = new Set(
    (formData.supported_languages || 'en,fa,ar,ru,es,de,zh')
      .split(',')
      .map((s) => s.trim().toLowerCase())
      .filter(Boolean)
  );

  return (
    <form onSubmit={handleSave} className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t('admin.tab_settings', 'Platform Settings')}</h2>
          <p className="text-xs text-slate-400">{t('settings.loc_subtitle', 'Manage runtime operational keys, auto-cleanup timers, Telegram keyboard modes, and multi-gateway crypto subscriptions')}</p>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="liquid-button px-5 py-2.5 rounded-2xl text-xs sm:text-sm font-bold text-white shadow-lg flex items-center gap-2 self-start sm:self-auto"
        >
          {saving ? <RefreshCw size={14} className="animate-spin" /> : <Save size={14} />}
          <span>{saving ? t('common.saving', 'Saving...') : t('admin.btn_save_all', 'Save Configuration')}</span>
        </button>
      </div>

      {savedSuccess && (
        <div className="p-4 rounded-2xl bg-emerald-500/15 border border-emerald-500/30 text-emerald-300 text-xs font-semibold flex items-center gap-2">
          <CheckCircle2 size={16} />
          <span>{t('admin.config_saved', 'Settings updated successfully!')}</span>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* 0. Localization & Multi-Language Settings */}
        <div className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4 md:col-span-2 border-cyan-500/20">
          <div className="flex items-center justify-between">
            <div className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
              <Globe size={15} className="text-cyan-400" />
              <span>{t('settings.loc_title', 'Localization & Language Settings')}</span>
            </div>
            <span className="text-[10px] text-cyan-300 bg-cyan-500/10 px-2.5 py-0.5 rounded-full border border-cyan-500/20 font-mono">
              7 Locales (RTL Ready)
            </span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
            {/* Default System Language */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">{t('settings.default_lang', 'Default System Language')}</label>
              <select
                value={formData.default_language || 'en'}
                onChange={(e) => handleChange('default_language', e.target.value)}
                className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs text-white"
              >
                {ALL_LANGUAGES.map((l) => (
                  <option key={l.code} value={l.code} className="bg-slate-900 text-white">
                    {l.flag} {l.nativeName} ({l.name}) {l.isRtl ? '— [RTL]' : ''}
                  </option>
                ))}
              </select>
              <p className="text-[10px] text-slate-400">{t('settings.default_lang_help', 'Default language assigned to new visitors and Telegram bot interactions.')}</p>
            </div>

            {/* Personal Session Language Switch */}
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">{t('settings.personal_lang', 'Your Session Language')}</label>
              <select
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs text-white border-cyan-500/30"
              >
                {ALL_LANGUAGES.map((l) => (
                  <option key={l.code} value={l.code} className="bg-slate-900 text-white">
                    {l.flag} {l.nativeName} ({l.name})
                  </option>
                ))}
              </select>
              <p className="text-[10px] text-slate-400">{t('settings.personal_lang_help', 'Change the language for your current browser session.')}</p>
            </div>
          </div>

          {/* Enabled Platform Languages Toggle Grid */}
          <div className="space-y-2 pt-2 border-t border-white/[0.06]">
            <label className="text-xs font-semibold text-slate-300">{t('settings.supported_langs', 'Enabled Platform Languages')}</label>
            <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2">
              {ALL_LANGUAGES.map((l) => {
                const isEnabled = supportedSet.has(l.code);
                return (
                  <button
                    key={l.code}
                    type="button"
                    onClick={() => handleLanguageToggle(l.code)}
                    className={`p-2.5 rounded-xl text-xs font-medium flex items-center justify-between transition-all ${
                      isEnabled
                        ? 'bg-cyan-500/15 border border-cyan-500/40 text-cyan-200 shadow-sm'
                        : 'bg-white/[0.02] border border-white/[0.06] text-slate-500 hover:text-slate-300'
                    }`}
                  >
                    <span className="flex items-center gap-2 truncate">
                      <img
                        src={`https://flagcdn.com/w40/${l.countryCode}.png`}
                        alt={l.name}
                        className="w-4 h-3 object-cover rounded-xs border border-white/20 shadow-xs shrink-0"
                        loading="lazy"
                      />
                      <span className="truncate">{l.nativeName}</span>
                    </span>
                    {isEnabled && <Check size={12} className="text-cyan-400 shrink-0" />}
                  </button>
                );
              })}
            </div>
            <p className="text-[10px] text-slate-400">{t('settings.supported_langs_help', 'Check the languages you want active and visible in the language switcher.')}</p>
          </div>
        </div>

        {/* 1. General & Security Settings */}
        <div className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4">
          <div className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
            <Shield size={14} className="text-cyan-400" />
            <span>{t('settings.media_sec_title', 'General & API Security Settings')}</span>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('onboarding.api_url_label', 'REST API Base URL')}</label>
            <input
              type="url"
              value={formData.api_url ?? ''}
              onChange={(e) => handleChange('api_url', e.target.value)}
              placeholder="http://localhost:8080"
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white"
            />
            <p className="text-[10px] text-slate-400">Public API base URL configured during wizard setup for backend endpoints & webhooks.</p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Telegram Mini App Web URL</label>
            <input
              type="url"
              value={formData.mini_app_url ?? ''}
              onChange={(e) => handleChange('mini_app_url', e.target.value)}
              placeholder="http://localhost:8080"
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white"
            />
            <p className="text-[10px] text-slate-400">URL opened when users click Telegram WebApp buttons inside bot chats.</p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Auto-Delete Cleanup Timer (Seconds)</label>
            <input
              type="number"
              value={formData.auto_delete_seconds || '120'}
              onChange={(e) => handleChange('auto_delete_seconds', e.target.value)}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white"
            />
            <p className="text-[10px] text-slate-400">Media messages sent by bot auto-delete after this duration to prevent leaking.</p>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Copyright Warning Text</label>
            <textarea
              rows={2}
              value={formData.copyright_warning_text || ''}
              onChange={(e) => handleChange('copyright_warning_text', e.target.value)}
              className="liquid-input w-full p-3 rounded-xl text-xs text-white resize-none"
            />
          </div>

          <div className="flex items-center justify-between pt-2">
            <span className="text-xs font-semibold text-slate-300">Force Channel Subscription Gate</span>
            <input
              type="checkbox"
              checked={formData.force_sub_enabled === 'true'}
              onChange={(e) => handleChange('force_sub_enabled', e.target.checked ? 'true' : 'false')}
              className="w-4 h-4 accent-cyan-500 rounded"
            />
          </div>
        </div>

        {/* 2. Auto-Publishing Channels Settings */}
        <div className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4">
          <div className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
            <Send size={14} className="text-indigo-400" />
            <span>Channel Auto-Publishing Targets</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-300">Enable Auto-Publishing to Channels</span>
            <input
              type="checkbox"
              checked={formData.auto_post_enabled === 'true'}
              onChange={(e) => handleChange('auto_post_enabled', e.target.checked ? 'true' : 'false')}
              className="w-4 h-4 accent-indigo-500 rounded"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Broadcast Channels (Comma/Newline separated)</label>
            <textarea
              rows={2}
              value={formData.auto_post_channels || ''}
              onChange={(e) => handleChange('auto_post_channels', e.target.value)}
              placeholder="@cinema_vip, -100123456789, @vyntech_stream"
              className="liquid-input w-full p-3 rounded-xl text-xs font-mono text-white resize-none"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Channel Post Format</label>
            <select
              value={formData.auto_post_format || 'teaser_with_button'}
              onChange={(e) => handleChange('auto_post_format', e.target.value)}
              className="liquid-input w-full px-3 py-2 rounded-xl text-xs text-white"
            >
              <option value="teaser_with_button">Teaser with Deep-Link Button (Recommended)</option>
              <option value="full_media">Full Media File Direct</option>
            </select>
          </div>
        </div>

        {/* 3. Keyboard Modes & Buttons */}
        <div className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4">
          <div className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
            <Radio size={14} className="text-cyan-400" />
            <span>Telegram Keyboard Mode & Buttons</span>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Keyboard Mode</label>
            <div className="grid grid-cols-3 gap-2">
              {KEYBOARD_MODES.map((km) => (
                <button
                  key={km.mode}
                  type="button"
                  onClick={() => handleChange('keyboard_mode', km.mode)}
                  className={`p-2.5 rounded-xl text-xs font-bold transition-all ${
                    formData.keyboard_mode === km.mode ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
                  }`}
                >
                  {km.label}
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-2 pt-1">
            <label className="flex items-center justify-between p-2.5 rounded-xl bg-white/[0.02] text-xs text-slate-300 cursor-pointer">
              <span>↗️ Forward Share</span>
              <input
                type="checkbox"
                checked={formData.show_forward_button === 'true'}
                onChange={(e) => handleChange('show_forward_button', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 accent-cyan-500 rounded"
              />
            </label>

            <label className="flex items-center justify-between p-2.5 rounded-xl bg-white/[0.02] text-xs text-slate-300 cursor-pointer">
              <span>🚨 Report Broken</span>
              <input
                type="checkbox"
                checked={formData.show_report_button === 'true'}
                onChange={(e) => handleChange('show_report_button', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 accent-cyan-500 rounded"
              />
            </label>

            <label className="flex items-center justify-between p-2.5 rounded-xl bg-white/[0.02] text-xs text-slate-300 cursor-pointer">
              <span>👍/👎 Reactions</span>
              <input
                type="checkbox"
                checked={formData.show_reactions === 'true'}
                onChange={(e) => handleChange('show_reactions', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 accent-cyan-500 rounded"
              />
            </label>

            <label className="flex items-center justify-between p-2.5 rounded-xl bg-white/[0.02] text-xs text-slate-300 cursor-pointer">
              <span>🚀 Mini App Web</span>
              <input
                type="checkbox"
                checked={formData.show_mini_app_button === 'true'}
                onChange={(e) => handleChange('show_mini_app_button', e.target.checked ? 'true' : 'false')}
                className="w-4 h-4 accent-cyan-500 rounded"
              />
            </label>
          </div>
        </div>

        {/* 4. Multi-Gateway Crypto Subscription System */}
        <div className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4">
          <div className="text-xs font-bold text-white uppercase tracking-wider flex items-center gap-2">
            <Coins size={14} className="text-amber-400" />
            <span>Crypto Subscription Gateways</span>
          </div>

          <div className="flex items-center justify-between">
            <span className="text-xs font-semibold text-slate-300">Enable Crypto VIP Subscriptions</span>
            <input
              type="checkbox"
              checked={formData.subscription_enabled === 'true'}
              onChange={(e) => handleChange('subscription_enabled', e.target.checked ? 'true' : 'false')}
              className="w-4 h-4 accent-amber-500 rounded"
            />
          </div>

          {/* Gateway Provider Selector */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Active Payment Gateway Provider</label>
            <div className="grid grid-cols-3 gap-2">
              {SUBSCRIPTION_GATEWAYS.map((gw) => (
                <button
                  key={gw.id}
                  type="button"
                  onClick={() => handleChange('subscription_gateway', gw.id)}
                  className={`p-2.5 rounded-xl text-xs font-bold transition-all text-center ${
                    activeGateway === gw.id ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
                  }`}
                >
                  <div className="truncate">{gw.name.split(' ')[0]}</div>
                  <div className="text-[9px] opacity-75 font-normal">{gw.badge}</div>
                </button>
              ))}
            </div>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">API Key ({activeGateway})</label>
            <input
              type="password"
              value={formData.subscription_api_key || formData.azpays_api_key || ''}
              onChange={(e) => {
                handleChange('subscription_api_key', e.target.value);
                handleChange('azpays_api_key', e.target.value);
              }}
              placeholder={`${activeGateway}_live_key_...`}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">Webhook Secret / Signature Key</label>
            <input
              type="password"
              value={formData.subscription_secret_key || formData.azpays_secret_key || ''}
              onChange={(e) => {
                handleChange('subscription_secret_key', e.target.value);
                handleChange('azpays_secret_key', e.target.value);
              }}
              placeholder={`${activeGateway}_secret_...`}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white"
            />
          </div>
        </div>
      </div>
    </form>
  );
};
