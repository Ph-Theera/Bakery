/** @type {import('next').NextConfig} */
import { PHASE_DEVELOPMENT_SERVER } from 'next/constants.js';
export default function config(phase) {
  const internalApi = (process.env.API_INTERNAL_BASE_URL || 'http://127.0.0.1:8080/api/v1').replace(/\/$/, '');
  return {
    reactStrictMode: true,
    distDir: phase === PHASE_DEVELOPMENT_SERVER ? '.next-dev' : '.next-build',
    async rewrites() {
      return [{ source: '/api/v1/:path*', destination: `${internalApi}/:path*` }];
    },
  };
}
