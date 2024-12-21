type AuthContext = {
  sessionKey: string;
  username: string;
  isAuthenticated: boolean;
};

export function newAuthContext(
  sessionKey?: string,
  username?: string,
): AuthContext {
  return {
    sessionKey: sessionKey || "",
    username: username || "Guest",
    isAuthenticated: false,
  };
}

export async function authenticateSession(context: AuthContext) {
  if (context.isAuthenticated) return;

  const res = await fetch("/auth/authenticate", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ sessionKey: context.sessionKey }),
  });

  if (!res.ok) {
    throw new Error(`Failed to authenticate: ${await res.text()}`);
  }
}

export async function endSession() {
  const res = await fetch("/auth/invalidate-token", {
    method: "POST",
  });
  if (!res.ok) {
    throw new Error(await res.text());
  }
}
