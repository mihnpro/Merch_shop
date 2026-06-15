
export type Role = "user" | "admin";

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
  role: Role;
}

export interface UserProfile {
  id: string;
  login: string;
  first_name: string;
  last_name: string;
  patronymic?: string;
  email: string;
  phone_number?: string;
  role: Role;
  status: string;
  last_login_at?: string;
  created_at?: string;
  updated_at?: string;
}

export interface UpdateProfileBody {
  first_name: string;
  last_name: string;
  patronymic?: string;
  email: string;
  phone_number?: string;
}

export interface ChangePasswordBody {
  old_password: string;
  new_password: string;
}

export interface Transaction {
  id: string;
  user_id: string;
  operation_id: string;
  order_id?: string;
  amount: number;
  reason: string;
  created_at: string;
}

export interface TransactionsResponse {
  transactions: Transaction[];
  next_page_token?: string;
}

export interface TransactionsParams {
  page_size?: number;
  page_token?: string;
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

export interface InventoryAdjustInput {
  product_id: string;
  delta: number;
  operation_id: string;
  reason: string;
}

export interface InventoryAdjustResult {
  product_id: string;
  available: number;
  version: number;
}

export interface StockItem {
  product_id: string;
  available: number;
  version: number;
}

export interface ListStockResponse {
  items: StockItem[];
}

export interface CartItem {
  id: string;
  product_id: string;
  product_name: string;
  quantity: number;
  price_per_unit: number;
  line_total: number;
}

export interface CartView {
  items: CartItem[];
  total: number;
  item_count: number;
}


export type OrderStatus =
  | "pending"
  | "confirmed"
  | "ready_to_pickup"
  | "delivered"
  | "cancelled";

export interface OrderItemView {
  id: string;
  product_id: string;
  product_name: string;
  quantity: number;
  price_points: number;
}

export interface OrderView {
  id: string;
  user_id: string;
  status: OrderStatus;
  total_points: number;
  note?: string;
  cancel_reason?: string;
  delivery_address: string;
  items: OrderItemView[];
  created_at: string;
  updated_at: string;
}

export interface ListOrdersResponse {
  orders: OrderView[];
  next_page_token?: string;
}

export interface ListOrdersParams {
  status?: string;
  page_size?: number;
  page_token?: string;
}


export interface AdminUser {
  id: string;
  login: string;
  first_name: string;
  last_name: string;
  patronymic?: string;
  email: string;
  phone_number?: string;
  role: Role;
  status: string;
  last_login_at?: string;
  created_at?: string;
  updated_at?: string;
}

export interface ListUsersResponse {
  users: AdminUser[];
  next_page_token?: string;
}

export interface ListUsersParams {
  search?: string;
  role?: string;
  status?: string;
  page_size?: number;
  page_token?: string;
}

export interface GrantPointsBody {
  amount: number;
  operation_id: string;
  reason: string;
}

export interface BalanceResponse {
  points: number;
  updated_at: string;
}

export interface ResetPasswordResponse {
  new_password: string;
}

export interface BlockUserBody {
  blocked: boolean;
  reason?: string;
}

export interface ChangeRoleBody {
  role: Role;
}

export interface AdminListOrdersParams {
  user_id?: string;
  status?: string;
  page_size?: number;
  page_token?: string;
}

export interface UpdateOrderStatusBody {
  status: OrderStatus;
  reason?: string;
}

export type AnalyticsPeriod = "day" | "week" | "month";

export interface TopProduct {
  product_id: string;
  product_name: string;
  quantity: number;
}

export interface AnalyticsView {
  period: AnalyticsPeriod;
  orders_count: number;
  points_spent: number;
  average_order_value: number;
  top_products: TopProduct[];
}

export interface UsersStats {
  new_users: number;
}
