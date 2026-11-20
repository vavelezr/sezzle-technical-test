import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { CalculatorClientContext } from '../../api/CalculatorClientContext';
import type { CalculatorClient } from '../../api/calculatorClient';
import { CalculatorError } from '../../domain/errors';
import type { Digit } from '../../domain/validation';
import type { CalculatorCommand } from './commands';
import { HISTORY_LIMIT, useCalculator } from './useCalculator';

type Calculate = CalculatorClient['calculate'];

function renderCalculator(calculate: Calculate) {
  const calculateMock = vi.fn(calculate);
  const client: CalculatorClient = { calculate: calculateMock };
  const wrapper = ({ children }: { children: ReactNode }) => (
    <CalculatorClientContext value={client}>{children}</CalculatorClientContext>
  );
  const hook = renderHook(() => useCalculator(), { wrapper });
  const press = (...commands: CalculatorCommand[]) => {
    act(() => {
      commands.forEach((command) => {
        hook.result.current.press(command);
      });
    });
  };
  return { ...hook, calculateMock, press };
}

const digit = (value: Digit): CalculatorCommand => ({ type: 'DIGIT', value });
const EQUALS: CalculatorCommand = { type: 'EQUALS' };
const CLEAR: CalculatorCommand = { type: 'CLEAR' };
const add = (a: Digit, b: Digit): CalculatorCommand[] => [
  digit(a),
  { type: 'OPERATOR', operation: 'add' },
  digit(b),
  EQUALS,
];

/** A promise that the test resolves or rejects manually. */
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: unknown) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe('useCalculator', () => {
  it('sends the calculation to the client and shows the result', async () => {
    const { result, calculateMock, press } = renderCalculator(() => Promise.resolve(5));

    press(...add('2', '3'));

    expect(result.current.isLoading).toBe(true);
    await waitFor(() => {
      expect(result.current.display.value).toBe('5');
    });
    expect(calculateMock).toHaveBeenCalledWith(
      { operation: 'add', operands: [2, 3] },
      expect.any(AbortSignal),
    );
    expect(result.current.isLoading).toBe(false);
    expect(result.current.history).toEqual([
      { id: 0, request: { operation: 'add', operands: [2, 3] }, result: 5 },
    ]);
  });

  it('exposes the pending operation', () => {
    const { result, press } = renderCalculator(() => Promise.resolve(0));

    press(digit('2'), { type: 'OPERATOR', operation: 'power' });

    expect(result.current.activeOperation).toBe('power');
  });

  it.each([
    [new CalculatorError('DIVISION_BY_ZERO', 'division by zero'), "You can't divide by zero"],
    [
      new CalculatorError('NETWORK_ERROR', 'offline'),
      "Can't reach the calculator service. Please try again.",
    ],
    [new Error('unexpected'), 'Something went wrong. Please try again.'],
  ])('shows a friendly message when the client fails with %o', async (error, message) => {
    const { result, press } = renderCalculator(() => Promise.reject(error));

    press(...add('1', '0'));

    await waitFor(() => {
      expect(result.current.display).toMatchObject({ value: message, isError: true });
    });
    expect(result.current.history).toEqual([]);
  });

  it('aborts the request and ignores its outcome when cleared', async () => {
    const pending = deferred<number>();
    let signal: AbortSignal | undefined;
    const { result, press } = renderCalculator((_request, abortSignal) => {
      signal = abortSignal;
      return pending.promise;
    });

    press(...add('1', '2'));
    press(CLEAR);
    expect(signal?.aborted).toBe(true);

    await act(async () => {
      pending.resolve(3);
      await pending.promise;
    });
    expect(result.current.state).toEqual({ status: 'idle' });
    expect(result.current.history).toEqual([]);
  });

  it('ignores a failure that arrives after the request was aborted', async () => {
    const pending = deferred<number>();
    const { result, calculateMock, press } = renderCalculator(() => pending.promise);

    press(...add('1', '2'));
    expect(calculateMock).toHaveBeenCalledTimes(1);
    press(CLEAR);
    await act(async () => {
      pending.reject(new Error('late'));
      await pending.promise.catch(() => undefined);
    });

    expect(result.current.state).toEqual({ status: 'idle' });
  });

  it(`keeps only the last ${String(HISTORY_LIMIT)} calculations, newest first`, async () => {
    const { result, press } = renderCalculator((request) =>
      Promise.resolve(request.operands.reduce((a, b) => a + b, 0)),
    );

    for (let i = 1; i <= HISTORY_LIMIT + 1; i++) {
      press(...add(String(i) as Digit, '0'));
      await waitFor(() => {
        expect(result.current.state.status).toBe('result');
      });
    }

    expect(result.current.history).toHaveLength(HISTORY_LIMIT);
    expect(result.current.history[0]?.result).toBe(HISTORY_LIMIT + 1);
    expect(result.current.history.at(-1)?.result).toBe(2);
  });

  it('throws a helpful error when no client is provided', () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    expect(() => renderHook(() => useCalculator())).toThrow(
      'useCalculatorClient must be used within a CalculatorClientContext provider',
    );
  });
});
