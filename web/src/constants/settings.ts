import { SettingsMap, KeyboardMode, AutoPostFormat, SubscriptionGateway } from '../types/settings';

export const KEYBOARD_MODES: Array<{ mode: KeyboardMode; label: string; desc: string }> = [
  { mode: 'both', label: 'Both Keyboards', desc: 'Inline buttons on post + persistent bottom menu' },
  { mode: 'inline', label: 'Inline Only', desc: 'Interactive buttons attached directly to posts' },
  { mode: 'persistent', label: 'Persistent Only', desc: 'Fixed persistent reply keyboard at bottom' },
];

export const AUTO_POST_FORMATS: Array<{ format: AutoPostFormat; label: string; desc: string }> = [
  { format: 'teaser_with_button', label: 'Teaser + Bot Watch Button', desc: 'Protects copyright & drives traffic to the bot player queue' },
  { format: 'full_media', label: 'Direct Media File', desc: 'Sends full video directly to channel' },
];

export const SUBSCRIPTION_GATEWAYS: Array<{ id: SubscriptionGateway; name: string; desc: string; badge: string }> = [
  { id: 'azpays', name: 'AzPays Crypto (Official Go SDK)', desc: 'Direct low-fee crypto processing with cryptographic HMAC webhooks', badge: 'Recommended' },
  { id: 'coinbase', name: 'Coinbase Commerce', desc: 'Accept crypto directly to Coinbase managed merchant accounts', badge: 'Enterprise' },
  { id: 'nowpayments', name: 'NOWPayments.io', desc: 'Accept 150+ cryptocurrencies with instant auto-conversion', badge: 'Global' },
];

export const DEFAULT_SETTINGS: SettingsMap = {
  web_enabled: 'true',
  admin_web_enabled: 'true',
  bot_admin_enabled: 'true',
  mini_app_enabled: 'true',
  mini_app_url: 'http://localhost:8080',
  auto_delete_seconds: '120',
  force_sub_enabled: 'true',
  copyright_warning_text: '⏳ *Copyright Protection*: This content will be automatically deleted in 2 minutes.',
  custom_buttons_json: '[[{"text":"🌐 Official Website","url":"https://vyntech.cloud"},{"text":"⭐ VIP Subscription","url":"https://t.me/yourbot?start=vip"}]]',
  keyboard_mode: 'both',
  show_forward_button: 'true',
  show_report_button: 'true',
  show_reactions: 'true',
  show_mini_app_button: 'true',
  auto_post_enabled: 'false',
  auto_post_channels: '',
  auto_post_format: 'teaser_with_button',
  subscription_enabled: 'false',
  subscription_gateway: 'azpays',
  subscription_api_key: '',
  subscription_secret_key: '',
  subscription_callback_url: 'http://localhost:8080/api/v1/payments/subscription/webhook',
  subscription_success_url: 'http://localhost:8080/payment/success',
  azpays_enabled: 'false',
  azpays_api_key: '',
  azpays_secret_key: '',
  azpays_callback_url: 'http://localhost:8080/api/v1/payments/azpays/webhook',
  azpays_success_url: 'http://localhost:8080/payment/success',
  api_url: 'http://localhost:8080',
  onboarding_step: '1',
  onboarding_completed: 'false',
};
