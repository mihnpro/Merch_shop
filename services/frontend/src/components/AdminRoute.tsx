import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../lib/AuthContext";

export default function AdminRoute({ children }: { children: ReactNode }) {
  const { status, isAdmin } = useAuth();

  if (status === "loading") {
    return <p className="info">Загрузка…</p>;
  }
  if (status === "anonymous") {
    return <Navigate to="/login" replace />;
  }
  if (!isAdmin) {
    return <Navigate to="/catalog" replace />;
  }
  return <>{children}</>;
}
