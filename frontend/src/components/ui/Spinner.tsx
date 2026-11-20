import styles from './Spinner.module.css';

/** Small decorative loading indicator; announce loading state separately. */
export function Spinner() {
  return <span className={styles.spinner} aria-hidden="true" data-testid="spinner" />;
}
