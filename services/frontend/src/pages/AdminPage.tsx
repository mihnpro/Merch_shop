import { useState } from "react";
import NavBar from "../components/NavBar";
import ProductsPanel from "./admin/ProductsPanel";
import CategoriesPanel from "./admin/CategoriesPanel";
import UsersPanel from "./admin/UsersPanel";

type Tab = "products" | "categories" | "users";

export default function AdminPage() {
  const [tab, setTab] = useState<Tab>("products");

  return (
    <div className="page">
      <NavBar />
      <h1>Админ-панель</h1>

      <div className="tabs">
        <button
          type="button"
          className={tab === "products" ? "tab active" : "tab"}
          onClick={() => setTab("products")}
        >
          Товары
        </button>
        <button
          type="button"
          className={tab === "categories" ? "tab active" : "tab"}
          onClick={() => setTab("categories")}
        >
          Категории
        </button>
        <button
          type="button"
          className={tab === "users" ? "tab active" : "tab"}
          onClick={() => setTab("users")}
        >
          Пользователи
        </button>
      </div>

      {tab === "products" && <ProductsPanel />}
      {tab === "categories" && <CategoriesPanel />}
      {tab === "users" && <UsersPanel />}
    </div>
  );
}
