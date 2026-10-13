import type { Category, Order, Product, StoreSettings } from './types';
import { mockCategories, mockCreateOrder, mockCreateProduct, mockCustomers, mockDeleteProduct, mockOrders, mockProducts, mockQuote, mockStore, mockUpdateProduct } from './mock';
import { getAdminToken, orderCredentials, rememberOrder, rememberUser, userToken } from './auth';

const base = (process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080/api/v1').replace(/\/$/, '');
export const imageUrl = (url: string) => url.startsWith('/api/v1/products/images/') ? `${new URL(base).origin}${url}` : url;
const mockEnabled = process.env.NEXT_PUBLIC_API_MOCK === 'true';
export class ApiError extends Error {
  constructor(message: string, public status: number) { super(message); }
}
async function request<T>(path: string, init?: RequestInit, access: { admin?: boolean; token?: string } = {}): Promise<T> {
  const headers = new Headers(init?.headers);
  if (init?.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json');
  const token = access.admin ? getAdminToken() : access.token;
  if (access.admin && !token) throw new ApiError('กรุณาเข้าสู่ระบบ Admin', 403);
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const response = await fetch(`${base}${path}`, { ...init, headers, cache: 'no-store' });
  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    const messages: Record<number, string> = { 400: 'ข้อมูลไม่ถูกต้อง กรุณาตรวจสอบอีกครั้ง', 403: 'ไม่มีสิทธิ์เข้าถึง หรือ token หมดอายุ', 404: 'ไม่พบข้อมูล', 409: 'สต็อกหรือสถานะเปลี่ยนไป กรุณาตรวจสอบอีกครั้ง', 503: 'ระบบยังไม่พร้อม กรุณาติดต่อร้าน' };
    throw new ApiError(messages[response.status] || error?.error?.message || `API request failed (${response.status})`, response.status);
  }
  return response.status === 204 ? undefined as T : response.json();
}
async function fallback<T>(call: () => Promise<T>, mock: () => T | Promise<T>): Promise<T> {
  // Explicit demo mode only: live errors must never become fake success.
  return mockEnabled ? mock() : call();
}
const json = (method: string, payload?: unknown): RequestInit => ({ method, ...(payload === undefined ? {} : { body: JSON.stringify(payload) }) });
const admin = { admin: true };
export const api = {
  store: () => fallback(() => request<StoreSettings>('/store'), () => mockStore),
  categories: () => fallback(() => request<Category[]>('/categories'), () => mockCategories),
  products: (query = '') => fallback(() => request<Product[]>(`/products?limit=100${query ? `&${query}` : ''}`), () => query.includes('category=') ? mockProducts.filter(p => p.category_id === query.split('category=')[1]) : mockProducts),
  product: (slug: string) => fallback(() => request<Product>(`/products/${encodeURIComponent(slug)}`), () => mockProducts.find(p => p.slug === slug) || mockProducts[0]),
  quote: (payload: unknown) => fallback(() => request<{ subtotal_thb: number; delivery_fee_thb: number; total_thb: number }>('/checkout/quote', json('POST', payload)), () => mockQuote(payload)),
  createOrder: async (payload: unknown, idempotencyKey: string) => {
    const order = await fallback(() => request<Order>('/orders', { ...json('POST', payload), headers: { 'Idempotency-Key': idempotencyKey } }, { token: userToken() }), () => mockCreateOrder(payload));
    if (order.access_token) rememberOrder(order.id, order.access_token);
    return order;
  },
  orders: () => fallback(async () => {
    const token = userToken();
    if (token) return request<Order[]>('/orders?limit=100', undefined, { token });
    return Promise.all(Object.entries(orderCredentials()).map(([id, token]) => request<Order>(`/orders/${id}`, undefined, { token })));
  }, () => mockOrders),
  order: (id: string) => fallback(() => request<Order>(`/orders/${id}`, undefined, { token: orderCredentials()[id] || userToken() }), () => mockOrders.find(o => o.id === id) || mockOrders[0]),
  uploadPaymentSlip: (paymentId: string, file: File, orderId: string) => {
    if (file.size > 5 * 1024 * 1024 || !['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) throw new ApiError('สลิปต้องเป็น JPG, PNG หรือ WebP ไม่เกิน 5MB', 400);
    const form = new FormData(); form.append('slip', file);
    return fallback(() => request(`/payments/${paymentId}/slip`, { method: 'POST', body: form }, { token: orderCredentials()[orderId] || userToken() }), () => ({ id: paymentId, status: 'uploaded' }));
  },
  submitPayment: (paymentId: string, orderId: string) => fallback(() => request(`/payments/${paymentId}/submit`, json('POST'), { token: orderCredentials()[orderId] || userToken() }), () => ({ id: paymentId, status: 'submitted' })),
  linkLine: async (accessToken: string) => {
    const result = await request<{ access_token: string }>('/users/line/link', json('POST', { access_token: accessToken }));
    rememberUser(result.access_token); return result;
  },
  adminDashboard: () => fallback(() => request<Record<string, unknown>>('/admin/dashboard', undefined, admin), () => ({ total_sales_thb: mockOrders.reduce((sum, o) => sum + o.total_thb, 0), total_orders: mockOrders.length, pending_payments: mockOrders.filter(o => o.payment_status !== 'approved').length })),
  adminProducts: () => fallback(() => request<Product[]>('/admin/products?limit=100', undefined, admin), () => mockProducts),
  createProduct: (payload: any) => fallback(() => request<Product>('/admin/products', json('POST', payload), admin), () => mockCreateProduct(payload)),
  uploadProductImage: (id: string, file: File, altText: string) => {
    const body = new FormData(); body.set('image', file); body.set('alt_text', altText);
    return fallback(() => request<Product>(`/admin/products/${id}/images`, { method: 'POST', body }, admin), () => {
      const product = mockProducts.find(p => p.id === id);
      if (!product) throw new ApiError('ไม่พบสินค้า', 404);
      product.images.unshift({ public_url: URL.createObjectURL(file), alt_text: altText, is_primary: true });
      return product;
    });
  },
  updateProduct: (id: string, payload: unknown) => fallback(() => request<Product>(`/admin/products/${id}`, json('PATCH', payload), admin), () => mockUpdateProduct(id, payload)),
  deleteProduct: (id: string) => fallback(() => request(`/admin/products/${id}`, json('DELETE'), admin), () => { mockDeleteProduct(id); return undefined; }),
  adminOrders: (query = '') => fallback(() => request<Order[]>(`/admin/orders?limit=100${query ? `&${query}` : ''}`, undefined, admin), () => mockOrders),
  updateOrder: (id: string, payload: { internal_note: string }) => fallback(() => request<Order>(`/admin/orders/${id}`, json('PATCH', payload), admin), () => Object.assign(mockOrders.find(o => o.id === id)!, payload)),
  updateOrderStatus: (id: string, status: string, note?: string) => fallback(() => request<Order>(`/admin/orders/${id}/status`, json('PATCH', { status, note }), admin), () => { const o = mockOrders.find(o => o.id === id)!; o.order_status = status; return o; }),
  approvePayment: (id: string) => fallback(() => request(`/admin/payments/${id}/approve`, json('POST'), admin), () => ({ id, status: 'approved' })),
  rejectPayment: (id: string, reason: string) => fallback(() => request(`/admin/payments/${id}/reject`, json('POST', { reason }), admin), () => ({ id, status: 'rejected' })),
  paymentSlip: async (id: string) => {
    const response = await fetch(`${base}/admin/payments/${id}/slip`, { headers: { Authorization: `Bearer ${getAdminToken()}` }, cache: 'no-store' });
    if (!response.ok) throw new ApiError('โหลดสลิปไม่สำเร็จ', response.status);
    return response.blob();
  },
  customers: () => fallback(() => request<Record<string, unknown>[]>('/admin/customers?limit=100', undefined, admin), () => mockCustomers),
  settings: () => fallback(() => request<StoreSettings>('/admin/settings', undefined, admin), () => mockStore),
  updateSettings: (payload: StoreSettings) => fallback(() => request<StoreSettings>('/admin/settings', json('PATCH', payload), admin), () => Object.assign(mockStore, payload)),
};
