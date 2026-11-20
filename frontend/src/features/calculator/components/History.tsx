import { formatNumber } from '../../../domain/format';
import { formatExpression } from '../selectors';
import type { HistoryEntry } from '../useCalculator';
import styles from './History.module.css';

export interface HistoryProps {
  readonly entries: readonly HistoryEntry[];
}

/** Most recent successful calculations of this session, newest first. */
export function History({ entries }: HistoryProps) {
  if (entries.length === 0) {
    return null;
  }
  return (
    <section className={styles.history} aria-labelledby="history-heading">
      <h2 id="history-heading" className={styles.heading}>
        Recent
      </h2>
      <ol className={styles.list}>
        {entries.map((entry) => (
          <li key={entry.id} className={styles.item}>
            <span className={styles.expression}>{formatExpression(entry.request)} =</span>
            <span className={styles.result}>{formatNumber(entry.result)}</span>
          </li>
        ))}
      </ol>
    </section>
  );
}
