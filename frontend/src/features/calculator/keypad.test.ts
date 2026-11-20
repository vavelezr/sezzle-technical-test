import { OPERATIONS } from '../../domain/operations';
import { DIGIT_KEYS, OPERATION_KEYS } from './keypad';

describe('keypad key factory', () => {
  it('creates one operator key per configured operation', () => {
    expect(OPERATION_KEYS).toHaveLength(OPERATIONS.length);
    OPERATIONS.forEach((meta, i) => {
      expect(OPERATION_KEYS[i]).toEqual({
        id: meta.id,
        label: meta.symbol,
        ariaLabel: meta.label,
        command: { type: 'OPERATOR', operation: meta.id },
        variant: 'operator',
      });
    });
  });

  it('creates a key for every digit', () => {
    const digits = DIGIT_KEYS.map((key) => key.label).sort();
    expect(digits).toEqual(['0', '1', '2', '3', '4', '5', '6', '7', '8', '9']);
    expect(DIGIT_KEYS.every((key) => key.command.type === 'DIGIT')).toBe(true);
  });
});
