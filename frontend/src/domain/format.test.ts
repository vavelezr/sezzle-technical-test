import { formatNumber } from './format';

describe('formatNumber', () => {
  it.each([
    [0, '0'],
    [-0, '0'],
    [42, '42'],
    [-7.25, '-7.25'],
    [0.1 + 0.2, '0.3'],
    [1 / 3, '0.333333333333333'],
    [123456789012345, '123456789012345'],
    [1e15, '1e+15'],
    [2 ** 64, '1.84467440737096e+19'],
    [0.000001, '0.000001'],
    [0.0000001, '1e-7'],
  ])('formatNumber(%d) = %j', (value, expected) => {
    expect(formatNumber(value)).toBe(expected);
  });
});
