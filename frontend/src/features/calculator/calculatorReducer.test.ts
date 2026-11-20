import type { BinaryOperation, Operation } from '../../domain/operations';
import { MAX_INPUT_DIGITS, type Digit } from '../../domain/validation';
import { calculatorReducer, initialState, type CalculatorState } from './calculatorReducer';
import type { CalculatorAction, CalculatorCommand } from './commands';

const digit = (value: Digit): CalculatorCommand => ({ type: 'DIGIT', value });
const op = (operation: Operation): CalculatorCommand => ({ type: 'OPERATOR', operation });
const DECIMAL: CalculatorCommand = { type: 'DECIMAL' };
const EQUALS: CalculatorCommand = { type: 'EQUALS' };
const BACKSPACE: CalculatorCommand = { type: 'BACKSPACE' };
const CLEAR: CalculatorCommand = { type: 'CLEAR' };
const succeeded = (result: number): CalculatorAction => ({ type: 'CALCULATION_SUCCEEDED', result });
const failed = (message: string): CalculatorAction => ({ type: 'CALCULATION_FAILED', message });

function run(actions: CalculatorAction[], from: CalculatorState = initialState): CalculatorState {
  return actions.reduce(calculatorReducer, from);
}

const enteringFirst = (input: string): CalculatorState => ({
  status: 'enteringFirstOperand',
  input,
});
const selected = (firstOperand: string, operation: BinaryOperation): CalculatorState => ({
  status: 'operatorSelected',
  firstOperand,
  operation,
});
const enteringSecond = (
  firstOperand: string,
  operation: BinaryOperation,
  input: string,
): CalculatorState => ({ status: 'enteringSecondOperand', firstOperand, operation, input });
const divideRequest = { operation: 'divide', operands: [12, 4] } as const;
const loading: CalculatorState = { status: 'loading', request: divideRequest };
const result: CalculatorState = { status: 'result', request: divideRequest, value: 3 };
const error: CalculatorState = { status: 'error', request: divideRequest, message: 'Oops' };

const ALL_STATES: CalculatorState[] = [
  initialState,
  enteringFirst('12'),
  selected('12', 'divide'),
  enteringSecond('12', 'divide', '4'),
  loading,
  result,
  error,
];

describe('calculatorReducer', () => {
  it.each(ALL_STATES.map((state) => [state.status, state] as const))(
    'CLEAR resets %s to idle',
    (_status, state) => {
      expect(calculatorReducer(state, CLEAR)).toEqual(initialState);
    },
  );

  it.each(ALL_STATES.filter((s) => s.status !== 'loading').map((s) => [s.status, s] as const))(
    'ignores calculation outcomes while %s',
    (_status, state) => {
      expect(calculatorReducer(state, succeeded(1))).toBe(state);
      expect(calculatorReducer(state, failed('x'))).toBe(state);
    },
  );

  describe('idle', () => {
    it.each<[string, CalculatorAction, CalculatorState]>([
      ['DIGIT starts the first operand', digit('7'), enteringFirst('7')],
      ['DECIMAL starts with "0."', DECIMAL, enteringFirst('0.')],
      ['binary OPERATOR without operand is ignored', op('add'), initialState],
      ['unary OPERATOR without operand is ignored', op('sqrt'), initialState],
      ['EQUALS is ignored', EQUALS, initialState],
      ['BACKSPACE is ignored', BACKSPACE, initialState],
    ])('%s', (_name, action, expected) => {
      expect(calculatorReducer(initialState, action)).toEqual(expected);
    });
  });

  describe('enteringFirstOperand', () => {
    it.each<[string, CalculatorState, CalculatorAction, CalculatorState]>([
      ['DIGIT appends', enteringFirst('1'), digit('2'), enteringFirst('12')],
      ['DIGIT replaces a lone zero', enteringFirst('0'), digit('5'), enteringFirst('5')],
      ['DECIMAL appends a point', enteringFirst('3'), DECIMAL, enteringFirst('3.')],
      ['second DECIMAL is ignored', enteringFirst('3.1'), DECIMAL, enteringFirst('3.1')],
      ['BACKSPACE removes a character', enteringFirst('12'), BACKSPACE, enteringFirst('1')],
      ['BACKSPACE on last character returns to idle', enteringFirst('1'), BACKSPACE, initialState],
      [
        'binary OPERATOR selects the operation',
        enteringFirst('12'),
        op('divide'),
        selected('12', 'divide'),
      ],
      ['EQUALS is ignored', enteringFirst('12'), EQUALS, enteringFirst('12')],
    ])('%s', (_name, from, action, expected) => {
      expect(calculatorReducer(from, action)).toEqual(expected);
    });

    it('sqrt is sent to the backend immediately', () => {
      expect(calculatorReducer(enteringFirst('16'), op('sqrt'))).toEqual({
        status: 'loading',
        request: { operation: 'sqrt', operands: [16] },
      });
    });

    it('accepts an operand ending in a decimal point', () => {
      expect(calculatorReducer(enteringFirst('2.'), op('add'))).toEqual(selected('2.', 'add'));
    });

    it('ignores an OPERATOR if the operand cannot be parsed', () => {
      const corrupt = enteringFirst('abc');
      expect(calculatorReducer(corrupt, op('add'))).toBe(corrupt);
    });

    it('enforces the maximum input length', () => {
      const full = enteringFirst('1'.repeat(MAX_INPUT_DIGITS));
      expect(calculatorReducer(full, digit('2'))).toEqual(full);
    });
  });

  describe('operatorSelected', () => {
    const from = selected('12', 'divide');

    it.each<[string, CalculatorAction, CalculatorState]>([
      ['DIGIT starts the second operand', digit('4'), enteringSecond('12', 'divide', '4')],
      [
        'DECIMAL starts the second operand with "0."',
        DECIMAL,
        enteringSecond('12', 'divide', '0.'),
      ],
      [
        'another binary OPERATOR replaces the pending one',
        op('multiply'),
        selected('12', 'multiply'),
      ],
      ['unary OPERATOR is ignored', op('sqrt'), from],
      ['EQUALS without a second operand is ignored', EQUALS, from],
      ['BACKSPACE is ignored', BACKSPACE, from],
    ])('%s', (_name, action, expected) => {
      expect(calculatorReducer(from, action)).toEqual(expected);
    });
  });

  describe('enteringSecondOperand', () => {
    const from = enteringSecond('12', 'divide', '4');

    it.each<[string, CalculatorAction, CalculatorState]>([
      ['DIGIT appends', digit('0'), enteringSecond('12', 'divide', '40')],
      ['DECIMAL appends a point', DECIMAL, enteringSecond('12', 'divide', '4.')],
      ['BACKSPACE removes a character', BACKSPACE, selected('12', 'divide')],
      ['EQUALS submits the calculation', EQUALS, loading],
      ['unary OPERATOR is ignored', op('sqrt'), from],
    ])('%s', (_name, action, expected) => {
      expect(calculatorReducer(from, action)).toEqual(expected);
    });

    it('BACKSPACE keeps editing while characters remain', () => {
      expect(calculatorReducer(enteringSecond('1', 'add', '25'), BACKSPACE)).toEqual(
        enteringSecond('1', 'add', '2'),
      );
    });

    it('a binary OPERATOR submits and chains the next operation', () => {
      expect(calculatorReducer(from, op('add'))).toEqual({
        status: 'loading',
        request: divideRequest,
        nextOperation: 'add',
      });
    });

    it('does not submit if an operand cannot be parsed', () => {
      const corrupt = enteringSecond('abc', 'add', '1');
      expect(calculatorReducer(corrupt, EQUALS)).toBe(corrupt);
      expect(calculatorReducer(corrupt, op('add'))).toBe(corrupt);
    });
  });

  describe('loading', () => {
    it.each<[string, CalculatorAction]>([
      ['DIGIT', digit('1')],
      ['DECIMAL', DECIMAL],
      ['OPERATOR', op('add')],
      ['EQUALS (prevents double submit)', EQUALS],
      ['BACKSPACE', BACKSPACE],
    ])('ignores %s', (_name, action) => {
      expect(calculatorReducer(loading, action)).toBe(loading);
    });

    it('CALCULATION_SUCCEEDED shows the result', () => {
      expect(calculatorReducer(loading, succeeded(3))).toEqual(result);
    });

    it('CALCULATION_FAILED shows the error', () => {
      expect(calculatorReducer(loading, failed('Oops'))).toEqual(error);
    });

    it('CALCULATION_SUCCEEDED with a chained operation selects it on the result', () => {
      const chained: CalculatorState = { ...loading, nextOperation: 'multiply' };
      expect(calculatorReducer(chained, succeeded(3))).toEqual(selected('3', 'multiply'));
    });
  });

  describe('result', () => {
    it.each<[string, CalculatorAction, CalculatorState]>([
      ['DIGIT starts a new calculation', digit('9'), enteringFirst('9')],
      ['DECIMAL starts a new calculation', DECIMAL, enteringFirst('0.')],
      ['binary OPERATOR continues from the result', op('add'), selected('3', 'add')],
      ['EQUALS is ignored', EQUALS, result],
      ['BACKSPACE is ignored', BACKSPACE, result],
    ])('%s', (_name, action, expected) => {
      expect(calculatorReducer(result, action)).toEqual(expected);
    });

    it('sqrt is applied to the result immediately', () => {
      expect(calculatorReducer(result, op('sqrt'))).toEqual({
        status: 'loading',
        request: { operation: 'sqrt', operands: [3] },
      });
    });

    it('keeps full precision when continuing from a result', () => {
      const precise: CalculatorState = { ...result, value: 0.1 + 0.2 };
      expect(calculatorReducer(precise, op('multiply'))).toEqual(
        selected('0.30000000000000004', 'multiply'),
      );
    });
  });

  describe('error', () => {
    it.each<[string, CalculatorAction, CalculatorState]>([
      ['DIGIT starts a new calculation', digit('5'), enteringFirst('5')],
      ['DECIMAL starts a new calculation', DECIMAL, enteringFirst('0.')],
      ['OPERATOR resets to idle', op('add'), initialState],
      ['EQUALS resets to idle', EQUALS, initialState],
      ['BACKSPACE resets to idle', BACKSPACE, initialState],
    ])('%s', (_name, action, expected) => {
      expect(calculatorReducer(error, action)).toEqual(expected);
    });
  });

  describe('full flows', () => {
    it('12 ÷ 4 = 3', () => {
      const submitted = run([digit('1'), digit('2'), op('divide'), digit('4'), EQUALS]);
      expect(submitted).toEqual(loading);
      expect(calculatorReducer(submitted, succeeded(3))).toEqual(result);
    });

    it('2 + 3 × 4 chains through the intermediate result', () => {
      const first = run([digit('2'), op('add'), digit('3'), op('multiply')]);
      expect(first).toMatchObject({ status: 'loading', nextOperation: 'multiply' });

      const second = run([succeeded(5), digit('4'), EQUALS], first);
      expect(second).toEqual({
        status: 'loading',
        request: { operation: 'multiply', operands: [5, 4] },
      });
    });
  });
});
