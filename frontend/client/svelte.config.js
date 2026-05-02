import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter({
			pages: '../server/files/build',
			assets: '../server/files/build',
			fallback: 'index.html', // Essential for SPA routing
			precompress: false,
			strict: true
		}),
		alias: {
			"$shadcn": "./shadcn",
			"$shadcn/*": "./shadcn/*"
		},
		csp: {
			directives: {
				'default-src': ['self'],
				'script-src': ['self', 'https://accounts.google.com/gsi/client'],
				'connect-src': ['self', 'https://api.example.oxl.app', 'http://localhost:8080'],
				'style-src': ['self', 'unsafe-inline', 'https://accounts.google.com/gsi/style'],
				'frame-src': ['self', 'https://accounts.google.com/'],
			},
			reportOnly: {
				'script-src': ['self'],
				'report-uri': ['/'],
			}
		}
	}
};

export default config;