import { useRef, useState } from "react";
import { uploadPhoto } from "../api/admin";
import { photoUrl } from "../lib/media";

const ALLOWED_TYPES = ["image/jpeg", "image/png", "image/webp"];
const MAX_SIZE = 5 * 1024 * 1024;

type Props = {
  value: string[];
  onChange: (keys: string[]) => void;
};

export default function PhotoGalleryUpload({ value, onChange }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [dragOver, setDragOver] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState("");

  async function handleFiles(files: FileList) {
    setError("");
    setUploading(true);
    const added: string[] = [];
    try {
      for (const file of Array.from(files)) {
        if (!ALLOWED_TYPES.includes(file.type)) {
          setError("Допустимы только JPEG, PNG или WEBP");
          continue;
        }
        if (file.size > MAX_SIZE) {
          setError(`Файл «${file.name}» больше 5 МБ`);
          continue;
        }
        const res = await uploadPhoto(file);
        added.push(res.photo_key);
      }
      if (added.length > 0) {
        onChange([...value, ...added]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Ошибка загрузки");
    } finally {
      setUploading(false);
    }
  }

  function onDrop(e: React.DragEvent) {
    e.preventDefault();
    setDragOver(false);
    if (e.dataTransfer.files?.length) void handleFiles(e.dataTransfer.files);
  }

  function onInputChange(e: React.ChangeEvent<HTMLInputElement>) {
    if (e.target.files?.length) void handleFiles(e.target.files);
    e.target.value = "";
  }

  function removeAt(index: number) {
    onChange(value.filter((_, i) => i !== index));
  }

  return (
    <div className="photo-upload">
      <span className="photo-upload-label">Фотографии</span>

      {value.length > 0 && (
        <div className="photo-thumbs">
          {value.map((key, i) => (
            <div key={key} className="photo-thumb">
              <img src={photoUrl(key)} alt={`Фото ${i + 1}`} />
              {i === 0 && <span className="photo-thumb-badge">обложка</span>}
              <button
                type="button"
                className="photo-thumb-remove"
                aria-label="Удалить фото"
                onClick={() => removeAt(i)}
              >
                ×
              </button>
            </div>
          ))}
        </div>
      )}

      <div
        className={`photo-dropzone${dragOver ? " is-dragover" : ""}`}
        onClick={() => inputRef.current?.click()}
        onDragOver={(e) => {
          e.preventDefault();
          setDragOver(true);
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={onDrop}
        role="button"
        tabIndex={0}
      >
        {uploading ? "Загрузка…" : "Перетащите фото сюда или нажмите, чтобы выбрать (можно несколько)"}
      </div>

      <input
        ref={inputRef}
        type="file"
        accept="image/jpeg,image/png,image/webp"
        multiple
        hidden
        onChange={onInputChange}
      />

      {error && <span className="error">{error}</span>}
    </div>
  );
}
