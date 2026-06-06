import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import NavBar from "../components/NavBar";
import { listCategories, listProducts } from "../api/catalog";
import type { Category, Product } from "../api/types";
import { photoUrl } from "../lib/media";

export default function CatalogPage() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [nextToken, setNextToken] = useState<string | undefined>();
  const [categoryId, setCategoryId] = useState("");
  const [search, setSearch] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    listCategories()
      .then((res) => setCategories(res.categories))
      .catch((err) => setError(err instanceof Error ? err.message : "Ошибка"));
  }, []);

  async function load(append = false, token?: string) {
    setError("");
    setLoading(true);
    try {
      const res = await listProducts({
        active_only: true,
        category_id: categoryId || undefined,
        search: search || undefined,
        page_token: token,
      });
      setProducts((prev) => (append ? [...prev, ...res.products] : res.products));
      setNextToken(res.next_page_token || undefined);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load(false);
  }, [categoryId]);

  function handleSearch(event: React.FormEvent) {
    event.preventDefault();
    load(false);
  }

  return (
    <div className="page">
      <NavBar />
      <h1>Каталог</h1>

      <form className="toolbar" onSubmit={handleSearch}>
        <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          <option value="">Все категории</option>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
        <input
          placeholder="Поиск по названию"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <button type="submit" disabled={loading}>
          Найти
        </button>
      </form>

      {error && <p className="error">{error}</p>}

      {products.length === 0 && !loading && !error && <p>Товаров пока нет.</p>}

      <div className="grid">
        {products.map((p) => (
          <Link key={p.id} to={`/catalog/${p.id}`} className="product-card">
            <div className="product-photo">
              {photoUrl(p.photo_keys[0]) ? (
                <img src={photoUrl(p.photo_keys[0])} alt={p.name} loading="lazy" />
              ) : (
                "нет фото"
              )}
            </div>
            <h3>{p.name}</h3>
            <p className="muted">{p.category.name}</p>
            <p className="price">{p.price_points} баллов</p>
            {p.sizes.length > 0 && (
              <div className="chips">
                {p.sizes.map((s) => (
                  <span key={s} className="chip">
                    {s}
                  </span>
                ))}
              </div>
            )}
            <p className="desc">{p.description}</p>
          </Link>
        ))}
      </div>

      {nextToken && (
        <button type="button" className="btn-secondary load-more" onClick={() => load(true, nextToken)} disabled={loading}>
          {loading ? "Загрузка…" : "Показать ещё"}
        </button>
      )}
    </div>
  );
}
