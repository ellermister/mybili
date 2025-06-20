import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import laravel from 'laravel-vite-plugin';
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
    server: {
        hmr: {
            host: 'localhost',
        },
        watch: {
        }
    },
    plugins: [
        laravel({
            input: ['resources/css/app.css', 'resources/js/main.ts'],
            refresh: true,
        }),
        vue({
            template: {
                transformAssetUrls: {
                    base: null,
                    includeAbsolute: false,
                },
            },
        }),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./resources/js', import.meta.url))
        }
    }
});
