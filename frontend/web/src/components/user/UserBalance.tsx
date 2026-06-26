import { Wallet } from 'lucide-react';
import { useAuthStore } from '../../stores/authStore';

export function UserBalance() {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated || !user) {
    return null;
  }

  const balance = user.balance_credits ?? 0;

  return (
    <div className="user-balance" title="Account balance" aria-label={`Account balance: ${balance} credits`}>
      <span className="user-balance__icon" aria-hidden="true">
        <Wallet size={15} />
      </span>
      <span className="user-balance__content">
        <span className="user-balance__label">Balance</span>
        <span className="user-balance__value">{balance.toLocaleString()} credits</span>
      </span>
    </div>
  );
}
