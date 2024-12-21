type ErrorSeverity = "info" | "warning" | "critical";

/**
 * Visually reports notification to user
 *
 */
export async function showNotification(message: string) {
  alert(message);
}

/**
 * Alerts the user about the error and logs it to the console
 *
 */
export async function reportError(message: string, severity: ErrorSeverity) {
  const prefix = `Error (${severity}):`;
  switch (severity) {
    case "info":
      console.info(prefix, message);
      break;
    case "warning":
      console.warn(prefix, message);
      break;
    case "critical":
      console.error(prefix, message);
      break;
  }
  // FIX: This won't work properly is message is not a `string`
  alert(`${prefix} ${message}`);
}
