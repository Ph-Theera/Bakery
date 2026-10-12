export type Category = { id: string; slug: string; name_th: string };
export type Product = { id: string; sku?: string; category_id?: string; name: string; slug: string; description_th: string; price_thb: number; compare_at_price_thb?: number | null; stock_quantity: number; is_active: boolean; is_featured: boolean; images: { public_url: string; alt_text?: string; is_primary?: boolean }[] };
export type CartItem = { product: Product; quantity: number; specialInstructions?: string };
export type Order = { id: string; access_token?: string; order_number: string; payment_id?: string; customer_name_snapshot: string; total_thb: number; order_status: string; payment_status?: string; created_at: string; items?: { product_name_snapshot: string; quantity: number; unit_price_thb: number }[] };
export type StoreSettings = { store_name: string; description: string; line_oa_url?: string; delivery_fee_thb: number; free_delivery_threshold_thb?: number; accepting_orders: boolean };
