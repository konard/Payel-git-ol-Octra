import { BarChart3, type LucideIcon } from 'lucide-react';
import styles from './EmptyDataPanel.module.css';

type EmptyDataPanelProps = {
  title: string;
  detail: string;
  actionHref?: string;
  actionLabel?: string;
  compact?: boolean;
  icon?: LucideIcon;
};

export function EmptyDataPanel({
  title,
  detail,
  actionHref,
  actionLabel,
  compact = false,
  icon: Icon = BarChart3,
}: EmptyDataPanelProps) {
  return (
    <div className={`${styles.panel} ${compact ? styles.compact : ''}`}>
      <span className={styles.icon} aria-hidden="true">
        <Icon size={20} />
      </span>
      <span className={styles.copy}>
        <strong>{title}</strong>
        <span>{detail}</span>
      </span>
      {actionHref && actionLabel ? (
        <a className={styles.action} href={actionHref}>
          {actionLabel}
        </a>
      ) : null}
    </div>
  );
}
