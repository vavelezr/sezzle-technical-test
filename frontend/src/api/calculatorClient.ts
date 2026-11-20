import type { CalculationRequest } from '../domain/calculation';
import { CalculatorError, isApiErrorCode } from '../domain/errors';
import { isValidOperand } from '../domain/validation';
import { isCalculateResponseDto, isErrorResponseDto, type CalculateRequestDto } from './dto';

/** Port used by the UI to run calculations. */
export interface CalculatorClient {
  /**
   * Runs a calculation on the backend and resolves with its result.
   * Rejects with a CalculatorError, or with the abort reason if `signal` aborts.
   */
  calculate(request: CalculationRequest, signal?: AbortSignal): Promise<number>;
}

export interface CalculatorClientOptions {
  /** API base URL, e.g. "/api/v1". */
  baseUrl: string;
  /** Fetch implementation; injectable for testing. Defaults to the global fetch. */
  fetch?: typeof fetch;
}

/**
 * Factory for the HTTP calculator client. The client is also an Adapter: it
 * converts domain requests to DTOs and backend responses/errors back into
 * domain values and typed CalculatorErrors.
 */
export function createCalculatorClient({
  baseUrl,
  fetch: fetchImpl = (...args) => globalThis.fetch(...args),
}: CalculatorClientOptions): CalculatorClient {
  const calculateUrl = `${baseUrl.replace(/\/+$/, '')}/calculate`;

  return {
    async calculate(request, signal) {
      if (!request.operands.every(isValidOperand)) {
        throw new CalculatorError('INVALID_OPERAND', 'operands must be finite numbers');
      }

      const body: CalculateRequestDto = {
        operation: request.operation,
        operands: [...request.operands],
      };

      let response: Response;
      try {
        response = await fetchImpl(calculateUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
          body: JSON.stringify(body),
          signal,
        });
      } catch (error) {
        if (signal?.aborted) {
          throw error;
        }
        throw new CalculatorError('NETWORK_ERROR', 'the calculator service is unreachable');
      }

      const payload = await readJson(response);
      if (response.ok) {
        if (!isCalculateResponseDto(payload)) {
          throw new CalculatorError(
            'INVALID_RESPONSE',
            'unexpected response format',
            response.status,
          );
        }
        return payload.result;
      }
      throw toCalculatorError(payload, response.status);
    },
  };
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return (await response.json()) as unknown;
  } catch {
    return undefined;
  }
}

function toCalculatorError(payload: unknown, status: number): CalculatorError {
  if (isErrorResponseDto(payload)) {
    const code = isApiErrorCode(payload.error.code) ? payload.error.code : 'UNKNOWN_ERROR';
    return new CalculatorError(code, payload.error.message, status);
  }
  return new CalculatorError(
    'UNKNOWN_ERROR',
    `request failed with status ${String(status)}`,
    status,
  );
}
