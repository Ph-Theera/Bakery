// Product detail interaction for the static storefront demo.
(function () {
  const grid = document.querySelector('#productGrid');
  if (!grid) return;
  const openDetail = (card) => {
    const image = card.querySelector('img');
    const title = card.querySelector('h3')?.textContent || '';
    const description = card.querySelector('.product-info p')?.textContent || '';
    const price = card.querySelector('.price')?.textContent || '';
    const id = card.querySelector('[data-add]')?.dataset.add || '';
    const overlay = document.createElement('div');
    overlay.className = 'product-detail-backdrop';
    overlay.innerHTML = `<article class="product-detail-modal" role="dialog" aria-modal="true" aria-label="รายละเอียดสินค้า"><button class="product-detail-close" aria-label="ปิด">×</button><div class="product-detail-photo"><img src="${image?.src || ''}" alt="${title}"></div><div class="product-detail-copy"><p class="eyebrow">HANDMADE TODAY</p><h2>${title}</h2><p>${description}</p><strong>${price}</strong><p class="muted">อบสดใหม่ตามออเดอร์ · แนะนำให้รับประทานภายใน 3 วัน</p><button class="btn btn-primary btn-full" data-add="${id}">เพิ่มลงตะกร้า <span>→</span></button></div></article>`;
    document.body.appendChild(overlay);
    const close = () => overlay.remove();
    overlay.querySelector('.product-detail-close').onclick = close;
    overlay.addEventListener('click', (event) => { if (event.target === overlay) close(); });
    overlay.querySelector('[data-add]')?.addEventListener('click', () => setTimeout(close, 100));
  };
  grid.addEventListener('click', (event) => {
    if (event.target.closest('button, a')) return;
    const card = event.target.closest('.product-card');
    if (card) openDetail(card);
  });
})();
