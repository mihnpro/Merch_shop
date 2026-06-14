import { useState } from "react";
import NavBar from "../components/NavBar";
import ProductsPanel from "./admin/ProductsPanel";
import CategoriesPanel from "./admin/CategoriesPanel";
import UsersPanel from "./admin/UsersPanel";
import OrdersPanel from "./admin/OrdersPanel";
import AnalyticsPanel from "./admin/AnalyticsPanel";

type Tab = "products" | "categories" | "users" | "orders" | "analytics";

const TABS: { value: Tab; label: string }[] = [
  { value: "products", label: "Товары" },
  { value: "categories", label: "Категории" },
  { value: "users", label: "Пользователи" },
  { value: "orders", label: "Заказы" },
  { value: "analytics", label: "Аналитика" },
];

export default function AdminPage() {
  const [tab, setTab] = useState<Tab>("products");

  return (
    <div className="page">
      <NavBar />
      <h1>Админ-панель</h1>

      <div className="tabs">
        {TABS.map((t) => (
          <button
            key={t.value}
            type="button"
            className={tab === t.value ? "tab active" : "tab"}
            onClick={() => setTab(t.value)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === "products" && <ProductsPanel />}
      {tab === "categories" && <CategoriesPanel />}
      {tab === "users" && <UsersPanel />}
      {tab === "orders" && <OrdersPanel />}
      {tab === "analytics" && <AnalyticsPanel />}
    </div>
  );
}
