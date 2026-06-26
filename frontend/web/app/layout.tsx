import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { DesktopTitleBar } from './components/DesktopTitleBar';
import '@xyflow/react/dist/style.css';
import './globals.css';

export const metadata: Metadata = {
  title: 'Octra',
  description: 'A new Octra interface for planning, routing, and reviewing AI delivery work.',
  icons: {
    icon: '/assets/octra-node-logo.svg',
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
