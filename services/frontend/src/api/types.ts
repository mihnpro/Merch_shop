export interface User {
  id: string;
  login: string;
  first_name: string;
  last_name: string;
  patronymic?: string;
  email: string;
  phone_number?: string;
}

export interface RegisterBody {
  login: string;
  password: string;
  first_name: string;
  last_name: string;
  email: string;
  patronymic?: string;
  phone_number?: string;
}

export interface LoginBody {
  login: string;
  password: string;
}

export interface LoginResponse {
  user: User;
}

export interface MeResponse {
  user_id: string;
  role: string;
}

export interface Category {
  id: string;
  code: string;
  name: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Product {
  id: string;
  name: string;
  description: string;
  price_points: number;
  category: Category;
  photo_keys: string[];
  active: boolean;
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ListProductsResponse {
  products: Product[];
  next_page_token?: string;
}

export interface CategoriesResponse {
  categories: Category[];
}

export interface ListProductsParams {
  category_id?: string;
  search?: string;
  active_only?: boolean;
  page_size?: number;
  page_token?: string;
}

export interface ProductInput {
  name: string;
  description: string;
  price_points: number;
  category_id: string;
  photo_keys: string[];
}

export interface UpdateProductInput extends ProductInput {
  active: boolean;
  version: number;
}

export interface CreateCategoryInput {
  code: string;
  name: string;
}

export interface UpdateCategoryInput {
  name: string;
  active: boolean;
}

export interface UploadPhotoResponse {
  photo_key: string;
}
