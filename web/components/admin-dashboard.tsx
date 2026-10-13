'use client';
import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import { setAdminToken } from '../lib/auth';
import type { Category, Order, Product, StoreSettings } from '../lib/types';

type Tab = 'overview' | 'orders' | 'products' | 'payments' | 'customers' | 'settings';
type Draft = { id?: string; category_id: string; name: string; slug: string; description_th: string; price_thb: number; stock_quantity: number; is_active: boolean; is_featured: boolean };
const empty: Draft = { category_id: '', name: '', slug: '', description_th: '', price_thb: 0, stock_quantity: 0, is_active: true, is_featured: false };
export default function AdminDashboard() {
  const [tab, setTab] = useState<Tab>('overview');
  const [loggedIn, setLoggedIn] = useState(false); const [token, setToken] = useState('');
  const [products, setProducts] = useState<Product[]>([]); const [categories, setCategories] = useState<Category[]>([]);
  const [orders, setOrders] = useState<Order[]>([]); const [customers, setCustomers] = useState<Record<string, unknown>[]>([]);
  const [settings, setSettings] = useState<StoreSettings | null>(null); const [dashboard, setDashboard] = useState<Record<string, unknown>>({});
  const [error, setError] = useState(''); const [busy, setBusy] = useState(false);
  const [draft, setDraft] = useState<Draft | null>(null); const [preview, setPreview] = useState('');
  useEffect(() => () => { if (preview) URL.revokeObjectURL(preview); }, [preview]);
  const load = async () => {
    const [d, o, p, c, s, cats] = await Promise.all([api.adminDashboard(), api.adminOrders(), api.adminProducts(), api.customers(), api.settings(), api.categories()]);
    setDashboard(d); setOrders(o); setProducts(p); setCustomers(c); setSettings(s); setCategories(cats);
  };
  const run = async (action: () => Promise<unknown>) => {
    setBusy(true); setError('');
    try { await action(); } catch (e) { setError(e instanceof Error ? e.message : 'ดำเนินการไม่สำเร็จ'); }
    finally { setBusy(false); }
  };
  const login = () => run(async () => { setAdminToken(token); await load(); setLoggedIn(true); setToken(''); });
  const logout = () => { setAdminToken(''); setLoggedIn(false); setProducts([]); setOrders([]); setCustomers([]); setSettings(null); setDraft(null); setPreview(''); setError(''); };
  const updateStatus = (order: Order, status: string) => run(async () => { await api.updateOrderStatus(order.id, status); await load(); });
  const removeProduct = (product: Product) => { if (confirm(`ลบ ${product.name} หรือไม่?`)) void run(async () => { await api.deleteProduct(product.id); await load(); }); };
  const saveProduct = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); if (!draft) return;
    void run(async () => {
      const { id, ...payload } = draft;
      if (id) await api.updateProduct(id, payload); else await api.createProduct(payload);
      setDraft(null); await load();
    });
  };
  const saveSettings = (event: React.FormEvent<HTMLFormElement>) => { event.preventDefault(); if (settings) void run(async () => { await api.updateSettings(settings); await load(); }); };
  const review = (order: Order, approve: boolean) => {
    if (!order.payment_id) return;
    const reason = approve ? '' : prompt('เหตุผลที่ปฏิเสธการชำระเงิน');
    if (!approve && !reason?.trim()) return;
    if (approve && !confirm('ตรวจสอบสลิปและยอดเงินเข้าจริงแล้ว ยืนยันอนุมัติหรือไม่?')) return;
    void run(async () => { if (approve) await api.approvePayment(order.payment_id!); else await api.rejectPayment(order.payment_id!, reason!); await load(); });
  };
  if (!loggedIn) return <main className="page container" style={{ padding: '60px 0' }}>
    <a className="brand" href="/">TongBai</a><form className="panel admin-form" onSubmit={e => { e.preventDefault(); void login(); }}>
      <h1>เข้าสู่ระบบ Admin</h1><p>ใช้ ADMIN_API_TOKEN ที่ตั้งไว้ใน backend สำหรับผู้ดูแลที่ได้รับอนุญาตเท่านั้น</p>
      <label>Admin token</label><input type="password" autoComplete="off" required value={token} onChange={e => setToken(e.target.value)}/>
      <p className="muted">เก็บ token ในหน่วยความจำเท่านั้น รีเฟรชหน้าต้องเข้าสู่ระบบใหม่ ควรใช้ HTTPS เมื่อเปิดใช้งานจริง</p>
      <button className="button primary" disabled={busy || !token.trim()}>{busy ? 'กำลังตรวจสอบ...' : 'เข้าสู่ระบบ'}</button>
      {error && <p className="error" role="alert">{error}</p>}
    </form>
  </main>;
  return <main className="page container" style={{ padding: '35px 0' }}>
    <div className="section-head"><div><div className="eyebrow">ADMIN</div><h1>จัดการร้านค้า</h1></div><button className="button secondary" disabled={busy} onClick={logout}>ออกจากระบบ</button></div>
    {error && <div className="error" role="alert">{error}</div>}
    <nav className="admin-tabs">{(['overview', 'orders', 'products', 'payments', 'customers', 'settings'] as Tab[]).map(item => <button className={tab === item ? 'selected' : ''} key={item} onClick={() => setTab(item)}>{({ overview: 'ภาพรวม', orders: 'คำสั่งซื้อ', products: 'สินค้า', payments: 'การชำระเงิน', customers: 'ลูกค้า', settings: 'ตั้งค่าร้าน' })[item]}</button>)}</nav>
    <button className="button secondary" disabled={busy} onClick={() => run(load)}>รีเฟรชข้อมูล</button>
    <p className="muted">แสดงข้อมูลล่าสุดสูงสุด 100 รายการต่อหมวด</p>
    {tab === 'overview' && <><div className="admin-grid">
      <div className="panel"><span>ยอดขายรวม</span><strong>฿{Number(dashboard.total_sales_thb || 0).toLocaleString('th-TH')}</strong></div>
      <div className="panel"><span>คำสั่งซื้อ</span><strong>{String(dashboard.total_orders || 0)}</strong></div>
      <div className="panel"><span>รอตรวจสอบ</span><strong>{String(dashboard.pending_payments || 0)}</strong></div>
    </div><OrderTable orders={orders.slice(0, 8)} onStatus={updateStatus} busy={busy}/></>}
    {tab === 'orders' && <OrderTable orders={orders} onStatus={updateStatus} busy={busy}/>}
    {tab === 'products' && <section className="panel">
      <div className="section-head"><h2>สินค้า</h2><button className="button primary" disabled={busy} onClick={() => setDraft({ ...empty, category_id: categories[0]?.id || '' })}>เพิ่มสินค้า</button></div>
      {draft && <form className="admin-form" onSubmit={saveProduct}>
        <h3>{draft.id ? 'แก้ไขสินค้า' : 'เพิ่มสินค้า'}</h3>
        <label>หมวดหมู่</label><select required value={draft.category_id} onChange={e => setDraft({ ...draft, category_id: e.target.value })}><option value="">เลือกหมวดหมู่</option>{categories.map(c => <option key={c.id} value={c.id}>{c.name_th}</option>)}</select>
        <label>ชื่อสินค้า</label><input required maxLength={200} value={draft.name} onChange={e => setDraft({ ...draft, name: e.target.value })}/>
        <label>Slug</label><input required value={draft.slug} onChange={e => setDraft({ ...draft, slug: e.target.value })}/>
        <label>คำอธิบาย</label><textarea value={draft.description_th} onChange={e => setDraft({ ...draft, description_th: e.target.value })}/>
        <label>ราคา (บาท)</label><input type="number" min="0" step="0.01" required value={draft.price_thb} onChange={e => setDraft({ ...draft, price_thb: Number(e.target.value) })}/>
        <label>สต็อก</label><input type="number" min="0" step="1" required value={draft.stock_quantity} onChange={e => setDraft({ ...draft, stock_quantity: Number(e.target.value) })}/>
        <label className="checkbox"><input type="checkbox" checked={draft.is_active} onChange={e => setDraft({ ...draft, is_active: e.target.checked })}/>เปิดขาย</label>
        <label className="checkbox"><input type="checkbox" checked={draft.is_featured} onChange={e => setDraft({ ...draft, is_featured: e.target.checked })}/>สินค้าแนะนำ</label>
        <p className="muted">API ยังไม่รองรับอัปโหลดรูปสินค้า รูปที่มีอยู่ยังคงเดิม</p>
        <button className="button primary" disabled={busy || !draft.category_id}>บันทึก</button><button type="button" className="button secondary" onClick={() => setDraft(null)}>ยกเลิก</button>
      </form>}
      <table className="table"><thead><tr><th>สินค้า</th><th>ราคา</th><th>สต็อก</th><th>สถานะ</th><th>จัดการ</th></tr></thead><tbody>{products.map(p => <tr key={p.id}><td>{p.name}</td><td>฿{p.price_thb}</td><td>{p.stock_quantity}</td><td>{p.is_active ? 'เปิดขาย' : 'ปิดขาย'}</td><td><button disabled={busy} onClick={() => setDraft({ id: p.id, category_id: p.category_id || '', name: p.name, slug: p.slug, description_th: p.description_th, price_thb: p.price_thb, stock_quantity: p.stock_quantity, is_active: p.is_active, is_featured: p.is_featured })}>แก้ไข</button><button disabled={busy} onClick={() => removeProduct(p)}>ลบ</button></td></tr>)}</tbody></table>
    </section>}
    {tab === 'payments' && <section className="panel"><h2>สลิปรอตรวจสอบ</h2><p>อนุมัติหลังตรวจสอบยอดเงินจริงเท่านั้น การอัปโหลดสลิปไม่ได้ยืนยันเงินเข้า</p>
      {orders.filter(o => o.payment_status === 'submitted').map(o => <div className="panel" key={o.id}><h3>{o.order_number}</h3><p>{o.customer_name_snapshot} · ฿{o.total_thb.toLocaleString('th-TH')}</p>
        <button disabled={busy || !o.payment_id} onClick={() => run(async () => { const blob = await api.paymentSlip(o.payment_id!); setPreview(URL.createObjectURL(blob)); })}>ดูสลิป</button>
        <button disabled={busy} onClick={() => review(o, true)}>อนุมัติ</button><button disabled={busy} onClick={() => review(o, false)}>ปฏิเสธ</button>
      </div>)}
      {!orders.some(o => o.payment_status === 'submitted') && <p>ไม่มีรายการรอตรวจสอบในข้อมูลที่โหลด</p>}
      {preview && <div className="panel"><button onClick={() => setPreview('')}>ปิดสลิป</button><img src={preview} alt="สลิปชำระเงิน" style={{ maxWidth: '100%', maxHeight: 600 }}/></div>}
    </section>}
    {tab === 'customers' && <table className="table"><thead><tr><th>ชื่อ</th><th>โทรศัพท์</th><th>ออเดอร์</th></tr></thead><tbody>{customers.map((c, i) => <tr key={String(c.id || i)}><td>{String(c.display_name || c.name || '-')}</td><td>{String(c.phone || '-')}</td><td>{String(c.order_count || 0)}</td></tr>)}</tbody></table>}
    {tab === 'settings' && settings && <form className="panel admin-form" onSubmit={saveSettings}>
      <h2>ตั้งค่าร้าน</h2><label>ชื่อร้าน</label><input required value={settings.store_name} onChange={e => setSettings({ ...settings, store_name: e.target.value })}/>
      <label>คำอธิบาย</label><textarea value={settings.description} onChange={e => setSettings({ ...settings, description: e.target.value })}/>
      <label>LINE OA URL</label><input type="url" value={settings.line_oa_url || ''} onChange={e => setSettings({ ...settings, line_oa_url: e.target.value })}/>
      <label>ค่าจัดส่ง</label><input type="number" min="0" step="0.01" value={settings.delivery_fee_thb} onChange={e => setSettings({ ...settings, delivery_fee_thb: Number(e.target.value) })}/>
      <label>ยอดส่งฟรี (0 = ปิดส่งฟรี)</label><input type="number" min="0" step="0.01" value={settings.free_delivery_threshold_thb || 0} onChange={e => setSettings({ ...settings, free_delivery_threshold_thb: Number(e.target.value) })}/>
      <label className="checkbox"><input type="checkbox" checked={settings.accepting_orders} onChange={e => setSettings({ ...settings, accepting_orders: e.target.checked })}/>เปิดรับออเดอร์</label>
      <button className="button primary" disabled={busy}>บันทึก</button>
    </form>}
  </main>;
}
function OrderTable({ orders, onStatus, busy }: { orders: Order[]; onStatus: (order: Order, status: string) => void; busy: boolean }) {
  const transitions: Record<string, string[]> = { pending_payment: ['cancelled'], pending_verification: ['cancelled'], confirmed: ['preparing'], preparing: ['ready_for_pickup', 'shipped'], ready_for_pickup: ['completed'], shipped: ['completed'] };
  return <table className="table"><thead><tr><th>ออเดอร์</th><th>ลูกค้า</th><th>ยอดรวม</th><th>สถานะ</th><th>เปลี่ยนสถานะ</th></tr></thead><tbody>{orders.map(o => <tr key={o.id}><td>{o.order_number}</td><td>{o.customer_name_snapshot}</td><td>฿{o.total_thb.toLocaleString('th-TH')}</td><td>{o.order_status}</td><td><select disabled={busy || !transitions[o.order_status]?.length} value="" onChange={e => { if (e.target.value) onStatus(o, e.target.value); }}><option value="">เลือกสถานะถัดไป</option>{(transitions[o.order_status] || []).map(s => <option key={s} value={s}>{s}</option>)}</select></td></tr>)}</tbody></table>;
}
