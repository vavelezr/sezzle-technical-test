import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { OPERATIONS } from '../../../domain/operations';
import { Keypad } from './Keypad';

describe('Keypad', () => {
  it('renders an accessible button for every operation in the config', () => {
    render(<Keypad onCommand={vi.fn()} isLoading={false} />);

    for (const op of OPERATIONS) {
      expect(screen.getByRole('button', { name: op.label })).toHaveTextContent(op.symbol);
    }
  });

  it('renders digits, decimal point, clear, backspace and equals', () => {
    render(<Keypad onCommand={vi.fn()} isLoading={false} />);

    for (const name of ['0', '1', '9', 'Decimal point', 'Clear', 'Backspace', 'Equals']) {
      expect(screen.getByRole('button', { name })).toBeInTheDocument();
    }
  });

  it.each([
    ['7', { type: 'DIGIT', value: '7' }],
    ['Decimal point', { type: 'DECIMAL' }],
    ['Divide', { type: 'OPERATOR', operation: 'divide' }],
    ['Square root', { type: 'OPERATOR', operation: 'sqrt' }],
    ['Equals', { type: 'EQUALS' }],
    ['Backspace', { type: 'BACKSPACE' }],
    ['Clear', { type: 'CLEAR' }],
  ])('pressing %s dispatches its command', async (name, command) => {
    const onCommand = vi.fn();
    render(<Keypad onCommand={onCommand} isLoading={false} />);

    await userEvent.click(screen.getByRole('button', { name }));

    expect(onCommand).toHaveBeenCalledWith(command);
  });

  it('marks the pending operation as pressed', () => {
    render(<Keypad onCommand={vi.fn()} activeOperation="multiply" isLoading={false} />);

    expect(screen.getByRole('button', { name: 'Multiply' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(screen.getByRole('button', { name: 'Add' })).toHaveAttribute('aria-pressed', 'false');
  });

  it('disables equals while a calculation is loading', () => {
    render(<Keypad onCommand={vi.fn()} isLoading />);

    expect(screen.getByRole('button', { name: 'Equals' })).toBeDisabled();
    expect(screen.getByRole('button', { name: '7' })).toBeEnabled();
  });
});
