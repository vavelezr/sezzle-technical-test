import { render, screen, within } from '@testing-library/react';
import { History } from './History';

describe('History', () => {
  it('renders nothing when empty', () => {
    const { container } = render(<History entries={[]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it('lists entries in the given order', () => {
    render(
      <History
        entries={[
          { id: 1, request: { operation: 'sqrt', operands: [16] }, result: 4 },
          { id: 0, request: { operation: 'divide', operands: [12, 4] }, result: 3 },
        ]}
      />,
    );

    const items = within(screen.getByRole('region', { name: 'Recent' })).getAllByRole('listitem');
    expect(items.map((item) => item.textContent)).toEqual(['√(16) =4', '12 ÷ 4 =3']);
  });
});
