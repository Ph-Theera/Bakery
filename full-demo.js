/* Complete browser-only demo flows. Backend/database integration can replace these services later. */
(function () {
  const $ = (selector) => document.querySelector(selector);
  const money = (value) => '฿' + Number(value || 0).toLocaleString('th-TH');
  const read = (key, fallback) => JSON.parse(localStorage.getItem(key) || JSON.stringify(fallback));
  const write = (key, value) => localStorage.setItem(key, JSON.stringify(value));
  const getCart = () => read('oo_cart', []);
  const getOrders = () => read('oo_orders', []);

  function notify(message) {
    const toast = $('#toast');
    if (!toast) return;
    toast.textContent = message; toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 2200);
  }

  function cartTotal() {
    const cart = getCart();
    const subtotal = cart.reduce((sum, item) => sum + item.price * item.qty, 0);
    return { cart, subtotal, shipping: subtotal >= 500 || subtotal === 0 ? 0 : 45, total: subtotal + (subtotal >= 500 || subtotal === 0 ? 0 : 45) };
  }

  function checkout() {
    const totals = cartTotal();
    if (!totals.cart.length) { notify('กรุณาเลือกขนมก่อนสั่งซื้อ'); return; }
    const backdrop = $('#modalBackdrop');
    const modal = $('#checkoutModal');
    let step = 1; let delivery = 'delivery'; let payment = 'bank_transfer'; let slip = null;
    const data = { name: '', phone: '', line: '', address: '', note: '', pickupDate: '', pickupTime: '' };
    backdrop.classList.remove('hidden');

    const render = () => {
      const stepper = `<div class="stepper">${[1, 2, 3, 4].map(n => `<span class="${n <= step ? 'active' : ''}"></span>`).join('')}</div>`;
      let body = '';
      if (step === 1) body = `<h2>ข้อมูลลูกค้า</h2><p class="muted">ข้อมูลสำหรับติดต่อและยืนยันคำสั่งซื้อ</p><label>ชื่อผู้รับ</label><input id="fullName" value="${data.name}" placeholder="เช่น คุณมินตรา"><label>เบอร์โทรศัพท์</label><input id="phone" value="${data.phone}" placeholder="08x-xxx-xxxx"><label>LINE display name <span class="muted">(ไม่บังคับ)</span></label><input id="lineName" value="${data.line}" placeholder="ชื่อที่ใช้ใน LINE">`;
      if (step === 2) body = `<h2>วิธีรับสินค้า</h2><div class="radio-row"><div class="radio-card ${delivery === 'delivery' ? 'selected' : ''}" data-delivery="delivery">จัดส่งถึงบ้าน<br><small class="muted">${totals.shipping ? money(totals.shipping) : 'ฟรี'}</small></div><div class="radio-card ${delivery === 'pickup' ? 'selected' : ''}" data-delivery="pickup">รับที่ร้าน<br><small class="muted">ฟรี</small></div></div>${delivery === 'delivery' ? '<label>ที่อยู่จัดส่ง</label><textarea id="address" rows="3" placeholder="บ้านเลขที่ ถนน แขวง เขต จังหวัด">'+data.address+'</textarea><label>หมายเหตุการจัดส่ง</label><input id="note" value="'+data.note+'" placeholder="เช่น ฝากวางหน้าบ้าน">' : '<label>วันที่รับสินค้า</label><input id="pickupDate" type="date" value="'+data.pickupDate+'"><label>เวลารับสินค้า</label><select id="pickupTime"><option>10:00–12:00</option><option>13:00–15:00</option><option>16:00–18:00</option></select>'}`;
      if (step === 3) body = `<h2>ตรวจสอบคำสั่งซื้อ</h2><div class="review-list">${totals.cart.map(item => `<div><span>${item.name} × ${item.qty}</span><b>${money(item.price * item.qty)}</b></div>`).join('')}</div><div class="review-total"><span>ยอดสินค้า</span><b>${money(totals.subtotal)}</b><span>ค่าจัดส่ง</span><b>${totals.shipping ? money(totals.shipping) : 'ฟรี'}</b><strong>ยอดรวมทั้งหมด</strong><strong>${money(totals.total)}</strong></div><div class="review-customer"><b>${data.name || 'ลูกค้า Olive & Oven'}</b><span>${delivery === 'delivery' ? data.address || 'ยังไม่ได้กรอกที่อยู่' : 'รับที่ร้าน · '+(data.pickupDate || 'เลือกวันที่')}</span></div>`;
      if (step === 4) body = `<h2>เลือกวิธีชำระเงิน</h2><div class="radio-row"><div class="radio-card ${payment === 'bank_transfer' ? 'selected' : ''}" data-payment="bank_transfer">โอนผ่านธนาคาร<br><small class="muted">ส่งสลิปผ่าน LINE</small></div><div class="radio-card ${payment === 'promptpay' ? 'selected' : ''}" data-payment="promptpay">PromptPay QR<br><small class="muted">อัปโหลดสลิป</small></div></div>${payment === 'bank_transfer' ? '<div class="payment-box"><b>ธนาคารกสิกรไทย</b><span>ชื่อบัญชี Olive & Oven</span><strong>123-4-56789-0</strong><span>ยอดที่ต้องชำระ: <b>'+money(totals.total)+'</b></span><button class="btn btn-outline btn-small" id="copyPayment">คัดลอกเลขบัญชี</button></div>' : '<div class="payment-box promptpay"><div class="qr-placeholder">QR<br><small>DEMO</small></div><b>PromptPay Demo QR</b><span>ยอดที่ต้องชำระ: <b>'+money(totals.total)+'</b></span><label class="upload-slip">เลือกไฟล์สลิป<input id="slipInput" type="file" accept="image/jpeg,image/png,image/webp"></label>'+(slip ? '<small class="file-ok">✓ '+slip.name+'</small>' : '<small class="muted">รองรับ JPG, PNG, WebP ไม่เกิน 5MB</small>')+'</div>'}`;
      modal.innerHTML = `${stepper}${body}<div class="checkout-actions"><button class="btn btn-outline" id="checkoutBack">${step === 1 ? 'ยกเลิก' : 'ย้อนกลับ'}</button><button class="btn btn-primary" id="checkoutNext">${step === 4 ? 'ยืนยันคำสั่งซื้อ · '+money(totals.total) : 'ถัดไป →'}</button></div>`;
      modal.querySelectorAll('[data-delivery]').forEach(el => el.onclick = () => { delivery = el.dataset.delivery; render(); });
      modal.querySelectorAll('[data-payment]').forEach(el => el.onclick = () => { payment = el.dataset.payment; render(); });
      $('#checkoutBack').onclick = () => { if (step === 1) backdrop.classList.add('hidden'); else { step--; render(); } };
      $('#checkoutNext').onclick = () => { if (step === 1) { data.name = $('#fullName').value.trim(); data.phone = $('#phone').value.trim(); data.line = $('#lineName').value.trim(); if (!data.name || !data.phone) return notify('กรุณากรอกชื่อและเบอร์โทรศัพท์'); } if (step === 2) { data.address = $('#address')?.value.trim() || ''; data.note = $('#note')?.value.trim() || ''; data.pickupDate = $('#pickupDate')?.value || ''; if (delivery === 'delivery' && !data.address) return notify('กรุณากรอกที่อยู่จัดส่ง'); } if (step < 4) { step++; render(); } else placeOrder(); };
      $('#copyPayment')?.addEventListener('click', async () => { await navigator.clipboard?.writeText('123-4-56789-0'); notify('คัดลอกเลขบัญชีแล้ว'); });
      $('#slipInput')?.addEventListener('change', (event) => { const file = event.target.files?.[0]; if (!file) return; if (!['image/jpeg','image/png','image/webp'].includes(file.type) || file.size > 5 * 1024 * 1024) { notify('ไฟล์ต้องเป็น JPG, PNG, WebP และไม่เกิน 5MB'); event.target.value = ''; return; } slip = file; render(); });
    };
    const placeOrder = () => {
      const id = 'OO-' + String(Date.now()).slice(-8);
      const orders = getOrders();
      orders.unshift({ id, customer: data.name, date: new Date().toLocaleDateString('th-TH'), total: totals.total, status: payment === 'promptpay' ? 'รอตรวจสอบการชำระเงิน' : 'รอชำระเงิน', payment: payment === 'promptpay' ? 'PromptPay' : 'โอนผ่านธนาคาร', items: totals.cart.map(item => ({ name: item.name, qty: item.qty, price: item.price })) });
      write('oo_orders', orders); write('oo_cart', []); backdrop.classList.add('hidden'); renderConfirmation(orders[0]);
    };
    render();
  }

  function renderConfirmation(order) {
    const backdrop = $('#modalBackdrop'); const modal = $('#checkoutModal'); backdrop.classList.remove('hidden');
    modal.innerHTML = `<div class="checkout-success"><p class="eyebrow">ORDER RECEIVED</p><h2>ขอบคุณสำหรับคำสั่งซื้อ</h2><p class="muted">เราได้รับออเดอร์ของคุณแล้ว กรุณาชำระเงินตามวิธีที่เลือก</p><div class="order-id">${order.id}</div><div class="confirmation-summary"><span>ผู้สั่งซื้อ</span><b>${order.customer}</b><span>ยอดรวม</span><b>${money(order.total)}</b><span>สถานะการชำระเงิน</span><b>${order.status}</b></div><div class="timeline"><div class="done"><i>✓</i><span>รับคำสั่งซื้อ</span></div><div><i>2</i><span>ตรวจสอบการชำระเงิน</span></div><div><i>3</i><span>เตรียมขนม</span></div><div><i>4</i><span>จัดส่ง / รับที่ร้าน</span></div></div><div class="checkout-actions"><button class="btn btn-outline" id="copyOrder">คัดลอกสรุปออเดอร์</button><button class="btn btn-primary" id="closeConfirmation">ดูคำสั่งซื้อ</button></div></div>`;
    $('#copyOrder').onclick = async () => { const summary = `Olive & Oven\nOrder: ${order.id}\nยอดรวม: ${money(order.total)}\nผู้สั่งซื้อ: ${order.customer}`; await navigator.clipboard?.writeText(summary); notify('คัดลอกสรุปออเดอร์แล้ว'); };
    $('#closeConfirmation').onclick = () => { backdrop.classList.add('hidden'); location.hash = '#orders'; if (typeof renderOrders === 'function') renderOrders(); };
  }

  document.addEventListener('click', (event) => {
    if (event.target.closest('#checkoutBtn')) { event.preventDefault(); event.stopImmediatePropagation(); checkout(); }
  }, true);
})();
