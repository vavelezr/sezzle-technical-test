import type { CalculatorCommand } from './commands';
import { commandForKey } from './keyboard';

describe('commandForKey', () => {
  it.each<[string, CalculatorCommand | undefined]>([
    ['0', { type: 'DIGIT', value: '0' }],
    ['7', { type: 'DIGIT', value: '7' }],
    ['.', { type: 'DECIMAL' }],
    [',', { type: 'DECIMAL' }],
    ['Enter', { type: 'EQUALS' }],
    ['=', { type: 'EQUALS' }],
    ['Backspace', { type: 'BACKSPACE' }],
    ['Escape', { type: 'CLEAR' }],
    ['Delete', { type: 'CLEAR' }],
    ['+', { type: 'OPERATOR', operation: 'add' }],
    ['-', { type: 'OPERATOR', operation: 'subtract' }],
    ['*', { type: 'OPERATOR', operation: 'multiply' }],
    ['/', { type: 'OPERATOR', operation: 'divide' }],
    ['^', { type: 'OPERATOR', operation: 'power' }],
    ['r', { type: 'OPERATOR', operation: 'sqrt' }],
    ['%', { type: 'OPERATOR', operation: 'percentage' }],
    ['a', undefined],
    ['Tab', undefined],
    [' ', undefined],
  ])('maps %j', (key, expected) => {
    expect(commandForKey(key)).toEqual(expected);
  });
});
