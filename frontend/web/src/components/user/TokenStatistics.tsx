import { useMemo, useState } from 'react';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  Legend,
  type ChartData,
  type ChartOptions,
} from 'chart.js';
import { Bar } from 'react-chartjs-2';
import { BarChart3, Trash2 } from 'lucide-react';
import {
  useStatisticsStore,
  TASK_CATEGORIES,
  CATEGORY_COLORS,
  type TaskCategory,
} from '../../stores/statisticsStore';
import { useThemeStore } from '../../stores/themeStore';
import { useI18n } from '../../hooks/useI18n';

// Chart.js is tree-shakeable: only the pieces the bar chart needs are registered
// so the bundle does not pull in the whole library.
ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip, Legend);

type CategoryFilter = TaskCategory | 'all';

/**
 * Settings → Statistics panel. Renders a Chart.js bar chart of the tokens spent
 * per task category, with a filter so the user can focus on a single kind of work
 * (search, development, presentation, document). Data comes from the persisted
 * statisticsStore, so the numbers are global across sessions (issue #67).
 */
export function TokenStatistics() {
  const { t } = useI18n();
  const isDark = useThemeStore((state) => state.isDark);
  // Subscribe to the raw records so the chart re-renders whenever a task reports
  // new usage; totals are derived below.
  const records = useStatisticsStore((state) => state.records);
  const clearStatistics = useStatisticsStore((state) => state.clearStatistics);
  const [filter, setFilter] = useState<CategoryFilter>('all');

  const totals = useMemo(() => {
    const result: Record<TaskCategory, number> = {
      search: 0,
      development: 0,
      presentation: 0,
      document: 0,
    };
    for (const record of records) {
      if (result[record.category] !== undefined) {
        result[record.category] += record.tokens;
      }
    }
    return result;
  }, [records]);

  const totalTokens = useMemo(
    () => TASK_CATEGORIES.reduce((sum, category) => sum + totals[category], 0),
    [totals],
  );

  const visibleCategories = filter === 'all' ? TASK_CATEGORIES : [filter];

  const categoryLabel = (category: TaskCategory) =>
    t(`statistics.categories.${category}`);

  const chartData: ChartData<'bar'> = useMemo(
    () => ({
      labels: visibleCategories.map(categoryLabel),
      datasets: [
        {
          label: t('statistics.chartLabel'),
          data: visibleCategories.map((category) => totals[category]),
          backgroundColor: visibleCategories.map((category) => CATEGORY_COLORS[category]),
          borderRadius: 6,
          maxBarThickness: 64,
        },
      ],
    }),
    // categoryLabel/t are stable enough for this view; recompute on data + filter.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [visibleCategories.join(','), totals, t],
  );

  const axisColor = isDark ? '#9ca3af' : '#6b7280';
  const gridColor = isDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.06)';

  const chartOptions: ChartOptions<'bar'> = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        callbacks: {
          label: (context) =>
            ` ${context.parsed.y.toLocaleString()} ${t('statistics.tokensAxis')}`,
        },
      },
    },
    scales: {
      x: {
        ticks: { color: axisColor },
        grid: { display: false },
      },
      y: {
        beginAtZero: true,
        ticks: {
          color: axisColor,
          callback: (value) => Number(value).toLocaleString(),
        },
        grid: { color: gridColor },
      },
    },
  };

  const hasData = totalTokens > 0;

  return (
    <div className="space-y-4 max-w-2xl">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-sm font-medium text-[var(--text)] mb-1 flex items-center gap-2">
            <BarChart3 size={16} />
            {t('statistics.title')}
          </div>
          <div className="text-xs text-[var(--text-muted)]">
            {t('statistics.description')}
          </div>
        </div>
        {hasData && (
          <button
            onClick={clearStatistics}
            className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium bg-[var(--background)] border border-[var(--border)] rounded-md text-[var(--text-muted)] hover:border-red-500 hover:text-red-500 transition-colors flex-shrink-0"
            title={t('statistics.clear')}
          >
            <Trash2 size={14} />
            {t('statistics.clear')}
          </button>
        )}
      </div>

      {/* Total summary */}
      <div className="rounded-lg border border-[var(--border)] bg-[var(--background)] px-4 py-3">
        <div className="text-xs text-[var(--text-muted)]">{t('statistics.total')}</div>
        <div className="text-2xl font-semibold text-[var(--text)]">
          {totalTokens.toLocaleString()}
        </div>
      </div>

      {/* Category filter */}
      <div className="flex flex-wrap gap-2">
        <FilterButton
          active={filter === 'all'}
          onClick={() => setFilter('all')}
          label={t('statistics.filterAll')}
        />
        {TASK_CATEGORIES.map((category) => (
          <FilterButton
            key={category}
            active={filter === category}
            onClick={() => setFilter(category)}
            label={categoryLabel(category)}
            color={CATEGORY_COLORS[category]}
          />
        ))}
      </div>

      {/* Chart / empty state */}
      {hasData ? (
        <div className="h-64 rounded-lg border border-[var(--border)] bg-[var(--background)] p-3">
          <Bar data={chartData} options={chartOptions} />
        </div>
      ) : (
        <div className="h-64 flex flex-col items-center justify-center gap-2 rounded-lg border border-dashed border-[var(--border)] bg-[var(--background)] text-[var(--text-muted)]">
          <BarChart3 size={28} />
          <div className="text-sm">{t('statistics.empty')}</div>
        </div>
      )}
    </div>
  );
}

interface FilterButtonProps {
  active: boolean;
  onClick: () => void;
  label: string;
  color?: string;
}

function FilterButton({ active, onClick, label, color }: FilterButtonProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-xs font-medium border transition-colors ${
        active
          ? 'border-[var(--accent)] bg-[var(--accent)]/10 text-[var(--accent)]'
          : 'border-[var(--border)] bg-[var(--background)] text-[var(--text-muted)] hover:border-[var(--accent)]/50'
      }`}
    >
      {color && (
        <span
          className="w-2.5 h-2.5 rounded-full flex-shrink-0"
          style={{ backgroundColor: color }}
        />
      )}
      {label}
    </button>
  );
}
