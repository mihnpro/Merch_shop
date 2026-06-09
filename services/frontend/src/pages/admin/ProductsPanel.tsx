import { useEffect, useState } from "react";
import { listCategories, listProducts } from "../../api/catalog";
import {
  adjustInventory,
  createProduct,
  deactivateProduct,
  listInventory,
  updateProduct,
} from "../../api/admin";
import type { Category, Product } from "../../api/types";
import PhotoGalleryUpload from "../../components/PhotoGalleryUpload";

export default function ProductsPanel() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [stock, setStock] = useState<Record<string, number>>({});
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");

  // форма создания
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [price, setPrice] = useState("");
  const [quantity, setQuantity] = useState("0");
  const [categoryId, setCategoryId] = useState("");
  const [photoKeys, setPhotoKeys] = useState<string[]>([]);

  async function reload() {
    setError("");
    try {
      const [cats, prods, inv] = await Promise.all([
        listCategories(false),
        listProducts({ active_only: false, page_size: 100 }),
        listInventory(),
      ]);
      setCategories(cats.categories);
      setProducts(prods.products);
      const map: Record<string, number> = {};
      for (const item of inv.items) map[item.product_id] = item.available;
      setStock(map);
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
      const created = await createProduct({
        name: name.trim(),
        description: description.trim(),
        price_points: Number(price),
        category_id: categoryId,
        photo_keys: photoKeys,
      });
      // Начальный остаток задаём отдельным вызовом в inventory-сервис.
      const initial = Number(quantity);
      if (Number.isInteger(initial) && initial > 0) {
        await adjustInventory({
          product_id: created.id,
          delta: initial,
          operation_id: crypto.randomUUID(),
          reason: "Начальный остаток при создании товара",
        });
      }
      setName("");
      setDescription("");
      setPrice("");
      setQuantity("0");
      setPhotoKeys([]);
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

  async function handleActivate(p: Product) {
    setError("");
    setInfo("");
    try {
      await updateProduct(p.id, {
        name: p.name,
        description: p.description,
        price_points: p.price_points,
        category_id: p.category.id,
        photo_keys: p.photo_keys,
        active: true,
        version: p.version,
      });
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
      await reload();
    }
  }

  // Установка абсолютного остатка: дельту считаем от текущего значения.
  async function handleSetStock(product: Product, target: number) {
    setError("");
    setInfo("");
    if (!Number.isInteger(target) || target < 0) {
      setError("Количество должно быть целым числом ≥ 0");
      return;
    }
    const current = stock[product.id] ?? 0;
    const delta = target - current;
    if (delta === 0) return;
    try {
      const res = await adjustInventory({
        product_id: product.id,
        delta,
        operation_id: crypto.randomUUID(),
        reason: "Корректировка остатка через админку",
      });
      setInfo(`Остаток обновлён: ${product.name} → ${res.available}`);
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
          <label>
            Количество
            <input type="number" min={0} value={quantity} onChange={(e) => setQuantity(e.target.value)} />
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
        </div>
        <PhotoGalleryUpload value={photoKeys} onChange={setPhotoKeys} />
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
              <th>Количество</th>
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
                available={stock[p.id] ?? 0}
                onChanged={reload}
                onError={setError}
                onDeactivate={handleDeactivate}
                onActivate={handleActivate}
                onSetStock={handleSetStock}
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
  available,
  onChanged,
  onError,
  onDeactivate,
  onActivate,
  onSetStock,
}: {
  product: Product;
  categories: Category[];
  available: number;
  onChanged: () => void;
  onError: (msg: string) => void;
  onDeactivate: (id: string) => void;
  onActivate: (p: Product) => void;
  onSetStock: (p: Product, target: number) => void;
}) {
  const [editing, setEditing] = useState(false);
  const [name, setName] = useState(product.name);
  const [description, setDescription] = useState(product.description);
  const [price, setPrice] = useState(String(product.price_points));
  const [categoryId, setCategoryId] = useState(product.category.id);
  const [photoKeys, setPhotoKeys] = useState<string[]>(product.photo_keys);

  async function save() {
    onError("");
    try {
      await updateProduct(product.id, {
        name: name.trim(),
        description: description.trim(),
        price_points: Number(price),
        category_id: categoryId,
        photo_keys: photoKeys,
        active: product.active,
        version: product.version,
      });
      setEditing(false);
      onChanged();
    } catch (err) {
      onError(err instanceof Error ? err.message : "Ошибка");
      onChanged();
    }
  }

  const stockCell = (
    <StockCell available={available} onSet={(target) => onSetStock(product, target)} />
  );

  if (!editing) {
    return (
      <tr>
        <td>{product.name}</td>
        <td>{product.category.name}</td>
        <td>{product.price_points}</td>
        <td>{stockCell}</td>
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
        <PhotoGalleryUpload value={photoKeys} onChange={setPhotoKeys} />
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
      <td>{stockCell}</td>
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

// Управление остатком: абсолютное значение, дельту считает родитель.
function StockCell({ available, onSet }: { available: number; onSet: (target: number) => void }) {
  const [value, setValue] = useState(String(available));

  useEffect(() => {
    setValue(String(available));
  }, [available]);

  return (
    <div style={{ display: "flex", gap: 6, alignItems: "center" }}>
      <input
        type="number"
        min={0}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        style={{ width: 70 }}
      />
      <button
        type="button"
        className="btn-secondary"
        disabled={Number(value) === available}
        onClick={() => onSet(Number(value))}
      >
        OK
      </button>
    </div>
  );
}
