/** Wire format of the backend REST API (see the README API reference). */

export interface CalculateRequestDto {
  operation: string;
  operands: number[];
}

export interface CalculateResponseDto {
  operation: string;
  operands: number[];
  result: number;
}

export interface ErrorResponseDto {
  error: {
    code: string;
    message: string;
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

export function isCalculateResponseDto(value: unknown): value is CalculateResponseDto {
  return isRecord(value) && typeof value.result === 'number' && Number.isFinite(value.result);
}

export function isErrorResponseDto(value: unknown): value is ErrorResponseDto {
  return (
    isRecord(value) &&
    isRecord(value.error) &&
    typeof value.error.code === 'string' &&
    typeof value.error.message === 'string'
  );
}
