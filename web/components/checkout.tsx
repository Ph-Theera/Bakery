'use client';
import { useEffect, useState } from 'react';
import { api } from '../lib/api';
import type { CartItem, Order } from '../lib/types';
import PaymentSlip from './payment-slip';

export default function Checkout() {
  const [cart, setCart] = useState<CartItem[]>([]);
  const [step, setStep] = useState(1);
  const [method, setMethod] = useState('bank_transfer');
  const [name, setName] = useState(''); const [phone, setPhone] = useState(''); const [address, setAddress] = useState('');
  const [quote, setQuote] = useState<{ total_thb: number; delivery_fee_thb: number } | null>(null);
  const [order, setOrder] = useState<Order | null>(null);
  const [error, setError] = useState(''); const [busy, setBusy] = useState(false);
  useEffect(() => {
    try { setCart(JSON.parse(window.localStorage.getItem('olive-cart') || '[]')); }
    catch { setError('ข้อมูลตะกร้าเสียหาย กรุณาเลือกสินค้าใหม่'); }
  }, []);
  const items = () => cart.map(item => ({ productId: item.product.id, quantity: item.quantity }));
  const getQuote = async () => {
    setBusy(true); setError('');
    try { setQuote(await api.quote({ items: items(), fulfillment: { type: 'delivery', address } })); setStep(3); }
    catch (e) { setError(e instanceof Error ? e.message : 'คำนวณราคาไม่สำเร็จ'); }
    finally { setBusy(false); }
  };
  const submit = async () => {
    setBusy(true); setError('');
    try {
      const payload = { customer: { name: name.trim(), phone: phone.trim() }, fulfillment: { type: 'delivery', address: address.trim() }, items: items(), paymentMethod: method };
      const serialized = JSON.stringify(payload);
      const saved = JSON.parse(window.sessionStorage.getItem('tongbai-checkout-request') || 'null');
      const key = saved?.payload === serialized ? saved.key : crypto.randomUUID();
      window.sessionStorage.setItem('tongbai-checkout-request', JSON.stringify({ payload: serialized, key }));
      const result = await api.createOrder(payload, key);
      setOrder(result);
      window.localStorage.removeItem('olive-cart');
      window.sessionStorage.removeItem('tongbai-checkout-request');
    } catch (e) { setError(e instanceof Error ? e.message : 'สร้างออเดอร์ไม่สำเร็จ'); }
    finally { setBusy(false); }
  };
  if (order) return <main className="page container" style={{ padding: '60px 0' }}>
    <div className="panel"><div className="eyebrow">ORDER RECEIVED</div><h1>สร้างคำสั่งซื้อแล้ว</h1><h2>{order.order_number}</h2>
      <p>ยอดรวม ฿{order.total_thb.toLocaleString('th-TH')} · {order.payment_status === 'submitted' ? 'รอร้านตรวจสอบสลิป' : 'รอชำระเงิน'}</p>
      <a className="button secondary" href="/orders">ดูคำสั่งซื้อของฉัน</a>
    </div><PaymentSlip order={order} onUpdated={setOrder}/>
    <p className="muted">เก็บสิทธิ์ติดตามออเดอร์ในแท็บนี้ หากปิดแท็บจะต้องใช้ข้อมูลเข้าถึงอีกครั้ง</p>
  </main>;
  return <main className="page container" style={{ padding: '50px 0' }}>
    <a className="brand" href="/">TongBai</a><h1>สั่งซื้อ</h1><p className="muted">ขั้นตอน {step} / 4</p>
    {!cart.length ? <div className="empty">ตะกร้าว่าง <a href="/">กลับไปเลือกขนม</a></div> :
    <div className="checkout-layout"><section className="panel">
      {step === 1 && <><h2>ข้อมูลลูกค้า</h2><label>ชื่อผู้รับ</label><input value={name} onChange={e => setName(e.target.value)}/><label>เบอร์โทรศัพท์</label><input type="tel" value={phone} onChange={e => setPhone(e.target.value)}/><button className="button primary" onClick={() => name.trim() && phone.trim() ? (setError(''), setStep(2)) : setError('กรุณากรอกชื่อและเบอร์โทรศัพท์')}>ถัดไป</button></>}
      {step === 2 && <><h2>ข้อมูลจัดส่ง</h2><label>ที่อยู่จัดส่ง</label><textarea value={address} onChange={e => setAddress(e.target.value)} rows={5}/><button className="button primary" disabled={busy} onClick={() => address.trim() ? getQuote() : setError('กรุณากรอกที่อยู่')}>คำนวณยอดรวม</button></>}
      {step === 3 && <><h2>ตรวจสอบรายการ</h2><p>ค่าจัดส่ง ฿{quote?.delivery_fee_thb.toLocaleString('th-TH')}</p><div className="summary-total">รวม ฿{quote?.total_thb.toLocaleString('th-TH')}</div><button className="button primary" onClick={() => setStep(4)}>เลือกวิธีชำระเงิน</button></>}
      {step === 4 && <><h2>วิธีชำระเงิน</h2><div className="payment-choice"><button className={method === 'bank_transfer' ? 'selected' : ''} onClick={() => setMethod('bank_transfer')}>โอนผ่านธนาคาร</button><button className={method === 'promptpay' ? 'selected' : ''} onClick={() => setMethod('promptpay')}>PromptPay</button></div><p>สร้างออเดอร์ก่อน แล้วอัปโหลดสลิปในขั้นตอนถัดไป กรุณาขอข้อมูลบัญชีหรือ QR ที่ถูกต้องจากร้าน</p><button className="button primary" disabled={busy} onClick={submit}>{busy ? 'กำลังส่ง...' : 'ยืนยันคำสั่งซื้อ'}</button></>}
      {step > 1 && <button className="button secondary" disabled={busy} onClick={() => { setError(''); setStep(step - 1); }}>ย้อนกลับ</button>}
      {error && <p className="error" role="alert">{error}</p>}
    </section><aside className="panel"><h2>สรุปคำสั่งซื้อ</h2>{cart.map(item => <div key={item.product.id} className="summary-row"><span>{item.product.name} × {item.quantity}</span><b>฿{(item.product.price_thb * item.quantity).toLocaleString('th-TH')}</b></div>)}{quote && <div className="summary-total">รวม ฿{quote.total_thb.toLocaleString('th-TH')}</div>}</aside></div>}
  </main>;
}
