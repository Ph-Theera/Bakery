(function () {
  const $ = (s) => document.querySelector(s);
  const read = (key, fallback) => JSON.parse(localStorage.getItem(key) || JSON.stringify(fallback));
  const write = (key, value) => localStorage.setItem(key, JSON.stringify(value));
  const money = (n) => '฿' + Number(n || 0).toLocaleString('th-TH');
  const notify = (m) => { const t = $('#toast'); if (!t) return; t.textContent = m; t.classList.add('show'); setTimeout(() => t.classList.remove('show'), 2200); };

  const seedFromDom = () => Array.from(document.querySelectorAll('.product-card')).map((card, index) => ({
    id: card.querySelector('[data-add]')?.dataset.add || 'p' + (index + 1), name: card.querySelector('h3')?.textContent || 'New Product', category: 'คุกกี้', price: Number((card.querySelector('.price')?.textContent || '0').replace(/[^0-9]/g, '')), stock: 10, desc: card.querySelector('.product-info p')?.textContent || '', img: card.querySelector('img')?.src || '', featured: false
  }));
  const products = () => read('oo_products', seedFromDom());

  function modal(title, product = {}) {
    const overlay = document.createElement('div'); overlay.className = 'product-detail-backdrop';
    overlay.innerHTML = `<article class="admin-form-modal"><button class="product-detail-close">×</button><p class="eyebrow">PRODUCT MANAGEMENT</p><h2>${title}</h2><label>ชื่อสินค้า</label><input id="adminProductName" value="${product.name || ''}" placeholder="ชื่อสินค้า"><label>คำอธิบาย</label><textarea id="adminProductDesc" rows="3">${product.desc || ''}</textarea><div class="admin-form-two"><div><label>ราคา (บาท)</label><input id="adminProductPrice" type="number" value="${product.price || ''}"></div><div><label>สต็อก</label><input id="adminProductStock" type="number" value="${product.stock || 0}"></div></div><label>URL รูปสินค้า</label><input id="adminProductImg" value="${product.img || ''}" placeholder="https://..."><div class="checkout-actions"><button class="btn btn-outline" id="cancelAdminForm">ยกเลิก</button><button class="btn btn-primary" id="saveAdminProduct">บันทึกสินค้า</button></div></article>`;
    document.body.appendChild(overlay); const close = () => overlay.remove(); overlay.querySelector('.product-detail-close').onclick = close; $('#cancelAdminForm').onclick = close;
    $('#saveAdminProduct').onclick = () => { const list = products(); const item = { id: product.id || 'p' + Date.now(), name: $('#adminProductName').value.trim() || 'New Product', category: product.category || 'คุกกี้', price: Number($('#adminProductPrice').value || 0), stock: Number($('#adminProductStock').value || 0), desc: $('#adminProductDesc').value.trim(), img: $('#adminProductImg').value.trim() || 'https://images.unsplash.com/photo-1499636136210-6f4ee915583e?auto=format&fit=crop&w=700&q=80', featured: product.featured || false }; const next = product.id ? list.map(x => x.id === product.id ? item : x) : [...list, item]; write('oo_products', next); close(); notify('บันทึกสินค้าแล้ว'); setTimeout(() => location.reload(), 350); };
  }

  function renderCustomers() {
    const orders = read('oo_orders', []); const customers = [...new Map(orders.map(o => [o.customer, o])).values()];
    $('#adminTitle').textContent = 'ลูกค้า'; $('#adminContent').innerHTML = `<div class="panel"><div class="admin-section-head"><h2>รายชื่อลูกค้า</h2><span class="muted">ข้อมูลจากออเดอร์เดโม</span></div><table class="admin-table"><thead><tr><th>ลูกค้า</th><th>ออเดอร์ล่าสุด</th><th>ยอดล่าสุด</th><th>สถานะ</th></tr></thead><tbody>${customers.map(c => `<tr><td><b>${c.customer || 'Guest'}</b><br><span class="muted">ลูกค้าเดโม</span></td><td>${c.id}</td><td>${money(c.total)}</td><td><span class="status-pill">ใช้งาน</span></td></tr>`).join('') || '<tr><td colspan="4">ยังไม่มีข้อมูลลูกค้า</td></tr>'}</tbody></table></div>`;
  }

  const addCustomerNav = () => { const nav = document.querySelector('.admin-sidebar nav'); if (nav && !nav.querySelector('[data-admin="customers"]')) nav.insertAdjacentHTML('beforeend', '<a data-admin="customers">♙ <span>ลูกค้า</span></a>'); };
  addCustomerNav();
  document.addEventListener('click', (event) => {
    const customer = event.target.closest('[data-admin="customers"]'); if (customer) { event.preventDefault(); event.stopImmediatePropagation(); document.querySelectorAll('[data-admin]').forEach(x => x.classList.remove('active')); customer.classList.add('active'); renderCustomers(); return; }
    const add = event.target.closest('#addProduct'); if (add) { event.preventDefault(); event.stopImmediatePropagation(); modal('เพิ่มสินค้าใหม่'); return; }
    const row = event.target.closest('.admin-product-row');
    if (row && event.target.closest('[data-edit]')) { event.preventDefault(); event.stopImmediatePropagation(); const id = event.target.closest('[data-edit]').dataset.edit; modal('แก้ไขสินค้า', products().find(p => p.id === id) || {}); return; }
    if (row && event.target.closest('[data-delete]')) { event.preventDefault(); event.stopImmediatePropagation(); const id = event.target.closest('[data-delete]').dataset.delete; if (confirm('ลบสินค้าออกจากเดโมหรือไม่?')) { write('oo_products', products().filter(p => p.id !== id)); location.reload(); } }
  }, true);

  const observer = new MutationObserver(() => {
    document.querySelectorAll('.admin-product-row').forEach((row, index) => { const actions = row.querySelector('.row-actions'); if (!actions || actions.querySelector('[data-edit]')) return; const item = products()[index]; if (!item) return; actions.insertAdjacentHTML('afterbegin', `<button data-edit="${item.id}">แก้ไข</button><button data-delete="${item.id}">ลบ</button>`); });
    addCustomerNav();
  });
  observer.observe($('#adminContent') || document.body, { childList: true, subtree: true });
})();
