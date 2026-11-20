/**
 * Significant digits shown on the display: the most a float64 can hold
 * exactly in decimal. It also hides float noise such as 0.1 + 0.2.
 */
const DISPLAY_PRECISION = 15;
const EXPONENTIAL_UPPER_BOUND = 1e15;
const EXPONENTIAL_LOWER_BOUND = 1e-6;

/** Formats a number for the display, switching to exponent notation for extremes. */
export function formatNumber(value: number): string {
  const rounded = Number(value.toPrecision(DISPLAY_PRECISION));
  const magnitude = Math.abs(rounded);
  if (
    magnitude >= EXPONENTIAL_UPPER_BOUND ||
    (magnitude > 0 && magnitude < EXPONENTIAL_LOWER_BOUND)
  ) {
    return rounded.toExponential();
  }
  // Adding 0 turns -0 into 0.
  return String(rounded + 0);
}
