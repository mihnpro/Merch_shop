import { NavLink, useNavigate } from "react-router-dom";
import { logout } from "../api/auth";
import { clearTokens, getRefreshToken } from "../lib/tokens";
import { isAdmin } from "../lib/auth";

export default function NavBar() {
  const navigate = useNavigate();
  const admin = isAdmin();

  async function handleLogout() {
    const refreshToken = getRefreshToken();
    try {
      if (refreshToken) await logout(refreshToken);
    } catch {
    } finally {
      clearTokens();
      navigate("/login");
    }
  }

  return (
    <nav className="nav">
      <span className="nav-brand">MerchShop</span>
      <div className="nav-links">
        <NavLink to="/catalog">Каталог</NavLink>
        {admin && <NavLink to="/admin">Админка</NavLink>}
        <NavLink to="/profile">Профиль</NavLink>
      </div>
      <button type="button" className="btn-secondary nav-logout" onClick={handleLogout}>
        Выйти
      </button>
    </nav>
  );
}
