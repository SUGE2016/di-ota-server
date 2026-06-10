import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  base: '/ota/',
  plugins: [react()],
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/ota/api': {
        target: 'http://ota-api:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ota/, ''),
      },
      '/ota/device': {
        target: 'http://ota-api:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/ota/, ''),
      },
    },
  },
});
