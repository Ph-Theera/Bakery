import type { Metadata } from 'next';
import './globals.css';
import './production.css';
export const metadata: Metadata = { title: 'TongBai', description: 'ขนมโฮมเมดอบสดใหม่ด้วยความใส่ใจ' };
export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) { return <html lang="th"><body>{children}</body></html>; }
