import { CalculatorError, friendlyErrorMessage, isApiErrorCode } from './errors';

describe('CalculatorError', () => {
  it('carries a code, message and optional status', () => {
    const error = new CalculatorError('DIVISION_BY_ZERO', 'division by zero is not allowed', 422);

    expect(error).toBeInstanceOf(Error);
    expect(error.name).toBe('CalculatorError');
    expect(error.code).toBe('DIVISION_BY_ZERO');
    expect(error.message).toBe('division by zero is not allowed');
    expect(error.status).toBe(422);
    expect(new CalculatorError('NETWORK_ERROR', 'offline').status).toBeUndefined();
  });
});

describe('isApiErrorCode', () => {
  it.each(['DIVISION_BY_ZERO', 'INTERNAL_ERROR'])('accepts %s', (code) => {
    expect(isApiErrorCode(code)).toBe(true);
  });

  it.each(['NETWORK_ERROR', 'SOMETHING_NEW', 42, undefined])('rejects %j', (code) => {
    expect(isApiErrorCode(code)).toBe(false);
  });
});

describe('friendlyErrorMessage', () => {
  it.each([
    ['DIVISION_BY_ZERO', "You can't divide by zero"],
    ['NEGATIVE_SQUARE_ROOT', "Negative numbers don't have a real square root"],
    ['RESULT_OVERFLOW', 'That result is too large to display'],
    ['UNDEFINED_RESULT', "That result isn't a real number"],
    ['NETWORK_ERROR', "Can't reach the calculator service. Please try again."],
    ['INTERNAL_ERROR', 'Something went wrong. Please try again.'],
    ['UNKNOWN_ERROR', 'Something went wrong. Please try again.'],
  ] as const)('maps %s to a friendly message', (code, expected) => {
    expect(friendlyErrorMessage(code)).toBe(expected);
  });
});
