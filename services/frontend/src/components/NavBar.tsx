import { NavLink, useNavigate } from "react-router-dom";
import { logout } from "../api/auth";
import { useAuth } from "../lib/AuthContext";
import { useCart } from "../lib/CartContext";

export default function NavBar() {
  const navigate = useNavigate();
  const { isAdmin: admin, setAnonymous } = useAuth();
  const { cart } = useCart();
  const itemCount = cart?.item_count ?? 0;

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
        <NavLink to="/cart">
          Корзина{itemCount > 0 && <span className="cart-badge">{itemCount}</span>}
        </NavLink>
        <NavLink to="/orders">Мои заказы</NavLink>
        <NavLink to="/profile">Профиль</NavLink>
      </div>
      <button type="button" className="btn-secondary nav-logout" onClick={handleLogout}>
        Выйти
      </button>
    </nav>
  );
}
