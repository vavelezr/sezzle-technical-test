import { Spinner } from '../../../components/ui/Spinner';
import { cx } from '../../../components/ui/cx';
import type { DisplayModel } from '../selectors';
import styles from './Display.module.css';

export interface DisplayProps {
  readonly display: DisplayModel;
  readonly isLoading: boolean;
}

const LONG_VALUE_LENGTH = 10;
const VERY_LONG_VALUE_LENGTH = 16;

function sizeClass({ value, isError }: DisplayModel): string | undefined {
  if (isError) {
    return styles.message;
  }
  if (value.length > VERY_LONG_VALUE_LENGTH) {
    return styles.veryLong;
  }
  return value.length > LONG_VALUE_LENGTH ? styles.long : undefined;
}

/** Shows the current expression and the value, result or error message. */
export function Display({ display, isLoading }: DisplayProps) {
  return (
    <section className={styles.display} aria-label="Display" aria-busy={isLoading}>
      <p className={styles.expression} data-testid="expression">
        {display.expression}
        {isLoading && (
          <>
            <Spinner />
            <span className="visually-hidden">Calculating</span>
          </>
        )}
      </p>
      <output
        className={cx(styles.value, display.isError && styles.error, sizeClass(display))}
        aria-live="polite"
        aria-atomic="true"
        data-testid="value"
      >
        {display.value}
      </output>
    </section>
  );
}
