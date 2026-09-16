import React, { useState } from 'react';
import { DEFAULT_VIP_PLANS, API_ENDPOINTS } from '../../constants';
import { useTranslation } from '../../context/LanguageContext';
import { Crown, X, Coins, ExternalLink, RefreshCw } from 'lucide-react';

interface VipModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const VipModal: React.FC<VipModalProps> = ({ isOpen, onClose }) => {
  const { t } = useTranslation();
  const [selectedPlan, setSelectedPlan] = useState(DEFAULT_VIP_PLANS[1].id);
  const [currency, setCurrency] = useState('USDT');
  const [loading, setLoading] = useState(false);
  const [invoice, setInvoice] = useState<any>(null);

  if (!isOpen) return null;

  const handleCheckout = async () => {
    try {
      setLoading(true);
      const resp = await fetch(API_ENDPOINTS.PAYMENTS_CHECKOUT, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({
          plan_id: selectedPlan,
          currency: currency,
        }),
      });

      if (resp.ok) {
        const inv = await resp.json();
        setInvoice(inv);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const gatewayName = invoice?.gateway ? (invoice.gateway === 'coinbase' ? 'Coinbase Commerce' : invoice.gateway === 'nowpayments' ? 'NOWPayments.io' : 'AzPays') : 'Crypto Gateway';
  const planPrice = DEFAULT_VIP_PLANS.find(p => p.id === selectedPlan)?.price_usd || '24.99';

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-md">
      <div className="liquid-glass-card max-w-lg w-full p-6 sm:p-8 rounded-3xl border border-white/10 space-y-5 shadow-2xl relative text-start">
        <button
          onClick={onClose}
          className="absolute top-4 ltr:right-4 rtl:left-4 p-2 text-slate-400 hover:text-white rounded-xl"
        >
          <X size={16} />
        </button>

        <div className="text-center space-y-1">
          <div className="w-12 h-12 rounded-2xl bg-amber-500/20 text-amber-400 border border-amber-500/30 flex items-center justify-center mx-auto mb-2 shadow-lg shadow-amber-500/10">
            <Crown size={24} />
          </div>
          <h2 className="text-xl sm:text-2xl font-black text-white">{t('vip.title', 'Upgrade to VIP Access')}</h2>
          <p className="text-xs text-slate-400">{t('vip.subtitle', 'Unlock instant 4K streams, priority downloads, and exclusive releases.')}</p>
        </div>

        {invoice ? (
          <div className="space-y-4 text-center py-2">
            <div className="p-4 rounded-2xl bg-white/[0.03] border border-white/10 space-y-2 text-xs">
              <div className="text-slate-400">{t('vip_modal.invoice_created', { gateway: gatewayName }) || `Invoice Created (${gatewayName}):`}</div>
              <div className="font-mono text-cyan-400 font-bold text-sm" dir="ltr">{invoice.invoice_id || invoice.id || 'INV-29384'}</div>
              <div className="text-slate-300">{t('vip_modal.amount_label', 'Amount:')} <strong>${invoice.amount || '24.99'} {currency}</strong></div>
            </div>

            <a
              href={invoice.payment_url || invoice.checkout_url || `/api/v1/payments/mock-checkout?invoice_id=${invoice.invoice_id || 'INV-29384'}`}
              target="_blank"
              rel="noreferrer"
              className="liquid-button inline-flex items-center justify-center gap-2 w-full py-3 rounded-2xl text-sm font-bold text-white shadow-lg"
            >
              <Coins size={16} />
              <span>{t('vip_modal.complete_payment', { gateway: gatewayName }) || `Complete Payment in ${gatewayName}`}</span>
              <ExternalLink size={14} className="rtl:rotate-180" />
            </a>
          </div>
        ) : (
          <div className="space-y-4">
            {/* Plans List */}
            <div className="space-y-2">
              {DEFAULT_VIP_PLANS.map((plan) => {
                const isSelected = selectedPlan === plan.id;
                return (
                  <div
                    key={plan.id}
                    onClick={() => setSelectedPlan(plan.id)}
                    className={`p-3.5 rounded-2xl border cursor-pointer transition-all flex items-center justify-between gap-3 ${
                      isSelected
                        ? 'liquid-glass-card border-amber-500/50 bg-amber-500/10'
                        : 'bg-white/[0.02] border-white/5 hover:bg-white/[0.04]'
                    }`}
                  >
                    <div>
                      <div className="text-xs font-bold text-white flex items-center gap-1.5">
                        <span>{plan.name}</span>
                        {plan.is_popular && (
                          <span className="text-[9px] px-1.5 py-0.5 rounded-full bg-amber-500/20 text-amber-300 border border-amber-500/30 font-extrabold">
                            {t('vip_modal.popular_badge', 'Popular')}
                          </span>
                        )}
                      </div>
                      <div className="text-[11px] text-slate-400">
                        {plan.features[0]} • {plan.features[1]}
                      </div>
                    </div>

                    <div className="text-end shrink-0">
                      <div className="text-sm font-black text-amber-300" dir="ltr">${plan.price_usd}</div>
                      <div className="text-[10px] text-slate-500">{t('vip_modal.days_count', { days: plan.duration_days }) || `${plan.duration_days} days`}</div>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Currency Selector */}
            <div className="flex items-center justify-between text-xs">
              <span className="font-semibold text-slate-300">{t('vip_modal.payment_currency', 'Payment Currency:')}</span>
              <div className="flex items-center gap-1.5" dir="ltr">
                {['USDT', 'TON', 'BTC', 'ETH', 'USDC'].map((c) => (
                  <button
                    key={c}
                    type="button"
                    onClick={() => setCurrency(c)}
                    className={`px-2.5 py-1 rounded-xl text-xs font-bold transition-all ${
                      currency === c ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
                    }`}
                  >
                    {c}
                  </button>
                ))}
              </div>
            </div>

            <button
              onClick={handleCheckout}
              disabled={loading}
              className="liquid-button w-full py-3 rounded-2xl text-sm font-bold text-white shadow-lg flex items-center justify-center gap-2"
            >
              {loading ? <RefreshCw size={15} className="animate-spin" /> : <Coins size={15} />}
              <span>{loading ? t('vip_modal.initializing', 'Initializing Checkout...') : (t('vip_modal.pay_btn', { price: planPrice }) || `Pay $${planPrice} with Crypto`)}</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );
};
