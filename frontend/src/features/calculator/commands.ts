import type { Operation } from '../../domain/operations';
import type { Digit } from '../../domain/validation';

/**
 * User commands (Command pattern). Every on-screen button and keyboard key is
 * translated into one of these, so both input methods share the same logic.
 */
export type CalculatorCommand =
  | { readonly type: 'DIGIT'; readonly value: Digit }
  | { readonly type: 'DECIMAL' }
  | { readonly type: 'OPERATOR'; readonly operation: Operation }
  | { readonly type: 'EQUALS' }
  | { readonly type: 'BACKSPACE' }
  | { readonly type: 'CLEAR' };

/** Outcomes of a backend calculation, dispatched by the useCalculator hook. */
export type CalculationEvent =
  | { readonly type: 'CALCULATION_SUCCEEDED'; readonly result: number }
  | { readonly type: 'CALCULATION_FAILED'; readonly message: string };

export type CalculatorAction = CalculatorCommand | CalculationEvent;
