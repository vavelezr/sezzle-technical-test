/**
 * Single source of truth for the operations the calculator supports.
 * UI buttons and keyboard shortcuts are generated from this config
 * (Factory), so adding an operation means adding one entry here.
 */

export type BinaryOperation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'percentage';
export type UnaryOperation = 'sqrt';
export type Operation = BinaryOperation | UnaryOperation;

interface OperationMetaBase {
  /** Visible button text. */
  readonly symbol: string;
  /** Accessible name for screen readers. */
  readonly label: string;
  /** Keyboard keys (KeyboardEvent.key values) that trigger the operation. */
  readonly keys: readonly string[];
}

export interface BinaryOperationMeta extends OperationMetaBase {
  readonly id: BinaryOperation;
  readonly arity: 2;
}

export interface UnaryOperationMeta extends OperationMetaBase {
  readonly id: UnaryOperation;
  readonly arity: 1;
}

export type OperationMeta = BinaryOperationMeta | UnaryOperationMeta;

export const OPERATIONS: readonly OperationMeta[] = [
  { id: 'add', symbol: '+', label: 'Add', keys: ['+'], arity: 2 },
  { id: 'subtract', symbol: '−', label: 'Subtract', keys: ['-'], arity: 2 },
  { id: 'multiply', symbol: '×', label: 'Multiply', keys: ['*', 'x'], arity: 2 },
  { id: 'divide', symbol: '÷', label: 'Divide', keys: ['/'], arity: 2 },
  { id: 'power', symbol: 'xʸ', label: 'Power', keys: ['^'], arity: 2 },
  { id: 'sqrt', symbol: '√', label: 'Square root', keys: ['r'], arity: 1 },
  { id: 'percentage', symbol: '%', label: 'Percent of', keys: ['%'], arity: 2 },
];

/** Symbols used when rendering an expression such as "2 ^ 8". */
const EXPRESSION_SYMBOLS: Readonly<Record<Operation, string>> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
  sqrt: '√',
  percentage: '% of',
};

export function getOperationMeta(id: Operation): OperationMeta {
  const meta = OPERATIONS.find((op) => op.id === id);
  if (!meta) {
    throw new Error(`Unknown operation: ${id}`);
  }
  return meta;
}

export function isUnaryOperation(id: Operation): id is UnaryOperation {
  return getOperationMeta(id).arity === 1;
}

export function expressionSymbol(id: Operation): string {
  return EXPRESSION_SYMBOLS[id];
}

/** Finds the operation bound to a keyboard key, if any. */
export function operationForKey(key: string): Operation | undefined {
  return OPERATIONS.find((op) => op.keys.includes(key))?.id;
}
