import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';
import { cx } from './cx';

describe('Button', () => {
  it('renders a non-submitting button that forwards clicks', async () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>7</Button>);

    const button = screen.getByRole('button', { name: '7' });
    await userEvent.click(button);

    expect(button).toHaveAttribute('type', 'button');
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('applies the variant and custom class names', () => {
    render(
      <Button variant="operator" className="extra" aria-label="Add">
        +
      </Button>,
    );

    expect(screen.getByRole('button', { name: 'Add' })).toHaveClass('button', 'operator', 'extra');
  });

  it('does not fire clicks while disabled', async () => {
    const onClick = vi.fn();
    render(
      <Button disabled onClick={onClick}>
        =
      </Button>,
    );

    await userEvent.click(screen.getByRole('button', { name: '=' }));

    expect(onClick).not.toHaveBeenCalled();
  });
});

describe('cx', () => {
  it('joins truthy class names only', () => {
    expect(cx('a', false, undefined, null, 'b', '')).toBe('a b');
  });
});
