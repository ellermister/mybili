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
        vue(
            // {
            //     template: {
            //         transformAssetUrls: {
            //             // The Vue plugin will re-write asset URLs, when referenced
            //             // in Single File Components, to point to the Laravel web
            //             // server. Setting this to `null` allows the Laravel plugin
            //             // to instead re-write asset URLs to point to the Vite
            //             // server instead.
            //             base: null,

            //             // The Vue plugin will parse absolute URLs and treat them
            //             // as absolute paths to files on disk. Setting this to
            //             // `false` will leave absolute URLs un-touched so they can
            //             // reference assets in the public directory as expected.
            //             includeAbsolute: false,
            //         },
            //     },
            // }
        ),
    ],
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./resources/js', import.meta.url))
        }
    }
});
