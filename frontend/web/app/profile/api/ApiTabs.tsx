'use client';

import { KeyRound, Globe, Settings } from 'lucide-react';
import { usePathname } from 'next/navigation';
import { ROUTES } from '../../config/routes';

const tabs = [
  { slug: 'public', label: 'Public API', icon: Globe, href: `${ROUTES.PROFILE_API}/public` },
  { slug: 'api-keys', label: 'API keys', icon: KeyRound, href: `${ROUTES.PROFILE_API}/api-keys` },
  { slug: 'settings', label: 'Settings', icon: Settings, href: `${ROUTES.PROFILE_API}/settings` },
];

export function ApiTabs() {
  const pathname = usePathname();

  return (
    <div className="dashboard-tabs" aria-label="API sections" style={{ marginBottom: 14 }}>
      {tabs.map((t) => {
        const Icon = t.icon;
        const clean = pathname.replace(/\/+$/, '');
        const active = clean.endsWith(t.slug);
        return (
          <a className={active ? 'active' : ''} href={t.href} key={t.slug} style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            <Icon size={15} />
            {t.label}
          </a>
        );
      })}
    </div>
  );
}
