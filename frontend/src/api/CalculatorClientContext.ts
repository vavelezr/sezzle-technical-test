import { createContext, useContext } from 'react';
import type { CalculatorClient } from './calculatorClient';

/**
 * Dependency injection point for the API client. The app provides the real
 * HTTP client; tests can provide any CalculatorClient implementation.
 */
export const CalculatorClientContext = createContext<CalculatorClient | null>(null);

export function useCalculatorClient(): CalculatorClient {
  const client = useContext(CalculatorClientContext);
  if (!client) {
    throw new Error('useCalculatorClient must be used within a CalculatorClientContext provider');
  }
  return client;
}
