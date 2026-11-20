/** Joins truthy class names: cx('a', isActive && 'b') -> "a b". */
export function cx(...classNames: (string | false | null | undefined)[]): string {
  return classNames.filter(Boolean).join(' ');
}
