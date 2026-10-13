'use client';
import { useState } from 'react';
import { api } from '../lib/api';
import type { Order } from '../lib/types';

export default function PaymentSlip({ order, onUpdated }: { order: Order; onUpdated: (order: Order) => void }) {
  const [file, setFile] = useState<File | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  if (!order.payment_id || !['pending', 'rejected', 'needs_resubmission'].includes(order.payment_status || '') ||
      !['pending_payment', 'pending_verification'].includes(order.order_status)) return null;
  const submit = async () => {
    if (!file || !order.payment_id) return;
    setBusy(true); setError('');
    try {
      await api.uploadPaymentSlip(order.payment_id, file, order.id);
      await api.submitPayment(order.payment_id, order.id);
      onUpdated(await api.order(order.id));
    } catch (e) { setError(e instanceof Error ? e.message : 'ส่งสลิปไม่สำเร็จ'); }
    finally { setBusy(false); }
  };
  return <section className="panel">
    <h3>แจ้งชำระเงิน</h3>
    <p>กรุณาตรวจสอบบัญชีหรือ QR ของร้านก่อนโอน เว็บยังไม่มีข้อมูลบัญชี/QR จาก backend</p>
    <p>ยอดชำระ ฿{order.total_thb.toLocaleString('th-TH')} · ออเดอร์ {order.order_number}</p>
    <input aria-label="สลิปโอนเงิน" type="file" accept="image/jpeg,image/png,image/webp" disabled={busy} onChange={e => { setFile(e.target.files?.[0] || null); setError(''); }} />
    <p className="muted">JPG, PNG, WebP ไม่เกิน 5MB · ร้านตรวจสอบก่อนยืนยันการชำระเงิน</p>
    <button className="button primary" disabled={busy || !file} onClick={submit}>{busy ? 'กำลังส่ง...' : 'ส่งสลิปให้ร้านตรวจสอบ'}</button>
    {error && <p className="error" role="alert">{error}</p>}
  </section>;
}
