import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { CalculatorClientContext } from './api/CalculatorClientContext';
import { createCalculatorClient } from './api/calculatorClient';
import './theme/global.css';

const client = createCalculatorClient({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
});

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Root element #root not found');
}

createRoot(rootElement).render(
  <StrictMode>
    <CalculatorClientContext value={client}>
      <App />
    </CalculatorClientContext>
  </StrictMode>,
);
