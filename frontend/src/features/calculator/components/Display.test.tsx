import { render, screen } from '@testing-library/react';
import { Display } from './Display';

describe('Display', () => {
  it('shows the expression and announces the value politely', () => {
    render(
      <Display display={{ expression: '12 ÷', value: '4', isError: false }} isLoading={false} />,
    );

    expect(screen.getByTestId('expression')).toHaveTextContent('12 ÷');
    const value = screen.getByRole('status');
    expect(value).toHaveTextContent('4');
    expect(value).toHaveAttribute('aria-live', 'polite');
    expect(value).not.toHaveClass('error');
  });

  it('marks errors visually', () => {
    render(
      <Display
        display={{ expression: '1 ÷ 0 =', value: "You can't divide by zero", isError: true }}
        isLoading={false}
      />,
    );

    expect(screen.getByRole('status')).toHaveClass('error', 'message');
  });

  it('shows a loading indicator while busy', () => {
    render(<Display display={{ expression: '1 + 2 =', value: '…', isError: false }} isLoading />);

    expect(screen.getByRole('region', { name: 'Display' })).toHaveAttribute('aria-busy', 'true');
    expect(screen.getByTestId('spinner')).toBeInTheDocument();
    expect(screen.getByText('Calculating')).toBeInTheDocument();
  });

  it.each([
    ['12345', undefined],
    ['12345678901', 'long'],
    ['1.23456789012345e+21', 'veryLong'],
  ])('sizes the value %j to fit', (value, expectedClass) => {
    render(<Display display={{ expression: '', value, isError: false }} isLoading={false} />);

    const output = screen.getByRole('status');
    for (const className of ['long', 'veryLong']) {
      if (className === expectedClass) {
        expect(output).toHaveClass(className);
      } else {
        expect(output).not.toHaveClass(className);
      }
    }
  });
});
