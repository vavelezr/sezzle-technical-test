import type { ComponentPropsWithoutRef } from 'react';
import styles from './Button.module.css';
import { cx } from './cx';

export type ButtonVariant = 'digit' | 'operator' | 'action' | 'equals';

export interface ButtonProps extends ComponentPropsWithoutRef<'button'> {
  readonly variant?: ButtonVariant;
}

/** Calculator key. Purely presentational: behavior comes from the props. */
export function Button({ variant = 'digit', className, type = 'button', ...props }: ButtonProps) {
  return (
    <button type={type} className={cx(styles.button, styles[variant], className)} {...props} />
  );
}
