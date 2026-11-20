import { http, HttpResponse, type PathParams } from 'msw';
import { setupServer } from 'msw/node';
import type { CalculateRequestDto, CalculateResponseDto, ErrorResponseDto } from '../api/dto';

/** Absolute base URL used by tests; MSW intercepts every request to it. */
export const TEST_API_BASE_URL = 'http://calculator.test/api/v1';
export const CALCULATE_URL = `${TEST_API_BASE_URL}/calculate`;

/**
 * Test double of the backend's /calculate endpoint. It mirrors the real
 * contract for the cases exercised by tests; it is never used in the app.
 */
const fakeOperations: Record<string, (a: number, b: number) => number> = {
  add: (a, b) => a + b,
  subtract: (a, b) => a - b,
  multiply: (a, b) => a * b,
  divide: (a, b) => a / b,
  power: (a, b) => a ** b,
  percentage: (a, b) => (a / 100) * b,
  sqrt: (a) => Math.sqrt(a),
};

export function errorResponse(status: number, code: string, message: string) {
  return HttpResponse.json<ErrorResponseDto>({ error: { code, message } }, { status });
}

export const calculateHandler = http.post<
  PathParams,
  CalculateRequestDto,
  CalculateResponseDto | ErrorResponseDto
>(CALCULATE_URL, async ({ request }) => {
  const { operation, operands } = await request.json();
  const [a = 0, b = 0] = operands;
  const apply = fakeOperations[operation];

  if (!apply) {
    return errorResponse(400, 'UNSUPPORTED_OPERATION', `unsupported operation: "${operation}"`);
  }
  if (operation === 'divide' && b === 0) {
    return errorResponse(422, 'DIVISION_BY_ZERO', 'division by zero is not allowed');
  }
  if (operation === 'sqrt' && a < 0) {
    return errorResponse(
      422,
      'NEGATIVE_SQUARE_ROOT',
      'square root of a negative number is not allowed',
    );
  }
  return HttpResponse.json<CalculateResponseDto>({ operation, operands, result: apply(a, b) });
});

export const server = setupServer(calculateHandler);
