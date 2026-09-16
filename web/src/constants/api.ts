export const API_ENDPOINTS = {
  AUTH_TELEGRAM: '/api/v1/auth/telegram',
  AUTH_ME: '/api/v1/auth/me',
  ANALYTICS_OVERVIEW: '/api/v1/analytics/overview',
  ANALYTICS_EXPORT_USERS: '/api/v1/analytics/export/users.csv',
  ANALYTICS_EXPORT_POSTS: '/api/v1/analytics/export/posts.csv',
  POSTS: '/api/v1/posts',
  POST_REACT: '/api/v1/posts/react',
  REPORTS: '/api/v1/reports',
  MARKETING_BROADCAST: '/api/v1/marketing/broadcast',
  CHANNELS: '/api/v1/channels',
  USERS: '/api/v1/users',
  SETTINGS: '/api/v1/settings',
  SETUP_STATUS: '/api/v1/setup/status',
  PAYMENTS_PLANS: '/api/v1/payments/plans',
  PAYMENTS_CHECKOUT: '/api/v1/payments/checkout',
  PAYMENTS_AZPAYS_WEBHOOK: '/api/v1/payments/azpays/webhook',
  DOCS_SCALAR: '/docs',
  GRAPHQL_PLAYGROUND: '/graphql',
} as const;

export const VYNTECH_URLS = {
  WEBSITE: 'https://vyntech.com.au',
  CLOUD_PRODUCT: 'https://www.vyntech.com.au/products/cloud',
  DEVELOPER_CONSOLE: 'https://console.vyntech.com.au',
  GITHUB_REPO: 'https://github.com/vyntechau/TelegramPublisher',
} as const;
