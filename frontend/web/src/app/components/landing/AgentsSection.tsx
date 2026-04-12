import { motion } from 'motion/react';
import { useThemeStore } from '../../../stores/themeStore';
import workflowFull from '../../../images/main/dark/workfloy/workflow_dark_theme.png';
import workflowFullLight from '../../../images/main/light/main_screen.png';

export function AgentsSection() {
  const { isDark } = useThemeStore();

  return (
    <section id="agents" className="py-20 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Image */}
          <motion.div
            initial={{ opacity: 0, x: -30 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="relative"
          >
            <div className="relative rounded-xl overflow-hidden shadow-2xl border border-[var(--border)] bg-[var(--surface)]">
              <img
                src={isDark ? workflowFull : workflowFullLight}
                alt="CrewAI Full Workflow"
                className="w-full h-auto"
              />
            </div>
          </motion.div>

          {/* Text content */}
          <div>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5 }}
              className="inline-flex items-center gap-2 px-4 py-2 bg-[var(--accent)]/10 border border-[var(--accent)]/20 rounded-full text-sm font-medium text-[var(--accent)] mb-6"
            >
              Иерархия агентов
            </motion.div>

            <motion.h2
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="text-4xl sm:text-5xl font-bold text-[var(--text)] mb-6"
            >
              Три уровня
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-500 to-orange-600"> управления</span>
            </motion.h2>

            <motion.p
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="text-lg text-[var(--text-muted)] mb-8"
            >
              Организуйте AI-агентов в эффективную иерархическую структуру для оптимального управления задачами
            </motion.p>

            <motion.div
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="space-y-6"
            >
              {/* Boss */}
              <div className="flex items-start gap-4 p-4 rounded-lg bg-[var(--surface)] border border-[var(--border)]">
                <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-orange-500 to-orange-600 rounded-lg flex items-center justify-center text-white font-bold text-lg">
                  B
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-[var(--text)] mb-1">Boss</h3>
                  <p className="text-[var(--text-muted)]">Главный координатор, распределяет задачи между менеджерами</p>
                </div>
              </div>

              {/* Manager */}
              <div className="flex items-start gap-4 p-4 rounded-lg bg-[var(--surface)] border border-[var(--border)]">
                <div className="flex-shrink-0 w-12 h-12 bg-[var(--accent)]/10 border border-[var(--accent)]/20 rounded-lg flex items-center justify-center text-[var(--accent)] font-bold text-lg">
                  M
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-[var(--text)] mb-1">Manager</h3>
                  <p className="text-[var(--text-muted)]">Управляет группой воркеров, контролирует процесс выполнения</p>
                </div>
              </div>

              {/* Worker */}
              <div className="flex items-start gap-4 p-4 rounded-lg bg-[var(--surface)] border border-[var(--border)]">
                <div className="flex-shrink-0 w-12 h-12 bg-[var(--surface)] border border-[var(--border)] rounded-lg flex items-center justify-center text-[var(--text)] font-bold text-lg">
                  W
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-[var(--text)] mb-1">Worker</h3>
                  <p className="text-[var(--text-muted)]">Выполняет конкретные задачи, специализируется на определённых операциях</p>
                </div>
              </div>
            </motion.div>
          </div>
        </div>
      </div>
    </section>
  );
}
