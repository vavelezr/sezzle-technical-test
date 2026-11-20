import { operationForKey } from '../../domain/operations';
import { isDigit } from '../../domain/validation';
import type { CalculatorCommand } from './commands';

/**
 * Translates a KeyboardEvent.key value into the same command the matching
 * on-screen key dispatches. Returns undefined for unbound keys.
 */
export function commandForKey(key: string): CalculatorCommand | undefined {
  if (isDigit(key)) {
    return { type: 'DIGIT', value: key };
  }
  switch (key) {
    case '.':
    case ',':
      return { type: 'DECIMAL' };
    case 'Enter':
    case '=':
      return { type: 'EQUALS' };
    case 'Backspace':
      return { type: 'BACKSPACE' };
    case 'Escape':
    case 'Delete':
      return { type: 'CLEAR' };
  }
  const operation = operationForKey(key);
  return operation ? { type: 'OPERATOR', operation } : undefined;
}
