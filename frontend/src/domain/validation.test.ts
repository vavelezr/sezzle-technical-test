import {
  MAX_INPUT_DIGITS,
  appendDecimalPoint,
  appendDigit,
  isDigit,
  isValidOperand,
  parseOperand,
  removeLastCharacter,
} from './validation';

describe('validation', () => {
  describe('isDigit', () => {
    it.each(['0', '5', '9'])('accepts %s', (value) => {
      expect(isDigit(value)).toBe(true);
    });
    it.each(['a', '.', '10', ''])('rejects %j', (value) => {
      expect(isDigit(value)).toBe(false);
    });
  });

  describe('appendDigit', () => {
    it.each([
      ['', '7', '7'],
      ['12', '3', '123'],
      ['0', '5', '5'],
      ['0', '0', '0'],
      ['0.', '0', '0.0'],
      ['1.5', '2', '1.52'],
    ] as const)('appendDigit(%j, %j) = %j', (input, digit, expected) => {
      expect(appendDigit(input, digit)).toBe(expected);
    });

    it(`stops at ${String(MAX_INPUT_DIGITS)} digits`, () => {
      const full = '9'.repeat(MAX_INPUT_DIGITS);
      expect(appendDigit(full, '1')).toBe(full);
    });

    it('does not count the decimal point towards the limit', () => {
      const almostFull = `${'1'.repeat(MAX_INPUT_DIGITS - 1)}.`;
      expect(appendDigit(almostFull, '5')).toBe(`${almostFull}5`);
    });
  });

  describe('appendDecimalPoint', () => {
    it.each([
      ['', '0.'],
      ['3', '3.'],
      ['3.', '3.'],
      ['3.14', '3.14'],
    ])('appendDecimalPoint(%j) = %j', (input, expected) => {
      expect(appendDecimalPoint(input)).toBe(expected);
    });
  });

  describe('removeLastCharacter', () => {
    it.each([
      ['123', '12'],
      ['1.', '1'],
      ['5', ''],
      ['', ''],
    ])('removeLastCharacter(%j) = %j', (input, expected) => {
      expect(removeLastCharacter(input)).toBe(expected);
    });
  });

  describe('parseOperand', () => {
    it.each([
      ['42', 42],
      ['0.', 0],
      ['-3.5', -3.5],
      ['1e+21', 1e21],
    ])('parses %j as %d', (input, expected) => {
      expect(parseOperand(input)).toBe(expected);
    });

    it.each(['', '  ', 'abc', 'NaN', 'Infinity', '1e999'])('rejects %j', (input) => {
      expect(parseOperand(input)).toBeNull();
    });
  });

  describe('isValidOperand', () => {
    it('accepts finite numbers and rejects NaN and infinities', () => {
      expect(isValidOperand(1.5)).toBe(true);
      expect(isValidOperand(Number.NaN)).toBe(false);
      expect(isValidOperand(Number.POSITIVE_INFINITY)).toBe(false);
    });
  });
});
