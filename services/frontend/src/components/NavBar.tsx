import { NavLink, useNavigate } from "react-router-dom";
import { logout } from "../api/auth";
import { useAuth } from "../lib/AuthContext";

export default function NavBar() {
  const navigate = useNavigate();
  const { isAdmin: admin, setAnonymous } = useAuth();

  async function handleLogout() {
    try {
      await logout();
    } catch {
    } finally {
      setAnonymous();
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
