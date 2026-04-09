/**
 * SubscriptionModal Component
 * Модальное окно выбора подписки в стиле Perplexity
 */

import { useState, useEffect } from 'react';
import { Check } from 'lucide-react';
import {
  getSubscriptionPlans,
  subscribeUser,
  activatePromoCode,
  type SubscriptionPlan,
} from '../services/subscriptionService';
import { useAuthStore } from '../stores/authStore';

interface SubscriptionModalProps {
  onClose: () => void;
  onSubscribe: () => void;
}

type PaymentMethod = 'card' | 'promo';

const PLAN_FEATURES: Record<string, string[]> = {
  monthly: [
    'Доступ ко всем моделям ИИ',
    'Неограниченные запросы',
    'Приоритетная обработка задач',
    'Расширенные возможности генерации кода',
  ],
  quarterly: [
    'Всё из тарифа 1 месяц и:',
    'Экономия 16% по сравнению с помесячной оплатой',
    'Приоритетная поддержка',
    'Ранний доступ к новым функциям',
  ],
  semi_annually: [
    'Всё из тарифа 3 месяца и:',
    'Экономия 25% по сравнению с помесячной оплатой',
    'Персональная настройка агентов',
    'Расширенная аналитика',
  ],
  annually: [
    'Всё из тарифа 6 месяцев и:',
    'Экономия 33% — лучшая цена',
    'Выделенные ресурсы',
    'Приоритетный доступ к новым моделям ИИ',
  ],
};

export function SubscriptionModal({ onClose, onSubscribe }: SubscriptionModalProps) {
  const { user } = useAuthStore();
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [selectedPlan, setSelectedPlan] = useState<string | null>(null);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('card');
  const [promoCode, setPromoCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [step, setStep] = useState<'plans' | 'payment'>('plans');

  useEffect(() => {
    loadPlans();
  }, []);

  const loadPlans = async () => {
    try {
      const plansData = await getSubscriptionPlans();
      setPlans(plansData);
    } catch {
      // Use defaults
      setPlans([
        { id: 'monthly', name: '1 месяц', duration: 30, price: '990' },
        { id: 'quarterly', name: '3 месяца', duration: 90, price: '2490' },
        { id: 'semi_annually', name: '6 месяцев', duration: 180, price: '4490' },
        { id: 'annually', name: '1 год', duration: 365, price: '7990' },
      ]);
    }
  };

  const handleSelectPlan = (planId: string) => {
    setSelectedPlan(planId);
    setStep('payment');
  };

  const handleSubscribe = async () => {
    if (!selectedPlan || !user) return;
    setIsLoading(true);
    setError('');
    try {
      await subscribeUser(user.id, selectedPlan);
      onSubscribe();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка при оформлении подписки');
    } finally {
      setIsLoading(false);
    }
  };

  const handlePromoCode = async () => {
    if (!promoCode.trim() || !user) return;
    setIsLoading(true);
    setError('');
    try {
      await activatePromoCode(user.id, promoCode.trim());
      onSubscribe();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка при активации промокода');
    } finally {
      setIsLoading(false);
    }
  };

  const getPlanName = (planId: string) => {
    return plans.find((p) => p.id === planId)?.name || '';
  };

  const getPlanPrice = (planId: string) => {
    return plans.find((p) => p.id === planId)?.price || '0';
  };

  // Payment step
  if (step === 'payment') {
    return (
      <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
        <div
          className="relative bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-md mx-4 flex flex-col overflow-hidden"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border)]">
            <h2 className="text-lg font-medium text-[var(--text)]">
              Оформление: {getPlanName(selectedPlan || '')}
            </h2>
            <button onClick={onClose} className="p-1 hover:bg-[var(--background)] rounded-md transition-colors">
              <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <line x1="1" y1="1" x2="13" y2="13" />
                <line x1="13" y1="1" x2="1" y2="13" />
              </svg>
            </button>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-5 space-y-5">
            {error && (
              <div className="p-3 bg-red-500/10 border border-red-500 text-red-500 rounded-md text-sm">{error}</div>
            )}

            {/* Price summary */}
            <div className="p-4 bg-[var(--background)] border border-[var(--border)] rounded-lg">
              <p className="text-sm text-[var(--text-secondary)]">Тариф</p>
              <p className="text-base font-medium text-[var(--text)] mt-0.5">{getPlanName(selectedPlan || '')}</p>
              <div className="w-full h-px bg-[var(--border)] my-3" />
              <div className="flex items-end gap-1.5">
                <span className="text-2xl font-medium text-[var(--text)]">{getPlanPrice(selectedPlan || '')}₽</span>
              </div>
            </div>

            {/* Payment method tabs */}
            <div className="flex gap-2">
              <button
                onClick={() => setPaymentMethod('card')}
                className={`flex-1 py-2.5 px-3 rounded-lg border text-sm font-medium transition-all ${
                  paymentMethod === 'card'
                    ? 'border-[var(--accent)] bg-[var(--accent)]/10 text-[var(--accent)]'
                    : 'border-[var(--border)] bg-[var(--background)] text-[var(--text-secondary)] hover:text-[var(--text)]'
                }`}
              >
                💳 Карта
              </button>
              <button
                onClick={() => setPaymentMethod('promo')}
                className={`flex-1 py-2.5 px-3 rounded-lg border text-sm font-medium transition-all ${
                  paymentMethod === 'promo'
                    ? 'border-[var(--accent)] bg-[var(--accent)]/10 text-[var(--accent)]'
                    : 'border-[var(--border)] bg-[var(--background)] text-[var(--text-secondary)] hover:text-[var(--text)]'
                }`}
              >
                🎟️ Промокод
              </button>
            </div>

            {paymentMethod === 'card' ? (
              <button
                onClick={handleSubscribe}
                disabled={isLoading}
                className="w-full py-2.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-lg transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {isLoading ? 'Оформление...' : `Оплатить ${getPlanPrice(selectedPlan || '')}₽`}
              </button>
            ) : (
              <div className="space-y-3">
                <input
                  type="text"
                  value={promoCode}
                  onChange={(e) => setPromoCode(e.target.value)}
                  placeholder="Введите промокод"
                  className="w-full px-3 py-2.5 bg-[var(--background)] border border-[var(--border)] rounded-lg text-[var(--text)] text-sm placeholder:text-[var(--text-muted)] focus:outline-none focus:border-[var(--accent)] transition-colors"
                  onKeyDown={(e) => e.key === 'Enter' && handlePromoCode()}
                />
                <button
                  onClick={handlePromoCode}
                  disabled={isLoading || !promoCode.trim()}
                  className="w-full py-2.5 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium rounded-lg transition-colors text-sm disabled:opacity-60 disabled:cursor-not-allowed"
                >
                  {isLoading ? 'Активация...' : 'Активировать'}
                </button>
              </div>
            )}

            <button onClick={() => setStep('plans')} className="w-full py-2 text-sm text-[var(--text-secondary)] hover:text-[var(--text)] transition-colors">
              ← Назад к тарифам
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Plans step — Perplexity-style cards
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60" onClick={onClose}>
      <div
        className="relative bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl w-full max-w-5xl mx-4 flex flex-col overflow-hidden max-h-[90vh]"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-[var(--border)]">
          <div>
            <h2 className="text-xl font-medium text-[var(--text)]">Выберите подписку</h2>
            <p className="text-sm text-[var(--text-secondary)] mt-1">Расширенные ответы и лучшие модели ИИ</p>
          </div>
          <button onClick={onClose} className="p-1 hover:bg-[var(--background)] rounded-md transition-colors">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <line x1="1" y1="1" x2="13" y2="13" />
              <line x1="13" y1="1" x2="1" y2="13" />
            </svg>
          </button>
        </div>

        {/* Cards grid */}
        <div className="flex-1 overflow-y-auto p-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            {plans.map((plan, index) => {
              const isPopular = plan.id === 'quarterly' || plan.id === 'annually';
              const features = PLAN_FEATURES[plan.id] || [];

              return (
                <button
                  key={plan.id}
                  onClick={() => handleSelectPlan(plan.id)}
                  className={`flex flex-col p-5 rounded-xl border transition-all text-left hover:shadow-lg cursor-pointer
                    ${isPopular
                      ? 'border-[var(--accent)] bg-gradient-to-b from-[var(--accent)]/5 to-transparent'
                      : 'border-[var(--border)] bg-[var(--background)] hover:border-[var(--accent)]/50'
                    }`}
                >
                  {/* Header */}
                  <div className="flex items-start justify-between mb-3">
                    <div className="font-medium text-[var(--text)] text-base">{plan.name}</div>
                    {isPopular && (
                      <div className="bg-[var(--accent)]/10 text-[var(--accent)] text-xs font-medium px-2 py-0.5 rounded-md">
                        Популярное
                      </div>
                    )}
                  </div>

                  {/* Description */}
                  <p className="text-xs text-[var(--text-secondary)] mb-3">
                    {index === 0 && 'Базовый доступ ко всем моделям ИИ'}
                    {index === 1 && 'Расширенные ответы и лучшие модели ИИ'}
                    {index === 2 && 'Оптимальный баланс цены и возможностей'}
                    {index === 3 && 'Максимальное использование и лучшая производительность'}
                  </p>

                  {/* Price */}
                  <div className="mb-3">
                    <div className="flex items-end gap-1.5">
                      <span className="text-2xl font-medium text-[var(--text)]">{plan.price}₽</span>
                    </div>
                    <span className="text-xs text-[var(--text-muted)]">
                      {plan.duration} дней
                    </span>
                  </div>

                  {/* Divider */}
                  <div className="w-full h-px bg-[var(--border)] mb-3" />

                  {/* Features */}
                  <div className="mb-4">
                    <p className="text-xs text-[var(--text-muted)] mb-2">
                      {index === 0 ? 'Включает:' : `Всё из ${plans[index - 1]?.name} и:`}
                    </p>
                    <ul className="flex flex-col gap-2">
                      {features.map((feature, i) => (
                        <li key={i} className="flex flex-row gap-2 text-[var(--text-secondary)]">
                          <Check size={14} className="shrink-0 mt-0.5 text-[var(--accent)]" />
                          <span className="text-xs">{feature}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  {/* Button */}
                  <div className="mt-auto pt-3">
                    <div className={`w-full py-2 px-3 rounded-lg text-sm font-medium text-center transition-colors
                      ${isPopular
                        ? 'bg-[var(--accent)] text-white hover:bg-[var(--accent)]/90'
                        : 'bg-[var(--surface)] border border-[var(--border)] text-[var(--text)] hover:border-[var(--accent)]'
                      }`}>
                      Выбрать
                    </div>
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
