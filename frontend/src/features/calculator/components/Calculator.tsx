import { useCalculator } from '../useCalculator';
import { useKeyboardCommands } from '../useKeyboardCommands';
import styles from './Calculator.module.css';
import { Display } from './Display';
import { History } from './History';
import { Keypad } from './Keypad';

/** Container component: connects the calculator logic to the presentational parts. */
export function Calculator() {
  const { display, isLoading, activeOperation, history, press } = useCalculator();
  useKeyboardCommands(press);

  return (
    <div className={styles.calculator}>
      <Display display={display} isLoading={isLoading} />
      <Keypad onCommand={press} activeOperation={activeOperation} isLoading={isLoading} />
      <History entries={history} />
    </div>
  );
}
