import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import NavBar from "../components/NavBar";
import { getProduct } from "../api/catalog";
import type { Product } from "../api/types";
import { photoUrl } from "../lib/media";
import { useCart } from "../lib/CartContext";

export default function ProductDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [product, setProduct] = useState<Product | null>(null);
  const [active, setActive] = useState(0);
  const [error, setError] = useState("");

  const { addItem } = useCart();
  const [qty, setQty] = useState(1);
  const [addBusy, setAddBusy] = useState(false);
  const [addSuccess, setAddSuccess] = useState(false);
  const [addError, setAddError] = useState("");

  useEffect(() => {
    if (!id) return;
    getProduct(id)
      .then((p) => {
        setProduct(p);
        setActive(0);
      })
      .catch((err) => setError(err instanceof Error ? err.message : "Ошибка"));
  }, [id]);

  async function handleAddToCart() {
    if (!product || addBusy) return;
    setAddBusy(true);
    setAddSuccess(false);
    setAddError("");
    try {
      await addItem(product.id, qty);
      setAddSuccess(true);
      setTimeout(() => setAddSuccess(false), 3000);
    } catch (e) {
      setAddError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setAddBusy(false);
    }
  }

  return (
    <div className="page">
      <NavBar />
      <Link to="/catalog" className="back-link">
        ← Назад в каталог
      </Link>

      {error && <p className="error">{error}</p>}

      {!product && !error && <p>Загрузка…</p>}

      {product && (
        <div className="product-detail">
          <div className="gallery">
            <div className="gallery-main">
              {photoUrl(product.photo_keys[active]) ? (
                <img src={photoUrl(product.photo_keys[active])} alt={product.name} />
              ) : (
                <span className="muted">нет фото</span>
              )}
            </div>
            {product.photo_keys.length > 1 && (
              <div className="gallery-thumbs">
                {product.photo_keys.map((key, i) => (
                  <button
                    key={key}
                    type="button"
                    className={`gallery-thumb${i === active ? " is-active" : ""}`}
                    onClick={() => setActive(i)}
                  >
                    <img src={photoUrl(key)} alt={`Фото ${i + 1}`} loading="lazy" />
                  </button>
                ))}
              </div>
            )}
          </div>

          <div className="product-info">
            <h1>{product.name}</h1>
            <p className="muted">{product.category.name}</p>
            <p className="price">{product.price_points} баллов</p>
            <p className="desc">{product.description}</p>
            {!product.active && <p className="muted">Товар скрыт из каталога</p>}
            {product.active && (
              <div className="add-to-cart">
                <label>
                  Количество
                  <input
                    type="number"
                    min={1}
                    max={99}
                    value={qty}
                    onChange={(e) =>
                      setQty(Math.max(1, Math.min(99, parseInt(e.target.value) || 1)))
                    }
                  />
                </label>
                <button type="button" onClick={handleAddToCart} disabled={addBusy}>
                  {addBusy ? "Добавляем…" : "В корзину"}
                </button>
                {addSuccess && <p className="info">Добавлено в корзину!</p>}
                {addError && <p className="error">{addError}</p>}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
