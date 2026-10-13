/** @type {import('next').NextConfig} */
import { PHASE_DEVELOPMENT_SERVER } from 'next/constants.js';
export default function config(phase) {
  return { reactStrictMode: true, distDir: phase === PHASE_DEVELOPMENT_SERVER ? '.next-dev' : '.next-build' };
}
