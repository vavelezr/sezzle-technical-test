/** Error codes returned by the backend API. */
export const API_ERROR_CODES = [
  'INVALID_REQUEST',
  'PAYLOAD_TOO_LARGE',
  'UNSUPPORTED_OPERATION',
  'INVALID_OPERAND_COUNT',
  'INVALID_OPERAND',
  'DIVISION_BY_ZERO',
  'NEGATIVE_SQUARE_ROOT',
  'RESULT_OVERFLOW',
  'UNDEFINED_RESULT',
  'NOT_FOUND',
  'METHOD_NOT_ALLOWED',
  'REQUEST_TIMEOUT',
  'INTERNAL_ERROR',
] as const;

export type ApiErrorCode = (typeof API_ERROR_CODES)[number];

/** Error codes produced on the client side. */
export type ClientErrorCode = 'NETWORK_ERROR' | 'INVALID_RESPONSE' | 'UNKNOWN_ERROR';

export type CalculatorErrorCode = ApiErrorCode | ClientErrorCode;

/** Typed error raised by the API client for every failed calculation. */
export class CalculatorError extends Error {
  readonly code: CalculatorErrorCode;
  readonly status: number | undefined;

  constructor(code: CalculatorErrorCode, message: string, status?: number) {
    super(message);
    this.name = 'CalculatorError';
    this.code = code;
    this.status = status;
  }
}

export function isApiErrorCode(value: unknown): value is ApiErrorCode {
  return typeof value === 'string' && (API_ERROR_CODES as readonly string[]).includes(value);
}

const FRIENDLY_MESSAGES: Readonly<Partial<Record<CalculatorErrorCode, string>>> = {
  DIVISION_BY_ZERO: "You can't divide by zero",
  NEGATIVE_SQUARE_ROOT: "Negative numbers don't have a real square root",
  RESULT_OVERFLOW: 'That result is too large to display',
  UNDEFINED_RESULT: "That result isn't a real number",
  INVALID_OPERAND: 'Please enter a valid number',
  INVALID_OPERAND_COUNT: 'Please enter all the numbers this operation needs',
  REQUEST_TIMEOUT: 'The calculation took too long. Please try again.',
  NETWORK_ERROR: "Can't reach the calculator service. Please try again.",
};

const FALLBACK_MESSAGE = 'Something went wrong. Please try again.';

/** Maps an error code to a message suitable for end users. */
export function friendlyErrorMessage(code: CalculatorErrorCode): string {
  return FRIENDLY_MESSAGES[code] ?? FALLBACK_MESSAGE;
}
