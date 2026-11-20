import { http, HttpResponse } from 'msw';
import { API_ERROR_CODES, CalculatorError } from '../domain/errors';
import { CALCULATE_URL, TEST_API_BASE_URL, errorResponse, server } from '../test/apiMock';
import { createCalculatorClient } from './calculatorClient';
import type { CalculateRequestDto } from './dto';

const client = createCalculatorClient({ baseUrl: TEST_API_BASE_URL });

async function captureError(promise: Promise<unknown>): Promise<CalculatorError> {
  const error: unknown = await promise.then(
    () => undefined,
    (reason: unknown) => reason,
  );
  if (!(error instanceof CalculatorError)) {
    throw new Error(`expected a CalculatorError, got ${String(error)}`);
  }
  return error;
}

describe('createCalculatorClient', () => {
  it('resolves with the result of a successful calculation', async () => {
    await expect(client.calculate({ operation: 'divide', operands: [10, 4] })).resolves.toBe(2.5);
  });

  it('sends the request as a JSON DTO', async () => {
    let received: CalculateRequestDto | undefined;
    server.use(
      http.post(CALCULATE_URL, async ({ request }) => {
        expect(request.headers.get('Content-Type')).toBe('application/json');
        received = (await request.json()) as CalculateRequestDto;
        return HttpResponse.json({ operation: 'sqrt', operands: [9], result: 3 });
      }),
    );

    await client.calculate({ operation: 'sqrt', operands: [9] });

    expect(received).toEqual({ operation: 'sqrt', operands: [9] });
  });

  it('tolerates a trailing slash in the base URL', async () => {
    const slashed = createCalculatorClient({ baseUrl: `${TEST_API_BASE_URL}/` });
    await expect(slashed.calculate({ operation: 'add', operands: [1, 2] })).resolves.toBe(3);
  });

  it.each(API_ERROR_CODES)('maps the %s error code to a CalculatorError', async (code) => {
    server.use(http.post(CALCULATE_URL, () => errorResponse(400, code, `backend says ${code}`)));

    const error = await captureError(client.calculate({ operation: 'add', operands: [1, 2] }));

    expect(error.code).toBe(code);
    expect(error.message).toBe(`backend says ${code}`);
    expect(error.status).toBe(400);
  });

  it('maps an unrecognised error code to UNKNOWN_ERROR', async () => {
    server.use(http.post(CALCULATE_URL, () => errorResponse(418, 'TEAPOT', 'short and stout')));

    const error = await captureError(client.calculate({ operation: 'add', operands: [1, 2] }));

    expect(error).toMatchObject({ code: 'UNKNOWN_ERROR', message: 'short and stout', status: 418 });
  });

  it('maps a non-JSON error response to UNKNOWN_ERROR', async () => {
    server.use(http.post(CALCULATE_URL, () => new HttpResponse('Bad Gateway', { status: 502 })));

    const error = await captureError(client.calculate({ operation: 'add', operands: [1, 2] }));

    expect(error).toMatchObject({ code: 'UNKNOWN_ERROR', status: 502 });
  });

  it('rejects a successful response with an unexpected shape', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.json({ answer: 42 })));

    const error = await captureError(client.calculate({ operation: 'add', operands: [1, 2] }));

    expect(error).toMatchObject({ code: 'INVALID_RESPONSE', status: 200 });
  });

  it('maps a network failure to NETWORK_ERROR', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.error()));

    const error = await captureError(client.calculate({ operation: 'add', operands: [1, 2] }));

    expect(error.code).toBe('NETWORK_ERROR');
  });

  it.each([Number.NaN, Number.POSITIVE_INFINITY])(
    'never sends a non-finite operand (%d)',
    async (operand) => {
      const fetchSpy = vi.fn<typeof fetch>();
      const guarded = createCalculatorClient({ baseUrl: TEST_API_BASE_URL, fetch: fetchSpy });

      const error = await captureError(
        guarded.calculate({ operation: 'add', operands: [operand, 1] }),
      );

      expect(error.code).toBe('INVALID_OPERAND');
      expect(fetchSpy).not.toHaveBeenCalled();
    },
  );

  it('propagates the abort reason instead of reporting a network error', async () => {
    const controller = new AbortController();
    const abortError = new DOMException('aborted', 'AbortError');
    const fetchStub = vi.fn<typeof fetch>(() => {
      controller.abort();
      return Promise.reject(abortError);
    });
    const abortable = createCalculatorClient({ baseUrl: TEST_API_BASE_URL, fetch: fetchStub });

    await expect(
      abortable.calculate({ operation: 'add', operands: [1, 2] }, controller.signal),
    ).rejects.toBe(abortError);
  });

  it('uses the injected fetch implementation', async () => {
    const fetchStub = vi.fn<typeof fetch>(() =>
      Promise.resolve(Response.json({ operation: 'add', operands: [1, 2], result: 3 })),
    );
    const injected = createCalculatorClient({ baseUrl: '/api/v1', fetch: fetchStub });

    await expect(injected.calculate({ operation: 'add', operands: [1, 2] })).resolves.toBe(3);
    expect(fetchStub).toHaveBeenCalledWith(
      '/api/v1/calculate',
      expect.objectContaining({ method: 'POST' }),
    );
  });
});
