'use client';

import { useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { api, imageUrl } from '../lib/api';
import type { Category, CartItem, Product, StoreSettings } from '../lib/types';

type Props = {
  initialProducts: Product[];
  initialCategories: Category[];
  initialStore: StoreSettings | null;
};

const fallbackImages = [
  'https://images.unsplash.com/photo-1555507036-ab1f4038808a?auto=format&fit=crop&w=1000&q=85',
  'https://images.unsplash.com/photo-1499636136210-6f4ee915583e?auto=format&fit=crop&w=1000&q=85',
  'https://images.unsplash.com/photo-1464195244916-405fa0dc7b3d?auto=format&fit=crop&w=1000&q=85',
];

const imageFor = (product: Product | undefined, fallback: string) => imageUrl(product?.images?.[0]?.public_url || fallback);
const money = (amount: number) => new Intl.NumberFormat('th-TH', { style: 'currency', currency: 'THB', maximumFractionDigits: 0 }).format(amount);

export default function Storefront({ initialProducts, initialCategories, initialStore }: Props) {
  const [products, setProducts] = useState(initialProducts);
  const [category, setCategory] = useState('');
  const [query, setQuery] = useState('');
  const [cart, setCart] = useState<CartItem[]>([]);
  const [cartLoaded, setCartLoaded] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    try {
      const saved = JSON.parse(window.localStorage.getItem('olive-cart') || '[]');
      if (Array.isArray(saved)) setCart(saved);
    } catch {
      setError('ข้อมูลตะกร้าเดิมเสียหาย กรุณาเลือกสินค้าใหม่');
    } finally {
      setCartLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (cartLoaded) window.localStorage.setItem('olive-cart', JSON.stringify(cart));
  }, [cart, cartLoaded]);

  const filtered = useMemo(() => products.filter(product =>
    (!category || product.category_id === category) &&
    (!query || product.name.toLowerCase().includes(query.toLowerCase()))
  ), [products, category, query]);
  const featured = initialProducts.filter(product => product.is_featured).slice(0, 3);
  const heroProducts = initialProducts.slice(0, 3);
  const spotlight = featured[0] || initialProducts[0];
  const count = cart.reduce((sum, item) => sum + item.quantity, 0);

  const add = (product: Product) => setCart(items => {
    const found = items.find(item => item.product.id === product.id);
    return found
      ? items.map(item => item.product.id === product.id
        ? { ...item, quantity: Math.min(item.quantity + 1, product.stock_quantity) }
        : item)
      : [...items, { product, quantity: 1 }];
  });

  const refresh = async (next: string) => {
    setCategory(next);
    setError('');
    try {
      setProducts(await api.products(next ? `category=${encodeURIComponent(next)}` : ''));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'โหลดสินค้าไม่สำเร็จ');
    }
  };

  return <>
    <header className="site-header">
      <nav className="nav nav-left" aria-label="เมนูหลัก">
        <a href="#home">HOME</a><a href="#products">MENU</a><a href="#about">OUR STORY</a><a href="#gallery">GALLERY</a>
      </nav>
      <Link className="brand" href="/" aria-label="TongBai หน้าหลัก"><span className="brand-small">the</span><span>TongBai</span><span className="brand-small brand-last">bakery ✦</span></Link>
      <nav className="nav nav-right" aria-label="เมนูร้านค้า">
        <Link href="/orders">MY ORDERS</Link><Link href="/admin">ADMIN</Link><Link className="nav-order" href="/checkout">ORDER NOW <span aria-hidden="true">↗</span></Link>
      </nav>
      <Link className="mobile-order" href="/checkout">ตะกร้า {count} ↗</Link>
    </header>

    <main>
      <section className="hero" id="home">
        <div className="hero-inner container">
          <span className="eyebrow">HANDCRAFTED WITH CARE · TONGBAI BAKERY</span>
          <h1>Freshly Baked <em>Joy</em></h1>
          <p className="hero-subtitle">อบสดใหม่ทุกวัน เพื่อช่วงเวลาที่พิเศษในทุกคำ</p>
          <div className="hero-gallery" aria-label="ขนมจากร้าน TongBai">
            {[0, 1, 2].map(index => <div className={`hero-photo hero-photo-${index + 1}`} key={index}>
              <img src={imageFor(heroProducts[index], fallbackImages[index])} alt={heroProducts[index]?.name || 'ขนมอบจาก TongBai'} />
            </div>)}
          </div>
          <p className="hero-bottom">{initialStore?.description || 'ขนมโฮมเมดอบสดใหม่ด้วยความใส่ใจ'}</p>
        </div>
        <div className="category-ribbon" aria-label="หมวดหมู่สินค้า">
          <a href="#products">ทั้งหมด</a>
          {initialCategories.map(item => <a href="#products" key={item.id} onClick={() => void refresh(item.id)}>{item.name_th}</a>)}
          <a href="#products">ดูเมนู ↗</a>
        </div>
      </section>

      <section className="spotlight dark-section">
        <div className="container spotlight-grid">
          <div className="spotlight-visual">
            <span className="visual-orbit" aria-hidden="true" />
            <img src={imageFor(spotlight, fallbackImages[0])} alt={spotlight?.name || 'ขนมอบซิกเนเจอร์'} />
          </div>
          <div className="spotlight-copy">
            <span className="eyebrow">OUR SIGNATURE PICK</span>
            <h2>{spotlight?.name || 'Fresh From The Oven'}</h2>
            <p>{spotlight?.description_th || 'ขนมอบหอมกรุ่นจากครัวของเรา ทำด้วยความใส่ใจในทุกขั้นตอน'}</p>
            {spotlight && <strong className="spotlight-price">{money(spotlight.price_thb)}</strong>}
            {spotlight && <button className="gold-button" disabled={!spotlight.stock_quantity} onClick={() => add(spotlight)}>{spotlight.stock_quantity ? 'เพิ่มลงตะกร้า' : 'สินค้าหมด'} <span aria-hidden="true">+</span></button>}
            {!spotlight && <a className="gold-button" href="#products">ดูเมนู <span aria-hidden="true">↗</span></a>}
          </div>
        </div>
      </section>

      <section className="story section-cream" id="about">
        <div className="container">
          <span className="eyebrow">THE STORY BEHIND EVERY BITE</span>
          <h2>More Than<br /><em>Just Bakery</em></h2>
          <div className="story-grid">
            <img src="https://images.unsplash.com/photo-1608198093002-ad4e005484df?auto=format&fit=crop&w=1200&q=85" alt="ขนมปังอบสดในร้านเบเกอรี่" />
            <div className="story-copy">
              <p>{initialStore?.description || 'TongBai เชื่อว่าขนมที่ดีเริ่มจากความใส่ใจ เราอบขนมสดใหม่และเลือกสิ่งที่ดีที่สุดให้ทุกวันของคุณ'}</p>
              <ul><li>อบสดใหม่ด้วยความตั้งใจ</li><li>เลือกวัตถุดิบอย่างพิถีพิถัน</li><li>รสชาติที่อยากแบ่งปัน</li><li>ความสุขเล็ก ๆ ในทุกคำ</li></ul>
              <a className="text-link" href="#products">เลือกขนมที่ชอบ <span aria-hidden="true">↗</span></a>
            </div>
          </div>
          <div className="story-signature" aria-hidden="true">Made with love, baked for you.</div>
        </div>
      </section>

      {featured.length > 0 && <section className="popular section-cream">
        <div className="container popular-inner">
          <span className="eyebrow">CUSTOMER FAVORITES</span>
          <h2>Popular <em>Flavors</em></h2>
          <div className="popular-grid">
            {featured.map(product => <article className="popular-card" key={product.id}>
              <span className="popular-stars" aria-label="สินค้าแนะนำ">✦ ✦ ✦</span>
              <Link href={`/products/${product.slug}`}><h3>{product.name}</h3></Link>
              <p>{product.description_th}</p>
              <strong>{money(product.price_thb)}</strong>
              <Link className="popular-image" href={`/products/${product.slug}`}><img src={imageFor(product, fallbackImages[0])} alt={product.name} /></Link>
              <button className="round-add" disabled={!product.stock_quantity} onClick={() => add(product)} aria-label={`เพิ่ม ${product.name} ลงตะกร้า`}>{product.stock_quantity ? '+' : '×'}</button>
            </article>)}
          </div>
        </div>
      </section>}

      <section className="values section-cream">
        <div className="container values-grid">
          <div className="value-list"><div><span>✦</span><h3>Handmade Daily</h3><p>ขนมทุกชิ้นทำด้วยมือและความใส่ใจ</p></div><div><span>♡</span><h3>Thoughtfully Selected</h3><p>เลือกวัตถุดิบให้ได้รสชาติที่เราภูมิใจ</p></div></div>
          <div className="values-image"><img src={fallbackImages[1]} alt="ขนมอบจาก TongBai" /></div>
          <div className="value-list"><div><span>✷</span><h3>Fresh from the Oven</h3><p>กลิ่นหอมและความสดใหม่ในทุกวัน</p></div><div><span>☼</span><h3>Made for Sharing</h3><p>ความสุขเล็ก ๆ ที่เหมาะกับการแบ่งปัน</p></div></div>
        </div>
      </section>

      <section className="catalog section-cream" id="products">
        <div className="container">
          <span className="eyebrow">FIND YOUR FAVORITE</span>
          <h2>Explore Our <em>Menu</em></h2>
          <div className="catalog-controls">
            <div className="filters"><button className={!category ? 'active' : ''} onClick={() => void refresh('')}>ทั้งหมด</button>{initialCategories.map(item => <button className={category === item.id ? 'active' : ''} key={item.id} onClick={() => void refresh(item.id)}>{item.name_th}</button>)}</div>
            <input aria-label="ค้นหาขนม" value={query} onChange={event => setQuery(event.target.value)} placeholder="ค้นหาขนม..." />
          </div>
          {error && <div className="error" role="alert">{error}</div>}
          {!error && !filtered.length && <div className="empty">ไม่พบสินค้าที่ค้นหา</div>}
          <div className="product-grid">{filtered.map(product => <article className="product-card" key={product.id}>
            <span className="product-category">{initialCategories.find(item => item.id === product.category_id)?.name_th || 'TongBai'}</span>
            <Link href={`/products/${product.slug}`}><h3>{product.name}</h3></Link>
            <strong>{money(product.price_thb)}</strong>
            <Link className="product-image" href={`/products/${product.slug}`}><img src={imageFor(product, fallbackImages[0])} alt={product.images?.[0]?.alt_text || product.name} /></Link>
            <button className="round-add" disabled={!product.stock_quantity} onClick={() => add(product)} aria-label={`เพิ่ม ${product.name} ลงตะกร้า`}>{product.stock_quantity ? '+' : '×'}</button>
          </article>)}</div>
        </div>
      </section>

      <section className="delivery dark-section">
        <div className="container delivery-grid">
          <div className="delivery-photo"><img src={fallbackImages[0]} alt="ขนมอบพร้อมแบ่งปัน" /></div>
          <div><span className="eyebrow">FROM OUR OVEN TO YOUR DOOR</span><h2>Craving Something<br /><em>Wonderful?</em></h2><p>เลือกขนมที่ชอบ แล้วให้เราเตรียมความอร่อยส่งถึงคุณ</p><Link className="gold-button" href="/checkout">สั่งซื้อเลย <span aria-hidden="true">↗</span></Link><div className="delivery-notes"><span>✦ อบด้วยใจ</span><span>✦ พร้อมแบ่งปัน</span><span>✦ สั่งง่ายออนไลน์</span></div></div>
        </div>
      </section>

      <section className="gallery-section section-cream" id="gallery"><div className="container"><span className="eyebrow">A LITTLE GLIMPSE OF TONGBAI</span><h2>Life at <em>TongBai</em></h2><div className="gallery-grid">{[fallbackImages[0], fallbackImages[1], fallbackImages[2], ...initialProducts.slice(0, 3).map(product => imageFor(product, fallbackImages[0]))].map((src, index) => <img key={`${src}-${index}`} src={src} alt={`บรรยากาศและขนม TongBai ${index + 1}`} />)}</div><a className="gold-button gallery-cta" href="#products">ดูเมนูทั้งหมด <span aria-hidden="true">↗</span></a></div></section>
    </main>

    <footer className="site-footer dark-section"><div className="container"><div className="footer-brand">TongBai <span>bakery ✦</span></div><div className="footer-links"><a href="#home">HOME</a><a href="#products">MENU</a><a href="#about">OUR STORY</a><a href="#gallery">GALLERY</a><Link href="/orders">MY ORDERS</Link></div><p>© 2026 TongBai · {initialStore?.accepting_orders === false ? 'ขณะนี้ปิดรับออเดอร์' : 'เปิดรับออเดอร์'}</p></div></footer>
    <Link className="cart-fab" href="/checkout" aria-label={`ไปชำระเงิน สินค้าในตะกร้า ${count} ชิ้น`}>ตะกร้า <strong>{count}</strong> <span aria-hidden="true">↗</span></Link>
  </>;
}
