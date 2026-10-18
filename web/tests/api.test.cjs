const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const ts = require('typescript');
require.extensions['.ts'] = (module, filename) => {
  const source = fs.readFileSync(filename, 'utf8');
  module._compile(ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText, filename);
};
process.env.NEXT_PUBLIC_API_MOCK = 'false';
process.env.NEXT_PUBLIC_API_BASE_URL = 'http://localhost:8080/api/v1';
const { api, ApiError } = require('../lib/api.ts');
const auth = require('../lib/auth.ts');
let storage;
test.beforeEach(() => {
  storage = new Map();
  global.window = { sessionStorage: { getItem: key => storage.get(key) || null, setItem: (key, value) => storage.set(key, value) } };
  auth.setAdminToken('');
});
test('live errors do not fallback to mock products', async () => {
  global.fetch = async () => new Response(JSON.stringify({ error: { message: 'conflict' } }), { status: 409 });
  await assert.rejects(api.products(), e => e instanceof ApiError && e.status === 409);
});
test('checkout sends stable idempotency key and remembers returned order token', async () => {
  global.fetch = async (url, init) => {
    assert.equal(url, 'http://localhost:8080/api/v1/orders');
    assert.equal(init.headers.get('Idempotency-Key'), 'request-key');
    assert.equal(init.headers.get('Content-Type'), 'application/json');
    return Response.json({ id: 'order-1', access_token: 'order-token' });
  };
  await api.createOrder({ items: [] }, 'request-key');
  assert.equal(auth.orderCredentials()['order-1'], 'order-token');
});
test('guest order history uses individual scoped tokens', async () => {
  auth.rememberOrder('order-1', 'order-token');
  global.fetch = async (url, init) => {
    assert.equal(url, 'http://localhost:8080/api/v1/orders/order-1');
    assert.equal(init.headers.get('Authorization'), 'Bearer order-token');
    return Response.json({ id: 'order-1' });
  };
  assert.equal((await api.orders())[0].id, 'order-1');
});
test('slip uses multipart browser boundary and owner credentials', async () => {
  auth.rememberOrder('order-1', 'order-token');
  global.fetch = async (url, init) => {
    assert.equal(init.headers.get('Authorization'), 'Bearer order-token');
    assert.equal(init.headers.has('Content-Type'), false);
    assert.ok(init.body instanceof FormData);
    assert.equal(init.body.get('slip').name, 'slip.png');
    return Response.json({ status: 'uploaded' });
  };
  await api.uploadPaymentSlip('payment-1', new File(['png'], 'slip.png', { type: 'image/png' }), 'order-1');
});
test('admin endpoints require explicit token and reject forbidden access', async () => {
  global.fetch = async () => { throw new Error('must not fetch without token'); };
  await assert.rejects(api.adminProducts(), e => e.status === 403);
  auth.setAdminToken('private-admin-token');
  global.fetch = async (url, init) => {
    assert.equal(init.headers.get('Authorization'), 'Bearer private-admin-token');
    return new Response('{}', { status: 403 });
  };
  await assert.rejects(api.adminProducts(), e => e.status === 403);
});
test('product image upload uses admin token and multipart form', async () => {
  auth.setAdminToken('private-admin-token');
  global.fetch = async (url, init) => {
    assert.equal(url, 'http://localhost:8080/api/v1/admin/products/product-1/images');
    assert.equal(init.headers.get('Authorization'), 'Bearer private-admin-token');
    assert.equal(init.headers.has('Content-Type'), false);
    assert.equal(init.body.get('image').name, 'bread.png');
    assert.equal(init.body.get('alt_text'), 'Bread');
    return Response.json({ id: 'product-1', images: [] });
  };
  await api.uploadProductImage('product-1', new File(['png'], 'bread.png', { type: 'image/png' }), 'Bread');
});
test('wrong slip MIME is rejected before any request', () => {
  assert.throws(() => api.uploadPaymentSlip('p', new File(['text'], 'bad.txt', { type: 'text/plain' }), 'o'), /5MB/);
});
test('browser uses same-origin API while server uses internal backend URL', async () => {
  const modulePath = require.resolve('../lib/api.ts');
  process.env.NEXT_PUBLIC_API_BASE_URL = '/api/v1';
  process.env.API_INTERNAL_BASE_URL = 'http://127.0.0.1:8080/api/v1';
  try {
    delete require.cache[modulePath];
    const browserApi = require(modulePath);
    global.fetch = async url => { assert.equal(url, '/api/v1/products?limit=100'); return Response.json([]); };
    await browserApi.api.products();
    assert.equal(browserApi.imageUrl('/api/v1/products/images/id'), '/api/v1/products/images/id');
    delete global.window;
    delete require.cache[modulePath];
    const serverApi = require(modulePath);
    global.fetch = async url => { assert.equal(url, 'http://127.0.0.1:8080/api/v1/products?limit=100'); return Response.json([]); };
    await serverApi.api.products();
  } finally {
    process.env.NEXT_PUBLIC_API_BASE_URL = 'http://localhost:8080/api/v1';
    delete process.env.API_INTERNAL_BASE_URL;
    delete require.cache[modulePath];
  }
});
