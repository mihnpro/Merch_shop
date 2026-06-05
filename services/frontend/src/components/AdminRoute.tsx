import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { getAccessToken } from "../lib/tokens";
import { isAdmin } from "../lib/auth";

export default function AdminRoute({ children }: { children: ReactNode }) {
  if (!getAccessToken()) {
    return <Navigate to="/login" replace />;
  }
  if (!isAdmin()) {
    return <Navigate to="/catalog" replace />;
  }
  return <>{children}</>;
}
