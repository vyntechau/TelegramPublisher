import { PaymentPlan } from '../types/admin';

export const DEFAULT_VIP_PLANS: PaymentPlan[] = [
  {
    id: 'vip_1m',
    name: '1 Month VIP Pass',
    duration_days: 30,
    price_usd: 9.99,
    features: ['Unlimited 4K Streams', 'No Waiting Queue', 'Priority Download Server', 'Instant VIP Badge'],
  },
  {
    id: 'vip_3m',
    name: '3 Months VIP Quarterly',
    duration_days: 90,
    price_usd: 24.99,
    features: ['All 1-Month Benefits', 'Save 17%', 'Exclusive Early Releases', 'Private Support Channel'],
    is_popular: true,
  },
  {
    id: 'vip_1y',
    name: '1 Year Lifetime VIP',
    duration_days: 365,
    price_usd: 79.99,
    features: ['All VIP Privileges', 'Save 33%', 'Direct Developer Chat', 'Custom Reaction Emotes'],
  },
];
