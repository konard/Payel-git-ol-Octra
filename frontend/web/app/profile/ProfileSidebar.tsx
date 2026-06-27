'use client';

import { User, Settings, KeyRound } from 'lucide-react';
import { ROUTES } from '../config/routes';
import { usePathname } from 'next/navigation';

const items = [
  { slug: '', label: 'Profile', href: ROUTES.PROFILE, icon: User },
  { slug: 'settings', label: 'Settings', href: ROUTES.PROFILE_SETTINGS, icon: Settings },
  { slug: 'api', label: 'API', href: ROUTES.PROFILE_API, icon: KeyRound },
];

export function ProfileSidebar() {
  const pathname = usePathname();

  return items.map((item) => {
    const Icon = item.icon;
    const active = pathname === item.href || (item.slug && pathname.endsWith(item.slug));
    return (
      <a className={`side-icon${active ? ' active' : ''}`} href={item.href} key={item.slug} aria-label={item.label}>
        <Icon size={18} />
      </a>
    );
  });
}
