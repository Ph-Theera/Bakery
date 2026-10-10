// Extra interactions kept separate so the static demo remains easy to replace with React later.
const searchControl = document.querySelector('#searchBtn');
if (searchControl) searchControl.onclick = () => { location.hash = '#products'; setTimeout(() => document.querySelector('#productSearch')?.focus(), 50); };
const menuControl = document.querySelector('#menuBtn');
if (menuControl) menuControl.onclick = () => document.querySelector('.main-nav')?.classList.toggle('open');
document.querySelectorAll('.radio-card').forEach(card => card.addEventListener('click', () => {
  card.parentElement.querySelectorAll('.radio-card').forEach(item => item.classList.remove('selected'));
  card.classList.add('selected');
}));
const hint = localStorage.getItem('oo_demo_hint');
if (hint) {
  localStorage.removeItem('oo_demo_hint');
  if (hint === 'cart') setTimeout(() => document.querySelector('#cartBtn')?.click(), 250);
  if (hint === 'checkout') setTimeout(() => document.querySelector('#checkoutBtn')?.click(), 300);
}
