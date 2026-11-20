import type { CalculatorState } from './calculatorReducer';
import { selectActiveOperation, selectDisplay, type DisplayModel } from './selectors';

describe('selectDisplay', () => {
  it.each<[string, CalculatorState, DisplayModel]>([
    ['idle', { status: 'idle' }, { expression: '', value: '0', isError: false }],
    [
      'entering the first operand keeps partial input',
      { status: 'enteringFirstOperand', input: '12.' },
      { expression: '', value: '12.', isError: false },
    ],
    [
      'operator selected',
      { status: 'operatorSelected', firstOperand: '12', operation: 'divide' },
      { expression: '12 ÷', value: '12', isError: false },
    ],
    [
      'operator selected after a long result formats it',
      { status: 'operatorSelected', firstOperand: '0.30000000000000004', operation: 'add' },
      { expression: '0.3 +', value: '0.3', isError: false },
    ],
    [
      'operator selected keeps a trailing decimal point',
      { status: 'operatorSelected', firstOperand: '2.', operation: 'power' },
      { expression: '2. ^', value: '2.', isError: false },
    ],
    [
      'entering the second operand',
      { status: 'enteringSecondOperand', firstOperand: '12', operation: 'divide', input: '4' },
      { expression: '12 ÷', value: '4', isError: false },
    ],
    [
      'loading a binary calculation',
      { status: 'loading', request: { operation: 'divide', operands: [12, 4] } },
      { expression: '12 ÷ 4 =', value: '…', isError: false },
    ],
    [
      'loading a unary calculation',
      { status: 'loading', request: { operation: 'sqrt', operands: [16] } },
      { expression: '√(16) =', value: '…', isError: false },
    ],
    [
      'result',
      { status: 'result', request: { operation: 'percentage', operands: [20, 50] }, value: 10 },
      { expression: '20 % of 50 =', value: '10', isError: false },
    ],
    [
      'error',
      {
        status: 'error',
        request: { operation: 'divide', operands: [1, 0] },
        message: "You can't divide by zero",
      },
      { expression: '1 ÷ 0 =', value: "You can't divide by zero", isError: true },
    ],
  ])('%s', (_name, state, expected) => {
    expect(selectDisplay(state)).toEqual(expected);
  });
});

describe('selectActiveOperation', () => {
  it.each<[string, CalculatorState, string | undefined]>([
    ['idle', { status: 'idle' }, undefined],
    [
      'operator selected',
      { status: 'operatorSelected', firstOperand: '1', operation: 'add' },
      'add',
    ],
    [
      'entering second operand',
      { status: 'enteringSecondOperand', firstOperand: '1', operation: 'power', input: '2' },
      'power',
    ],
    [
      'result',
      { status: 'result', request: { operation: 'add', operands: [1, 2] }, value: 3 },
      undefined,
    ],
  ])('%s', (_name, state, expected) => {
    expect(selectActiveOperation(state)).toBe(expected);
  });
});
