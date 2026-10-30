import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: path.resolve(__dirname, '../internal/server/dist'),
    emptyOutDir: false,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/e': 'http://localhost:8080',
      '/microhog.js': 'http://localhost:8080',
      '/microhog-recorder.js': 'http://localhost:8080',
    },
  },
});
