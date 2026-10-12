import type { Category, Order, Product, StoreSettings } from './types';

const image = (seed: string) => `https://images.unsplash.com/${seed}?auto=format&fit=crop&w=900&q=85`;

export const mockStore: StoreSettings = {
  store_name: 'TongBai',
  description: 'ขนมโฮมเมดอบสดใหม่ทุกวัน ทำด้วยความใส่ใจในทุกคำ',
  delivery_fee_thb: 50,
  free_delivery_threshold_thb: 800,
  accepting_orders: true,
};

export const mockCategories: Category[] = [
  { id: 'cat-bread', slug: 'bread', name_th: 'ขนมปัง' },
  { id: 'cat-cookie', slug: 'cookie', name_th: 'คุกกี้' },
  { id: 'cat-cake', slug: 'cake', name_th: 'เค้ก' },
];

export let mockProducts: Product[] = [
  { id: 'prod-sourdough', sku: 'BRD-001', category_id: 'cat-bread', name: 'Sourdough Country Loaf', slug: 'sourdough-country-loaf', description_th: 'ซาวร์โดว์เปลือกกรอบ เนื้อเหนียวนุ่ม หอมกลิ่นหมักธรรมชาติ', price_thb: 180, compare_at_price_thb: 220, stock_quantity: 12, is_active: true, is_featured: true, images: [{ public_url: image('photo-1509440159596-0249088772ff'), alt_text: 'Sourdough bread', is_primary: true }] },
  { id: 'prod-croissant', sku: 'PAS-001', category_id: 'cat-bread', name: 'Butter Croissant', slug: 'butter-croissant', description_th: 'ครัวซองต์เนยสด กรอบบางเป็นชั้น หอมละมุน', price_thb: 95, stock_quantity: 18, is_active: true, is_featured: true, images: [{ public_url: image('photo-1555507036-ab1f4038808a'), alt_text: 'Butter croissant', is_primary: true }] },
  { id: 'prod-cookie', sku: 'CK-001', category_id: 'cat-cookie', name: 'Sea Salt Chocolate Cookie', slug: 'sea-salt-chocolate-cookie', description_th: 'คุกกี้ช็อกโกแลตเข้มข้น โรยเกลือทะเลเล็กน้อย', price_thb: 75, stock_quantity: 24, is_active: true, is_featured: false, images: [{ public_url: image('photo-1499636136210-6f4ee915583e'), alt_text: 'Chocolate cookie', is_primary: true }] },
  { id: 'prod-cake', sku: 'CKE-001', category_id: 'cat-cake', name: 'Olive Oil Lemon Cake', slug: 'olive-oil-lemon-cake', description_th: 'เค้กเลมอนน้ำมันมะกอก เนื้อนุ่มสดชื่น หวานกำลังดี', price_thb: 320, stock_quantity: 8, is_active: true, is_featured: true, images: [{ public_url: image('photo-1464195244916-405fa0dc7b3d'), alt_text: 'Lemon cake', is_primary: true }] },
];

export let mockOrders: Order[] = [
  { id: 'ord-demo-001', order_number: 'OO-26001', customer_name_snapshot: 'คุณสมชาย ใจดี', total_thb: 455, order_status: 'preparing', payment_status: 'approved', created_at: new Date().toISOString(), items: [{ product_name_snapshot: 'Butter Croissant', quantity: 2, unit_price_thb: 95 }, { product_name_snapshot: 'Sea Salt Chocolate Cookie', quantity: 3, unit_price_thb: 75 }] },
];

export const mockCustomers: Record<string, unknown>[] = [
  { id: 'cus-demo-001', display_name: 'คุณสมชาย ใจดี', phone: '0812345678', order_count: 3 },
  { id: 'cus-demo-002', display_name: 'คุณสุดา รักขนม', phone: '0898765432', order_count: 1 },
];

export function mockQuote(payload: any) {
  const subtotal_thb = (payload?.items || []).reduce((sum: number, item: any) => {
    const product = mockProducts.find((p) => p.id === item.productId);
    return sum + (product?.price_thb || 0) * Number(item.quantity || 0);
  }, 0);
  const delivery_fee_thb = subtotal_thb >= (mockStore.free_delivery_threshold_thb || Infinity) ? 0 : mockStore.delivery_fee_thb;
  return { subtotal_thb, delivery_fee_thb, total_thb: subtotal_thb + delivery_fee_thb };
}

export function mockCreateOrder(payload: any): Order {
  const quote = mockQuote(payload);
  const order: Order = { id: `ord-demo-${Date.now()}`, order_number: `OO-${String(mockOrders.length + 26002).padStart(5, '0')}`, payment_id: `pay-demo-${Date.now()}`, customer_name_snapshot: payload?.customer?.name || 'ลูกค้าทดลอง', total_thb: quote.total_thb, order_status: 'pending_payment', payment_status: 'pending', created_at: new Date().toISOString(), items: (payload?.items || []).map((item: any) => { const p = mockProducts.find((product) => product.id === item.productId); return { product_name_snapshot: p?.name || 'สินค้า', quantity: item.quantity, unit_price_thb: p?.price_thb || 0 }; }) };
  mockOrders = [order, ...mockOrders];
  return order;
}

export function mockCreateProduct(payload: any): Product {
  const product = { ...payload, id: `prod-demo-${Date.now()}`, images: payload.images || [] } as Product;
  mockProducts = [product, ...mockProducts];
  return product;
}

export function mockUpdateProduct(id: string, payload: any): Product {
  const product = mockProducts.find((p) => p.id === id) || mockProducts[0];
  Object.assign(product, payload);
  return product;
}

export function mockDeleteProduct(id: string) {
  mockProducts = mockProducts.filter((p) => p.id !== id);
}
