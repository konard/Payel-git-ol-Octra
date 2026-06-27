import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { DesktopTitleBar } from './components/DesktopTitleBar';
import { ASSETS } from './config/images';
import '@xyflow/react/dist/style.css';
import './styles/tokens.css';
import './globals.css';

export const metadata: Metadata = {
  title: 'Octra',
  description: 'A new Octra interface for planning, routing, and reviewing AI delivery work.',
  icons: {
    icon: ASSETS.LOGO,
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>
        <DesktopTitleBar />
        <div className="canvas-grid" aria-hidden="true" />
        <div className="layout-canvas">
          {children}
        </div>
      </body>
    </html>
  );
}
