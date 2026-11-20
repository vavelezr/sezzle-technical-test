import { useCallback, useEffect, useReducer, useRef, useState } from 'react';
import { useCalculatorClient } from '../../api/CalculatorClientContext';
import type { CalculationRequest } from '../../domain/calculation';
import { CalculatorError, friendlyErrorMessage } from '../../domain/errors';
import type { BinaryOperation } from '../../domain/operations';
import { calculatorReducer, initialState, type CalculatorState } from './calculatorReducer';
import type { CalculatorCommand } from './commands';
import { selectActiveOperation, selectDisplay, type DisplayModel } from './selectors';

/** Number of past calculations kept in the session history. */
export const HISTORY_LIMIT = 5;

export interface HistoryEntry {
  readonly id: number;
  readonly request: CalculationRequest;
  readonly result: number;
}

export interface UseCalculatorResult {
  readonly state: CalculatorState;
  readonly display: DisplayModel;
  readonly isLoading: boolean;
  readonly activeOperation: BinaryOperation | undefined;
  readonly history: readonly HistoryEntry[];
  readonly press: (command: CalculatorCommand) => void;
}

function toErrorMessage(error: unknown): string {
  return friendlyErrorMessage(error instanceof CalculatorError ? error.code : 'UNKNOWN_ERROR');
}

/**
 * Facade over the calculator: hides the reducer, the API client and the
 * request lifecycle behind `state` and `press(command)`.
 *
 * Entering the loading state triggers the request. Leaving it (e.g. CLEAR)
 * aborts the request, so a stale response can never overwrite newer state.
 */
export function useCalculator(): UseCalculatorResult {
  const client = useCalculatorClient();
  const [state, dispatch] = useReducer(calculatorReducer, initialState);
  const [history, setHistory] = useState<readonly HistoryEntry[]>([]);
  const nextHistoryId = useRef(0);

  const pendingRequest = state.status === 'loading' ? state.request : null;

  useEffect(() => {
    if (!pendingRequest) {
      return;
    }
    const controller = new AbortController();

    client.calculate(pendingRequest, controller.signal).then(
      (result) => {
        if (controller.signal.aborted) {
          return;
        }
        const entry = { id: nextHistoryId.current++, request: pendingRequest, result };
        setHistory((previous) => [entry, ...previous].slice(0, HISTORY_LIMIT));
        dispatch({ type: 'CALCULATION_SUCCEEDED', result });
      },
      (error: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        dispatch({ type: 'CALCULATION_FAILED', message: toErrorMessage(error) });
      },
    );

    return () => {
      controller.abort();
    };
  }, [client, pendingRequest]);

  const press = useCallback((command: CalculatorCommand) => {
    dispatch(command);
  }, []);

  return {
    state,
    display: selectDisplay(state),
    isLoading: state.status === 'loading',
    activeOperation: selectActiveOperation(state),
    history,
    press,
  };
}
