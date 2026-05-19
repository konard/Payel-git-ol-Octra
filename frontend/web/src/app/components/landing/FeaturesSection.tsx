import { motion } from 'motion/react';
import { useI18n } from '../../../hooks/useI18n';

interface Feature {
  title: string;
  description: string;
  icon: string;
  color: string;
}

export function FeaturesSection() {
  const { t } = useI18n();

  const features: Feature[] = [
    {
      title: t('landing.features.visualEditor.title'),
      description: t('landing.features.visualEditor.desc'),
      icon: '⊞',
      color: '#ff6d5a',
    },
    {
      title: t('landing.features.hierarchy.title'),
      description: t('landing.features.hierarchy.desc'),
      icon: '⊿',
      color: '#5a9bff',
    },
    {
      title: t('landing.features.realtime.title'),
      description: t('landing.features.realtime.desc'),
      icon: '◉',
      color: '#50e3c2',
    },
    {
      title: t('landing.features.flexible.title'),
      description: t('landing.features.flexible.desc'),
      icon: '⚙',
      color: '#d97706',
    },
    {
      title: t('landing.features.templates.title'),
      description: t('landing.features.templates.desc'),
      icon: '⊡',
      color: '#8b5cf6',
    },
    {
      title: t('landing.features.export.title'),
      description: t('landing.features.export.desc'),
      icon: '⇄',
      color: '#ec4899',
    },
    {
      title: t('landing.features.providers.title'),
      description: t('landing.features.providers.desc'),
      icon: '◎',
      color: '#06b6d4',
    },
    {
      title: t('landing.features.scalability.title'),
      description: t('landing.features.scalability.desc'),
      icon: '⬡',
      color: '#f59e0b',
    },
  ];

  return (
    <section id="features" className="features-section">
      <div className="features-section__content">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="features-section__badge"
        >
          {t('landing.features.badge')}
        </motion.div>

        <motion.h2
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="features-section__title"
        >
          {t('landing.features.title')}
          <span className="features-section__title-accent">{t('landing.features.titleAccent')}</span>
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="features-section__description"
        >
          {t('landing.features.description')}
        </motion.p>

        <div className="features-section__grid">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.4, delay: index * 0.05 }}
              className="feature-card"
              whileHover={{ y: -5, transition: { duration: 0.2 } }}
            >
              <div
                className="feature-card__icon"
                style={{ color: feature.color }}
              >
                {feature.icon}
              </div>
              <h3 className="feature-card__title">{feature.title}</h3>
              <p className="feature-card__description">{feature.description}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
