'use client';

import { LockKeyhole, Pause, Play, Trash2 } from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ENVIRONMENTS_STORAGE_KEY,
  deleteEnvironment,
  listActiveEnvironments,
  pauseEnvironment,
  readEnvironments,
  startEnvironment,
  writeEnvironments,
  type EnvironmentRecord,
} from '../lib/environments';
import { ROUTES } from '../config/routes';

const ENVIRONMENTS_CHANGED_EVENT = 'octra:environments-changed';

type EnvironmentPanelProps = {
  mode: 'active' | 'manage';
};

function getStorage(): Storage | null {
  return typeof window === 'undefined' ? null : window.localStorage;
}

function statusLabel(environment: EnvironmentRecord): string {
  return environment.active ? 'Active' : 'Paused';
}

export function EnvironmentPanel({ mode }: EnvironmentPanelProps) {
  const [environments, setEnvironments] = useState<EnvironmentRecord[]>([]);

  const reload = useCallback(() => {
    setEnvironments(readEnvironments(getStorage()));
  }, []);

  useEffect(() => {
    reload();
    const onStorage = (event: StorageEvent) => {
      if (!event.key || event.key === ENVIRONMENTS_STORAGE_KEY) reload();
    };
    window.addEventListener('storage', onStorage);
    window.addEventListener(ENVIRONMENTS_CHANGED_EVENT, reload);
    return () => {
      window.removeEventListener('storage', onStorage);
      window.removeEventListener(ENVIRONMENTS_CHANGED_EVENT, reload);
    };
  }, [reload]);

  const visibleEnvironments = useMemo(
    () => (mode === 'active' ? listActiveEnvironments(environments) : environments),
    [environments, mode],
  );

  const persist = useCallback((next: EnvironmentRecord[]) => {
    writeEnvironments(getStorage(), next);
    setEnvironments(next);
    window.dispatchEvent(new Event(ENVIRONMENTS_CHANGED_EVENT));
  }, []);

  const handlePause = (id: string) => {
    persist(pauseEnvironment(environments, id));
  };

  const handleStart = (id: string) => {
    persist(startEnvironment(environments, id));
  };

  const handleDelete = (id: string) => {
    persist(deleteEnvironment(environments, id));
  };

  if (visibleEnvironments.length === 0) {
    return (
      <p className="empty-flows-message">
        {mode === 'active' ? "You don't have any active environments." : "You don't have any environments yet."}
      </p>
    );
  }

  return (
    <div className="active-environment-list">
      {visibleEnvironments.map((environment) => (
        <article
          className={`active-environment-row${environment.active ? '' : ' is-paused'}`}
          key={environment.id}
        >
          <div>
            <span>{statusLabel(environment)}</span>
            <strong className="environment-name-line">
              <LockKeyhole size={14} aria-hidden="true" />
              {environment.name}
            </strong>
          </div>
          <div>
            <span>environment_id</span>
            <strong>{environment.id}</strong>
          </div>
          <div>
            <span>endpoint</span>
            <strong>{environment.endpoint}</strong>
          </div>
          <div>
            <span>cli_state</span>
            <strong>{environment.cliState}</strong>
          </div>
          <div className="environment-actions">
            {mode === 'active' ? (
              <button
                type="button"
                className="environment-icon-button"
                onClick={() => handlePause(environment.id)}
                aria-label={`Pause ${environment.name}`}
                title="Pause environment"
              >
                <Pause size={15} />
              </button>
            ) : (
              <>
                {environment.active ? (
                  <button
                    type="button"
                    className="environment-icon-button"
                    onClick={() => handlePause(environment.id)}
                    aria-label={`Pause ${environment.name}`}
                    title="Pause environment"
                  >
                    <Pause size={15} />
                  </button>
                ) : (
                  <button
                    type="button"
                    className="environment-icon-button"
                    onClick={() => handleStart(environment.id)}
                    aria-label={`Start ${environment.name}`}
                    title="Start environment"
                  >
                    <Play size={15} />
                  </button>
                )}
                <button
                  type="button"
                  className="environment-icon-button"
                  onClick={() => handleDelete(environment.id)}
                  aria-label={`Delete ${environment.name}`}
                  title="Delete environment"
                >
                  <Trash2 size={15} />
                </button>
              </>
            )}
            {mode === 'active' && (
              <a className="environment-manage-link" href={ROUTES.DASHBOARD_ENVIRONMENTS}>
                Manage
              </a>
            )}
          </div>
        </article>
      ))}
    </div>
  );
}
