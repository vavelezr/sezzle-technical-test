import { render, type RenderResult } from '@testing-library/react';
import type { ReactElement } from 'react';
import { CalculatorClientContext } from '../api/CalculatorClientContext';
import { createCalculatorClient, type CalculatorClient } from '../api/calculatorClient';
import { TEST_API_BASE_URL } from './apiMock';

/** Renders ui with an injected API client (the MSW-backed HTTP client by default). */
export function renderWithClient(
  ui: ReactElement,
  client: CalculatorClient = createCalculatorClient({ baseUrl: TEST_API_BASE_URL }),
): RenderResult {
  return render(<CalculatorClientContext value={client}>{ui}</CalculatorClientContext>);
}
