import { create } from 'zustand';
import { persist } from 'zustand/middleware';

// Token-usage statistics. The chart in Settings → Statistics is driven entirely
// by this store, which persists to localStorage so the numbers are global across
// sessions (issue #67). Usage is grouped by task category so the user can filter
// the chart by the kind of work that consumed the tokens.

// The four task categories surfaced in the UI map 1:1 to the backend task types
// emitted over the task websocket (orchestrator boss/classify.go):
//   code         -> development
//   research     -> search
//   document     -> document
//   presentation -> presentation
export type TaskCategory = 'search' | 'development' | 'presentation' | 'document';

export const TASK_CATEGORIES: TaskCategory[] = [
  'search',
  'development',
  'presentation',
  'document',
];

// Accent colour per category, reused by the chart so legend/bars stay consistent.
export const CATEGORY_COLORS: Record<TaskCategory, string> = {
  search: '#3b82f6', // blue
  development: '#f97316', // orange (matches the app accent)
  presentation: '#a855f7', // purple
  document: '#22c55e', // green
};

// A single recorded usage event. `tokens` is the number of tokens a finished task
// consumed; `timestamp` is epoch milliseconds so the data can be filtered by time
// later without a schema change.
export interface UsageRecord {
  category: TaskCategory;
  tokens: number;
  timestamp: number;
}

// Translate a raw backend task type (or any free-form hint) into one of the four
// UI categories. Unknown/empty values fall back to `development`, mirroring the
// orchestrator default where an unclassified task is treated as a code task.
export function mapTaskTypeToCategory(taskType: unknown): TaskCategory {
  const value = typeof taskType === 'string' ? taskType.trim().toLowerCase() : '';
  switch (value) {
    case 'research':
    case 'search':
      return 'search';
    case 'presentation':
      return 'presentation';
    case 'document':
    case 'text':
      return 'document';
    case 'code':
    case 'development':
    case '':
      return 'development';
    default:
      return 'development';
  }
}

interface StatisticsState {
  records: UsageRecord[];

  // Append a usage event. Non-positive or non-finite token counts are ignored so
  // a task that reports no usage never pollutes the chart with empty bars.
  recordUsage: (category: TaskCategory, tokens: number) => void;
  // Convenience wrapper that accepts a raw backend task type.
  recordTaskUsage: (taskType: unknown, tokens: number) => void;
  // Remove all recorded statistics (used by the "Clear" action in the UI).
  clearStatistics: () => void;
  // Total tokens per category, always returning every category (zero when unused)
  // so the chart renders a stable set of bars.
  getTotalsByCategory: () => Record<TaskCategory, number>;
  // Grand total across every category.
  getTotalTokens: () => number;
}

function emptyTotals(): Record<TaskCategory, number> {
  return { search: 0, development: 0, presentation: 0, document: 0 };
}

export const useStatisticsStore = create<StatisticsState>()(
  persist(
    (set, get) => ({
      records: [],

      recordUsage: (category, tokens) => {
        if (!Number.isFinite(tokens) || tokens <= 0) {
          return;
        }
        if (!TASK_CATEGORIES.includes(category)) {
          return;
        }
        set((state) => ({
          records: [
            ...state.records,
            { category, tokens: Math.round(tokens), timestamp: Date.now() },
          ],
        }));
      },

      recordTaskUsage: (taskType, tokens) => {
        get().recordUsage(mapTaskTypeToCategory(taskType), tokens);
      },

      clearStatistics: () => set({ records: [] }),

      getTotalsByCategory: () => {
        const totals = emptyTotals();
        for (const record of get().records) {
          if (totals[record.category] !== undefined) {
            totals[record.category] += record.tokens;
          }
        }
        return totals;
      },

      getTotalTokens: () =>
        get().records.reduce((sum, record) => sum + record.tokens, 0),
    }),
    {
      name: 'octra-token-statistics',
    }
  )
);
