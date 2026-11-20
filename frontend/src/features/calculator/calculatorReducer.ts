import type { CalculationRequest } from '../../domain/calculation';
import { isUnaryOperation, type BinaryOperation, type Operation } from '../../domain/operations';
import {
  appendDecimalPoint,
  appendDigit,
  parseOperand,
  removeLastCharacter,
  type Digit,
} from '../../domain/validation';
import type { CalculatorAction } from './commands';

/**
 * Calculator state machine (State pattern). Each status only carries the data
 * that is meaningful in it, and the reducer defines which actions are valid
 * in each status; anything else is ignored.
 *
 * Operands are kept as strings so partial input like "12." is shown as typed.
 */
export type CalculatorState =
  | { readonly status: 'idle' }
  | { readonly status: 'enteringFirstOperand'; readonly input: string }
  | {
      readonly status: 'operatorSelected';
      readonly firstOperand: string;
      readonly operation: BinaryOperation;
    }
  | {
      readonly status: 'enteringSecondOperand';
      readonly firstOperand: string;
      readonly operation: BinaryOperation;
      readonly input: string;
    }
  | {
      readonly status: 'loading';
      readonly request: CalculationRequest;
      /** Operator pressed to chain calculations ("2 + 3 ×"), applied to the result. */
      readonly nextOperation?: BinaryOperation;
    }
  | { readonly status: 'result'; readonly request: CalculationRequest; readonly value: number }
  | { readonly status: 'error'; readonly request: CalculationRequest; readonly message: string };

export type CalculatorStatus = CalculatorState['status'];

export const initialState: CalculatorState = { status: 'idle' };

export function calculatorReducer(
  state: CalculatorState,
  action: CalculatorAction,
): CalculatorState {
  if (action.type === 'CLEAR') {
    return initialState;
  }

  switch (state.status) {
    case 'idle':
      return startFirstOperand(action) ?? state;
    case 'enteringFirstOperand':
      return fromEnteringFirstOperand(state, action);
    case 'operatorSelected':
      return fromOperatorSelected(state, action);
    case 'enteringSecondOperand':
      return fromEnteringSecondOperand(state, action);
    case 'loading':
      return fromLoading(state, action);
    case 'result':
      return fromResult(state, action);
    case 'error':
      return fromError(state, action);
  }
}

type StateOf<S extends CalculatorStatus> = Extract<CalculatorState, { status: S }>;

function fromEnteringFirstOperand(
  state: StateOf<'enteringFirstOperand'>,
  action: CalculatorAction,
): CalculatorState {
  switch (action.type) {
    case 'DIGIT':
      return { ...state, input: appendDigit(state.input, action.value) };
    case 'DECIMAL':
      return { ...state, input: appendDecimalPoint(state.input) };
    case 'BACKSPACE': {
      const input = removeLastCharacter(state.input);
      return input === '' ? initialState : { ...state, input };
    }
    case 'OPERATOR':
      return applyOperator(state.input, action.operation) ?? state;
    default:
      return state;
  }
}

function fromOperatorSelected(
  state: StateOf<'operatorSelected'>,
  action: CalculatorAction,
): CalculatorState {
  switch (action.type) {
    case 'DIGIT':
      return { ...state, status: 'enteringSecondOperand', input: newInput(action.value) };
    case 'DECIMAL':
      return { ...state, status: 'enteringSecondOperand', input: appendDecimalPoint('') };
    case 'OPERATOR':
      // Pressing another binary operator replaces the pending one.
      return isUnaryOperation(action.operation) ? state : { ...state, operation: action.operation };
    default:
      return state;
  }
}

function fromEnteringSecondOperand(
  state: StateOf<'enteringSecondOperand'>,
  action: CalculatorAction,
): CalculatorState {
  switch (action.type) {
    case 'DIGIT':
      return { ...state, input: appendDigit(state.input, action.value) };
    case 'DECIMAL':
      return { ...state, input: appendDecimalPoint(state.input) };
    case 'BACKSPACE': {
      const input = removeLastCharacter(state.input);
      if (input === '') {
        return {
          status: 'operatorSelected',
          firstOperand: state.firstOperand,
          operation: state.operation,
        };
      }
      return { ...state, input };
    }
    case 'EQUALS':
      return submitBinary(state) ?? state;
    case 'OPERATOR': {
      if (isUnaryOperation(action.operation)) {
        return state;
      }
      const loading = submitBinary(state);
      return loading ? { ...loading, nextOperation: action.operation } : state;
    }
    default:
      return state;
  }
}

/** While a request is in flight only its outcome (or CLEAR) is accepted. */
function fromLoading(state: StateOf<'loading'>, action: CalculatorAction): CalculatorState {
  switch (action.type) {
    case 'CALCULATION_SUCCEEDED':
      if (state.nextOperation) {
        return {
          status: 'operatorSelected',
          firstOperand: String(action.result),
          operation: state.nextOperation,
        };
      }
      return { status: 'result', request: state.request, value: action.result };
    case 'CALCULATION_FAILED':
      return { status: 'error', request: state.request, message: action.message };
    default:
      return state;
  }
}

function fromResult(state: StateOf<'result'>, action: CalculatorAction): CalculatorState {
  if (action.type === 'OPERATOR') {
    // The previous result becomes the first operand of the next calculation.
    return applyOperator(String(state.value), action.operation) ?? state;
  }
  return startFirstOperand(action) ?? state;
}

/** Any key dismisses the error; number input also starts a new calculation. */
function fromError(state: StateOf<'error'>, action: CalculatorAction): CalculatorState {
  if (action.type === 'CALCULATION_SUCCEEDED' || action.type === 'CALCULATION_FAILED') {
    return state;
  }
  return startFirstOperand(action) ?? initialState;
}

/** Starts a fresh first operand on DIGIT or DECIMAL; otherwise returns undefined. */
function startFirstOperand(action: CalculatorAction): CalculatorState | undefined {
  switch (action.type) {
    case 'DIGIT':
      return { status: 'enteringFirstOperand', input: newInput(action.value) };
    case 'DECIMAL':
      return { status: 'enteringFirstOperand', input: appendDecimalPoint('') };
    default:
      return undefined;
  }
}

function newInput(digit: Digit): string {
  return appendDigit('', digit);
}

/**
 * Applies an operator to a complete first operand: unary operations are sent
 * immediately, binary ones wait for the second operand.
 */
function applyOperator(operand: string, operation: Operation): CalculatorState | undefined {
  const value = parseOperand(operand);
  if (value === null) {
    return undefined;
  }
  if (isUnaryOperation(operation)) {
    return { status: 'loading', request: { operation, operands: [value] } };
  }
  return { status: 'operatorSelected', firstOperand: operand, operation };
}

function submitBinary(state: StateOf<'enteringSecondOperand'>): StateOf<'loading'> | undefined {
  const first = parseOperand(state.firstOperand);
  const second = parseOperand(state.input);
  if (first === null || second === null) {
    return undefined;
  }
  return { status: 'loading', request: { operation: state.operation, operands: [first, second] } };
}
