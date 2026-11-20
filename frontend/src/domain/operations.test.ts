import {
  OPERATIONS,
  expressionSymbol,
  getOperationMeta,
  isUnaryOperation,
  operationForKey,
  type Operation,
} from './operations';

describe('operations metadata', () => {
  it('defines every operation exactly once', () => {
    const ids = OPERATIONS.map((op) => op.id);
    expect(ids).toEqual(['add', 'subtract', 'multiply', 'divide', 'power', 'sqrt', 'percentage']);
  });

  it('gives every operation a symbol, an accessible label and a unique key binding', () => {
    const keys = OPERATIONS.flatMap((op) => op.keys);
    expect(new Set(keys).size).toBe(keys.length);
    for (const op of OPERATIONS) {
      expect(op.symbol).not.toBe('');
      expect(op.label).not.toBe('');
      expect(op.keys.length).toBeGreaterThan(0);
    }
  });

  it.each<[Operation, boolean]>([
    ['sqrt', true],
    ['add', false],
    ['percentage', false],
  ])('isUnaryOperation(%s) is %s', (id, expected) => {
    expect(isUnaryOperation(id)).toBe(expected);
  });

  it('returns metadata by id', () => {
    expect(getOperationMeta('divide')).toMatchObject({ symbol: '÷', label: 'Divide', arity: 2 });
  });

  it('throws for an unknown id', () => {
    expect(() => getOperationMeta('modulo' as Operation)).toThrow('Unknown operation: modulo');
  });

  it.each<[string, Operation | undefined]>([
    ['+', 'add'],
    ['-', 'subtract'],
    ['*', 'multiply'],
    ['x', 'multiply'],
    ['/', 'divide'],
    ['^', 'power'],
    ['r', 'sqrt'],
    ['%', 'percentage'],
    ['a', undefined],
  ])('maps key %j to %s', (key, expected) => {
    expect(operationForKey(key)).toBe(expected);
  });

  it('provides expression symbols', () => {
    expect(expressionSymbol('power')).toBe('^');
    expect(expressionSymbol('percentage')).toBe('% of');
  });
});
