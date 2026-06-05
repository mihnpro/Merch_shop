import { useEffect, useState } from "react";
import { listCategories } from "../../api/catalog";
import { createCategory, updateCategory } from "../../api/admin";
import type { Category } from "../../api/types";

export default function CategoriesPanel() {
  const [categories, setCategories] = useState<Category[]>([]);
  const [code, setCode] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [info, setInfo] = useState("");

  async function reload() {
    setError("");
    try {
      const res = await listCategories(false);
      setCategories(res.categories);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  useEffect(() => {
    reload();
  }, []);

  async function handleCreate(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    setInfo("");
    try {
      await createCategory({ code: code.trim(), name: name.trim() });
      setCode("");
      setName("");
      setInfo("Категория создана");
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  async function save(category: Category, patch: Partial<Category>) {
    setError("");
    setInfo("");
    try {
      await updateCategory(category.id, {
        name: patch.name ?? category.name,
        active: patch.active ?? category.active,
      });
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка");
    }
  }

  return (
    <div>
      <form className="panel-form" onSubmit={handleCreate}>
        <h3>Новая категория</h3>
        <div className="row">
          <label>
            Код (a-z_)
            <input value={code} onChange={(e) => setCode(e.target.value)} placeholder="hoodies" required />
          </label>
          <label>
            Название
            <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Худи" required />
          </label>
        </div>
        <button type="submit">Создать категорию</button>
      </form>

      {error && <p className="error">{error}</p>}
      {info && <p className="info">{info}</p>}

      <div className="table-wrap">
      <table className="table">
        <thead>
          <tr>
            <th>Код</th>
            <th>Название</th>
            <th>Активна</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {categories.map((c) => (
            <CategoryRow key={c.id} category={c} onSave={save} />
          ))}
        </tbody>
      </table>
      </div>
    </div>
  );
}

function CategoryRow({
  category,
  onSave,
}: {
  category: Category;
  onSave: (c: Category, patch: Partial<Category>) => void;
}) {
  const [name, setName] = useState(category.name);

  return (
    <tr>
      <td>{category.code}</td>
      <td>
        <input value={name} onChange={(e) => setName(e.target.value)} />
      </td>
      <td>{category.active ? "да" : "нет"}</td>
      <td className="actions">
        <button type="button" className="btn-secondary" onClick={() => onSave(category, { name })}>
          Сохранить
        </button>
        <button type="button" className="btn-secondary" onClick={() => onSave(category, { active: !category.active })}>
          {category.active ? "Деактивировать" : "Активировать"}
        </button>
      </td>
    </tr>
  );
}
