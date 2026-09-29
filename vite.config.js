import { defineConfig } from 'vite'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
    plugins: [tailwindcss()],
    build: {
        outDir: 'static',
        emptyOutDir: false, // Prevents deleting your other static assets
        watch: {
            buildDelay: 5000,
        },
        rollupOptions: {
            // Define your frontend entry points
            input: {
                main: resolve(import.meta.dirname, 'vite/js/main.js'),
            },
            output: {
                // Keeps file names predictable without hashes (easier for Beego to link)
                entryFileNames: 'js/[name].js',
                chunkFileNames: 'js/[name].js',
                assetFileNames: '[ext]/[name].[ext]',
            },
        },
    },
})
