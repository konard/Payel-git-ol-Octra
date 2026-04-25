import { useEffect, useState } from 'react';
import { CheckCircle, ArrowRight } from 'lucide-react';
import { useAuthStore } from '../stores/authStore';

export function PaymentSuccess() {
  const { checkAuth } = useAuthStore();
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Обновляем статус аутентификации после оплаты
    const updateAuth = async () => {
      try {
        await checkAuth();
      } catch (error) {
        console.error('Failed to update auth after payment:', error);
      } finally {
        setIsLoading(false);
      }
    };

    updateAuth();
  }, [checkAuth]);

  const handleContinue = () => {
    window.location.href = '/';
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-[var(--background)]">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[var(--accent)] mx-auto mb-4"></div>
          <p className="text-[var(--text)]">Подтверждаем оплату...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-[var(--background)] px-4">
      <div className="max-w-md w-full bg-[var(--surface)] border border-[var(--border)] rounded-xl shadow-2xl p-8 text-center">
        {/* Success Icon */}
        <div className="w-16 h-16 bg-green-500 rounded-full flex items-center justify-center mx-auto mb-6">
          <CheckCircle size={32} className="text-white" />
        </div>

        {/* Title */}
        <h1 className="text-2xl font-bold text-[var(--text)] mb-2">
          Оплата прошла успешно!
        </h1>

        {/* Description */}
        <p className="text-[var(--text-secondary)] mb-8">
          Ваша подписка Pro активирована. Теперь у вас есть доступ ко всем функциям CrewAI.
        </p>

        {/* Features */}
        <div className="bg-[var(--background)] border border-[var(--border)] rounded-lg p-4 mb-6 text-left">
          <h3 className="font-medium text-[var(--text)] mb-3">Что теперь доступно:</h3>
          <ul className="space-y-2 text-sm text-[var(--text-secondary)]">
            <li className="flex items-center gap-2">
              <CheckCircle size={14} className="text-green-500 flex-shrink-0" />
              Доступ ко всем моделям ИИ
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle size={14} className="text-green-500 flex-shrink-0" />
              Неограниченные запросы
            </li>
            <li className="flex items-center gap-2">
              <CheckCircle size={14} className="text-green-500 flex-shrink-0" />
              Расширенные возможности генерации кода
            </li>
          </ul>
        </div>

        {/* Continue Button */}
        <button
          onClick={handleContinue}
          className="w-full flex items-center justify-center gap-2 bg-[var(--accent)] hover:bg-[var(--accent)]/90 text-white font-medium py-3 px-6 rounded-lg transition-colors"
        >
          Начать работу
          <ArrowRight size={18} />
        </button>
      </div>
    </div>
  );
}