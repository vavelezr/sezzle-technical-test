/**
 * Pure helpers that keep operand input valid while the user types.
 * Operand input is held as a string so it can represent partial numbers
 * such as "12." exactly as typed.
 */

export type Digit = '0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9';

/** Maximum number of digits accepted in a single operand. */
export const MAX_INPUT_DIGITS = 15;

const DIGITS: ReadonlySet<string> = new Set(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9']);
const DECIMAL_POINT = '.';

export function isDigit(value: string): value is Digit {
  return DIGITS.has(value);
}

function countDigits(input: string): number {
  return input.replace(/\D/g, '').length;
}

/**
 * Appends a digit, replacing a lone leading zero ("0" + "5" -> "5").
 * Input that already has the maximum number of digits is returned unchanged.
 */
export function appendDigit(input: string, digit: Digit): string {
  if (input === '0') {
    return digit;
  }
  if (countDigits(input) >= MAX_INPUT_DIGITS) {
    return input;
  }
  return input + digit;
}

/** Appends a decimal point, allowing at most one ("" -> "0."). */
export function appendDecimalPoint(input: string): string {
  if (input.includes(DECIMAL_POINT)) {
    return input;
  }
  return (input === '' ? '0' : input) + DECIMAL_POINT;
}

/** Removes the last typed character. */
export function removeLastCharacter(input: string): string {
  return input.slice(0, -1);
}

/**
 * Converts operand input into a number, or returns null if it is not a
 * finite number. This guarantees NaN/Infinity are never sent to the API.
 */
export function parseOperand(input: string): number | null {
  if (input.trim() === '') {
    return null;
  }
  const value = Number(input);
  return Number.isFinite(value) ? value : null;
}

export function isValidOperand(value: number): boolean {
  return Number.isFinite(value);
}
