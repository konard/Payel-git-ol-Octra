import type { ReactNode } from 'react';
import { ProfileSidebar } from './ProfileSidebar';
import { ASSETS } from '../config/images';
import { ROUTES } from '../config/routes';

export default function ProfileLayout({ children }: { children: ReactNode }) {
  return (
    <main className="dashboard-page">
      <aside className="app-sidebar" aria-label="Profile sections">
        <a className="square-brand" href={ROUTES.WORKSPACE} aria-label="Octra home">
          <img src={ASSETS.LOGO} alt="" />
        </a>
        <nav>
          <ProfileSidebar />
        </nav>
      </aside>
      <section className="dashboard-scene" aria-labelledby="profile-title">
        {children}
      </section>
    </main>
  );
}
