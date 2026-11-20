import { Button } from '../../../components/ui/Button';
import type { Operation } from '../../../domain/operations';
import type { CalculatorCommand } from '../commands';
import {
  BACKSPACE_KEY,
  CLEAR_KEY,
  DECIMAL_KEY,
  DIGIT_KEYS,
  EQUALS_KEY,
  OPERATION_KEYS,
  type KeyModel,
} from '../keypad';
import styles from './Keypad.module.css';

export interface KeypadProps {
  readonly onCommand: (command: CalculatorCommand) => void;
  readonly activeOperation?: Operation;
  readonly isLoading: boolean;
}

/** On-screen keys. Every key dispatches the same command as its keyboard shortcut. */
export function Keypad({ onCommand, activeOperation, isLoading }: KeypadProps) {
  const renderKey = (key: KeyModel, extra: { disabled?: boolean; pressed?: boolean } = {}) => (
    <Button
      key={key.id}
      variant={key.variant}
      aria-label={key.ariaLabel}
      aria-pressed={extra.pressed}
      disabled={extra.disabled}
      className={styles[key.id]}
      onClick={() => {
        onCommand(key.command);
      }}
    >
      {key.label}
    </Button>
  );

  return (
    <div className={styles.keypad}>
      <div className={styles.operations} role="group" aria-label="Operations">
        {OPERATION_KEYS.map((key) => renderKey(key, { pressed: key.id === activeOperation }))}
        {renderKey(CLEAR_KEY)}
      </div>
      <div className={styles.numbers} role="group" aria-label="Number pad">
        {DIGIT_KEYS.map((key) => renderKey(key))}
        {renderKey(DECIMAL_KEY)}
        {renderKey(BACKSPACE_KEY)}
        {renderKey(EQUALS_KEY, { disabled: isLoading })}
      </div>
    </div>
  );
}
