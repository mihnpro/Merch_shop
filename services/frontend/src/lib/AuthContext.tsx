import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { getMe } from "../api/auth";
import type { Role } from "../api/types";

type AuthStatus = "loading" | "authenticated" | "anonymous";

interface AuthState {
  status: AuthStatus;
  role: Role | null;
  userId: string | null;
}

interface AuthContextValue extends AuthState {
  isAdmin: boolean;
  refresh: () => Promise<void>;
  setAnonymous: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const ANONYMOUS: AuthState = { status: "anonymous", role: null, userId: null };

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    status: "loading",
    role: null,
    userId: null,
  });

  const refresh = useCallback(async () => {
    try {
      const me = await getMe();
      setState({ status: "authenticated", role: me.role, userId: me.user_id });
    } catch {
      setState(ANONYMOUS);
    }
  }, []);

  const setAnonymous = useCallback(() => setState(ANONYMOUS), []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const value: AuthContextValue = {
    ...state,
    isAdmin: state.role === "admin",
    refresh,
    setAnonymous,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
