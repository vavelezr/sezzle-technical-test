import type { CalculationRequest } from '../../domain/calculation';
import { formatNumber } from '../../domain/format';
import { expressionSymbol, type Operation } from '../../domain/operations';
import type { CalculatorState } from './calculatorReducer';

/** What the display shows: a secondary expression line and a main value. */
export interface DisplayModel {
  readonly expression: string;
  readonly value: string;
  readonly isError: boolean;
}

/** Formats operand input; complete numbers are formatted, partial input is kept as typed. */
function formatOperand(operand: string): string {
  const value = Number(operand);
  return operand.endsWith('.') || !Number.isFinite(value) ? operand : formatNumber(value);
}

function formatRequest({ operation, operands }: CalculationRequest): string {
  const formatted = operands.map(formatNumber);
  const symbol = expressionSymbol(operation);
  return formatted.length === 1
    ? `${symbol}(${formatted.join('')})`
    : formatted.join(` ${symbol} `);
}

function pending(operand: string, operation: Operation): string {
  return `${formatOperand(operand)} ${expressionSymbol(operation)}`;
}

export function selectDisplay(state: CalculatorState): DisplayModel {
  switch (state.status) {
    case 'idle':
      return { expression: '', value: '0', isError: false };
    case 'enteringFirstOperand':
      return { expression: '', value: state.input, isError: false };
    case 'operatorSelected':
      return {
        expression: pending(state.firstOperand, state.operation),
        value: formatOperand(state.firstOperand),
        isError: false,
      };
    case 'enteringSecondOperand':
      return {
        expression: pending(state.firstOperand, state.operation),
        value: state.input,
        isError: false,
      };
    case 'loading':
      return { expression: `${formatRequest(state.request)} =`, value: '…', isError: false };
    case 'result':
      return {
        expression: `${formatRequest(state.request)} =`,
        value: formatNumber(state.value),
        isError: false,
      };
    case 'error':
      return {
        expression: `${formatRequest(state.request)} =`,
        value: state.message,
        isError: true,
      };
  }
}
