import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

export default defineConfig({
    plugins: [vue()],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),
        },
    },
    server: {
        host: true,
        proxy: {
            '/api': 'http://127.0.0.1:8080',
            '/storage': 'http://127.0.0.1:8080',
        },
    },
    build: {
        outDir: 'dist',
    },
});
