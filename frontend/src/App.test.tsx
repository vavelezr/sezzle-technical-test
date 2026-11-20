import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { App } from './App';
import { CALCULATE_URL, server } from './test/apiMock';
import { renderWithClient } from './test/render';

function setup() {
  const user = userEvent.setup();
  renderWithClient(<App />);
  const click = async (...names: string[]) => {
    for (const name of names) {
      await user.click(screen.getByRole('button', { name }));
    }
  };
  return { user, click };
}

const display = () => screen.getByRole('status');
const expression = () => screen.getByTestId('expression');

describe('App', () => {
  it('renders the header and an idle display', () => {
    setup();

    expect(screen.getByRole('heading', { name: 'Calculator' })).toBeInTheDocument();
    expect(screen.getByText('Built for Sezzle technical assessment')).toBeInTheDocument();
    expect(display()).toHaveTextContent('0');
  });

  it('calculates 12 ÷ 4 = 3 through the backend', async () => {
    const { click } = setup();

    await click('1', '2', 'Divide', '4');
    expect(expression()).toHaveTextContent('12 ÷');
    expect(display()).toHaveTextContent('4');

    await click('Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent('3');
    });
    expect(expression()).toHaveTextContent('12 ÷ 4 =');
    const history = screen.getByRole('region', { name: 'Recent' });
    expect(within(history).getByText('12 ÷ 4 =')).toBeInTheDocument();
  });

  it('shows a friendly error for division by zero', async () => {
    const { click } = setup();

    await click('5', 'Divide', '0', 'Equals');

    expect(await screen.findByText("You can't divide by zero")).toBeInTheDocument();
    expect(display()).toHaveClass('error');
  });

  it('applies square root to the current value immediately', async () => {
    const { click } = setup();

    await click('8', '1', 'Square root');

    await waitFor(() => {
      expect(display()).toHaveTextContent('9');
    });
    expect(expression()).toHaveTextContent('√(81) =');
  });

  it('supports full keyboard input', async () => {
    const { user } = setup();

    await user.keyboard('7{Backspace}9*3{Enter}');
    await waitFor(() => {
      expect(display()).toHaveTextContent('27');
    });

    await user.keyboard('{Escape}');
    expect(display()).toHaveTextContent('0');

    await user.keyboard('2.5^2=');
    await waitFor(() => {
      expect(display()).toHaveTextContent('6.25');
    });
  });

  it('ignores keyboard shortcuts with modifier keys', async () => {
    const { user } = setup();

    await user.keyboard('{Control>}5{/Control}');

    expect(display()).toHaveTextContent('0');
  });

  it('shows a loading state and prevents double submits', async () => {
    let release: () => void = () => undefined;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    let requests = 0;
    server.use(
      http.post(CALCULATE_URL, async () => {
        requests += 1;
        await gate;
        return HttpResponse.json({ operation: 'add', operands: [2, 3], result: 5 });
      }),
    );
    const { click, user } = setup();

    await click('2', 'Add', '3', 'Equals');

    expect(screen.getByRole('button', { name: 'Equals' })).toBeDisabled();
    expect(screen.getByRole('region', { name: 'Display' })).toHaveAttribute('aria-busy', 'true');
    await user.keyboard('{Enter}');

    release();
    await waitFor(() => {
      expect(display()).toHaveTextContent('5');
    });
    expect(screen.getByRole('button', { name: 'Equals' })).toBeEnabled();
    expect(requests).toBe(1);
  });

  it('shows a friendly message when the service is unreachable', async () => {
    server.use(http.post(CALCULATE_URL, () => HttpResponse.error()));
    const { click } = setup();

    await click('1', 'Add', '1', 'Equals');

    expect(
      await screen.findByText("Can't reach the calculator service. Please try again."),
    ).toBeInTheDocument();
  });

  it('chains operations using the previous result', async () => {
    const { click } = setup();

    await click('2', 'Add', '3', 'Multiply');
    await waitFor(() => {
      expect(expression()).toHaveTextContent('5 ×');
    });
    await click('4', 'Equals');

    await waitFor(() => {
      expect(display()).toHaveTextContent('20');
    });
  });
});
