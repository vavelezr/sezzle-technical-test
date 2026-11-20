import type { ButtonVariant } from '../../components/ui/Button';
import { OPERATIONS, type OperationMeta } from '../../domain/operations';
import type { Digit } from '../../domain/validation';
import type { CalculatorCommand } from './commands';

/** View model of one calculator key. */
export interface KeyModel {
  readonly id: string;
  readonly label: string;
  /** Accessible name, when the visible label is a symbol. */
  readonly ariaLabel?: string;
  readonly command: CalculatorCommand;
  readonly variant: ButtonVariant;
}

function createOperationKey(meta: OperationMeta): KeyModel {
  return {
    id: meta.id,
    label: meta.symbol,
    ariaLabel: meta.label,
    command: { type: 'OPERATOR', operation: meta.id },
    variant: 'operator',
  };
}

function createDigitKey(value: Digit): KeyModel {
  return {
    id: `digit-${value}`,
    label: value,
    command: { type: 'DIGIT', value },
    variant: 'digit',
  };
}

export const CLEAR_KEY: KeyModel = {
  id: 'clear',
  label: 'C',
  ariaLabel: 'Clear',
  command: { type: 'CLEAR' },
  variant: 'action',
};

export const BACKSPACE_KEY: KeyModel = {
  id: 'backspace',
  label: '⌫',
  ariaLabel: 'Backspace',
  command: { type: 'BACKSPACE' },
  variant: 'action',
};

export const DECIMAL_KEY: KeyModel = {
  id: 'decimal',
  label: '.',
  ariaLabel: 'Decimal point',
  command: { type: 'DECIMAL' },
  variant: 'digit',
};

export const EQUALS_KEY: KeyModel = {
  id: 'equals',
  label: '=',
  ariaLabel: 'Equals',
  command: { type: 'EQUALS' },
  variant: 'equals',
};

/** Operator keys, generated from the operation config (Factory). */
export const OPERATION_KEYS: readonly KeyModel[] = OPERATIONS.map(createOperationKey);

/** Digit keys in keypad order; placement is handled by CSS grid areas. */
export const DIGIT_KEYS: readonly KeyModel[] = (
  ['7', '8', '9', '4', '5', '6', '1', '2', '3', '0'] as const
).map(createDigitKey);
