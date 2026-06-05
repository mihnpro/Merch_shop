import { useEffect, useState } from "react";
import { listCategories, listProducts } from "../../api/catalog";
import { createProduct, deactivateProduct, updateProduct } from "../../api/admin";
import type { Category, Product } from "../../api/types";

function parseSizes(raw: string): string[] {
  return raw
    .split(",")
    .map((s) => s.trim())
    .filter((s) => s.length > 0);
}

export default function ProductsPanel() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");

  // форма создания
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [sizes, setSizes] = useState("");
  const [photoKey, setPhotoKey] = useState("");

  async function reload() {
    setError("");
    try {
      const [cats, prods] = await Promise.all([listCategories(false), listProducts({ active_only: false, page_size: 100 })]);
      setCategories(cats.categories);
      setProducts(prods.products);
      if (!categoryId && cats.categories.length > 0) setCategoryId(cats.categories[0].id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  useEffect(() => {
    reload();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function handleCreate(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setInfo("");
    try {
      await createProduct({
        name: name.trim(),
        description: description.trim(),
        price_points: Number(price),
        category_id: categoryId,
        sizes: parseSizes(sizes),
        photo_key: photoKey.trim() || undefined,
      });
      setName("");
      setDescription("");
      setPrice("");
      setSizes("");
      setPhotoKey("");
      setInfo("Товар создан");
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  async function handleDeactivate(id: string) {
    setError("");
    setInfo("");
    try {
      await deactivateProduct(id);
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  // Возврат товара в активный статус: PUT со всеми текущими полями + active=true.
  async function handleActivate(p: Product) {
    setError("");
    setInfo("");
    try {
      await updateProduct(p.id, {
        name: p.name,
        description: p.description,
        price_points: p.price_points,
        category_id: p.category.id,
        sizes: p.sizes,
        photo_key: p.photo_key || undefined,
        active: true,
        version: p.version,
      });
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
      await reload();
    }
  }

  return (
    <div>
      <form className="panel-form" onSubmit={handleCreate}>
        <h3>Новый товар</h3>
        <div className="row">
          <label>
            Название
            <input value={name} onChange={(e) => setName(e.target.value)} required />
          </label>
          <label>
            Цена (баллы)
            <input type="number" min={1} value={price} onChange={(e) => setPrice(e.target.value)} required />
          </label>
        </div>
        <label>
          Описание
          <input value={description} onChange={(e) => setDescription(e.target.value)} required />
        </label>
        <div className="row">
          <label>
            Категория
            <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)} required>
              {categories.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            Размеры (через запятую)
            <input value={sizes} onChange={(e) => setSizes(e.target.value)} placeholder="XS, S, M, L" />
          </label>
        </div>
        <label>
          photo_key (необязательно)
          <input value={photoKey} onChange={(e) => setPhotoKey(e.target.value)} placeholder="products/<uuid>.jpg" />
        </label>
        <button type="submit">Создать товар</button>
      </form>

      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}

      <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            <th>Название</th>
            <th>Категория</th>
            <th>Цена</th>
            <th>Размеры</th>
            <th>Активен</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {products.map((p) => (
            <ProductRow
              key={p.id}
              product={p}
              categories={categories}
              onChanged={reload}
              onError={setError}
              onDeactivate={handleDeactivate}
              onActivate={handleActivate}
            />
          ))}
        </tbody>
      </table>
      </div>
    </div>
  );
}

function ProductRow({
  product,
  categories,
  onChanged,
  onError,
  onDeactivate,
  onActivate,
}: {
  product: Product;
  categories: Category[];
  onChanged: () => void;
  onError: (msg: string) => void;
  onDeactivate: (id: string) => void;
  onActivate: (p: Product) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(product.name);
  const [description, setDescription] = useState(product.description);
  const [price, setPrice] = useState(String(product.price_points));
  const [categoryId, setCategoryId] = useState(product.category.id);
  const [sizes, setSizes] = useState(product.sizes.join(", "));

  async function save() {
    onError("");
    try {
      await updateProduct(product.id, {
        name: name.trim(),
        description: description.trim(),
        price_points: Number(price),
        category_id: categoryId,
        sizes: parseSizes(sizes),
        photo_key: product.photo_key || undefined,
        active: product.active,
        version: product.version,
      });
      setEditing(false);
      onChanged();
    } catch (err) {
      onError(err instanceof Error ? err.message : "Ошибка");
      onChanged(); // подтянуть актуальную версию при конфликте
    }
  }

  if (!editing) {
    return (
      <tr>
        <td>{product.name}</td>
        <td>{product.category.name}</td>
        <td>{product.price_points}</td>
        <td>{product.sizes.join(", ") || "—"}</td>
        <td>{product.active ? "да" : "нет"}</td>
        <td className="actions">
          <button type="button" className="btn-secondary" onClick={() => setEditing(true)}>
            Изменить
          </button>
          {product.active ? (
            <button type="button" className="btn-secondary" onClick={() => onDeactivate(product.id)}>
              Деактивировать
            </button>
          ) : (
            <button type="button" className="btn-secondary" onClick={() => onActivate(product)}>
              Активировать
            </button>
          )}
        </td>
      </tr>
    );
  }

  return (
    <tr>
      <td>
        <input value={name} onChange={(e) => setName(e.target.value)} />
        <input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="описание" />
      </td>
      <td>
        <select value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
          {categories.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>
      </td>
      <td>
        <input type="number" min={1} value={price} onChange={(e) => setPrice(e.target.value)} />
      </td>
      <td>
        <input value={sizes} onChange={(e) => setSizes(e.target.value)} placeholder="XS, S, M" />
      </td>
      <td>{product.active ? "да" : "нет"}</td>
      <td className="actions">
        <button type="button" onClick={save}>
          Сохранить
        </button>
        <button type="button" className="btn-secondary" onClick={() => setEditing(false)}>
          Отмена
        </button>
      </td>
    </tr>
  );
}
