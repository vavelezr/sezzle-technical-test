import { useEffect } from 'react';
import type { CalculatorCommand } from './commands';
import { commandForKey } from './keyboard';

/**
 * Listens for key presses on the window and forwards bound keys as commands.
 * Shortcuts using Ctrl, Meta or Alt are left to the browser.
 */
export function useKeyboardCommands(onCommand: (command: CalculatorCommand) => void): void {
  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.metaKey || event.altKey) {
        return;
      }
      const command = commandForKey(event.key);
      if (command) {
        // Also stops Enter from clicking a focused button, which would double-dispatch.
        event.preventDefault();
        onCommand(command);
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [onCommand]);
}
