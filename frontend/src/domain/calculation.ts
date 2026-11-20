import type { Operation } from './operations';

/** A calculation to be performed by the backend. */
export interface CalculationRequest {
  readonly operation: Operation;
  readonly operands: readonly number[];
}
