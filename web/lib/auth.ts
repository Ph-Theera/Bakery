let adminToken = '';
const key = 'tongbai-order-access';
export function setAdminToken(token: string) { adminToken = token.trim(); }
export function getAdminToken() { return adminToken; }
export function orderCredentials(): Record<string, string> {
  if (typeof window === 'undefined') return {};
  try { return JSON.parse(window.sessionStorage.getItem(key) || '{}'); } catch { return {}; }
}
export function rememberOrder(id: string, token: string) { window.sessionStorage.setItem(key, JSON.stringify({ ...orderCredentials(), [id]: token })); }
export function userToken() { return typeof window === 'undefined' ? '' : window.sessionStorage.getItem('tongbai-user-access') || ''; }
export function rememberUser(token: string) { window.sessionStorage.setItem('tongbai-user-access', token); }
