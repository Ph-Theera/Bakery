'use client';
import { useEffect, useState } from 'react';
import { api } from '../../lib/api';
import type { Order } from '../../lib/types';
import PaymentSlip from '../../components/payment-slip';

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [busy, setBusy] = useState(true); const [error, setError] = useState('');
  const load = async () => {
    setBusy(true); setError('');
    try { setOrders(await api.orders()); }
    catch (e) { setError(e instanceof Error ? e.message : 'โหลดคำสั่งซื้อไม่สำเร็จ'); }
    finally { setBusy(false); }
  };
  useEffect(() => { void load(); }, []);
  return <main className="page container" style={{ padding: '50px 0' }}>
    <a className="brand" href="/">TongBai</a><h1>คำสั่งซื้อของฉัน</h1>
    <button className="button secondary" disabled={busy} onClick={load}>รีเฟรชสถานะ</button>
    {busy && <p>กำลังโหลด...</p>}{error && <p className="error" role="alert">{error}</p>}
    {!busy && !error && !orders.length && <div className="empty">ยังไม่มีคำสั่งซื้อในแท็บนี้ <a href="/">เลือกขนม</a></div>}
    {orders.map(order => <section className="panel" key={order.id}>
      <h2>{order.order_number}</h2><p>{new Date(order.created_at).toLocaleDateString('th-TH')} · ฿{order.total_thb.toLocaleString('th-TH')}</p>
      <p>ออเดอร์: {order.order_status} · การชำระเงิน: {order.payment_status}</p>
      {order.items?.map((item, i) => <p key={i}>{item.product_name_snapshot} × {item.quantity}</p>)}
      <PaymentSlip order={order} onUpdated={updated => setOrders(current => current.map(o => o.id === updated.id ? updated : o))}/>
    </section>)}
  </main>;
}
