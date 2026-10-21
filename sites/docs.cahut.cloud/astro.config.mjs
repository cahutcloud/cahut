// @ts-check
import starlight from '@astrojs/starlight';
import brand from '@cahut/brand';
import { defineConfig } from 'astro/config';

export default defineConfig({
	site: 'https://docs.cahut.cloud',
	integrations: [
		brand(),
		starlight({
			title: 'Cahut',
			logo: {
				light: '@charte/logos/cahut-logo-horizontal.svg',
				dark: '@charte/logos/cahut-logo-horizontal-inverse.svg',
				replacesTitle: true,
			},
			favicon: '/favicon.svg',
			head: [
				{ tag: 'link', attrs: { rel: 'icon', href: '/favicon.ico', sizes: '32x32' } },
				{ tag: 'link', attrs: { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' } },
				{ tag: 'link', attrs: { rel: 'manifest', href: '/site.webmanifest' } },
			],
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/getcahut/cahut' }],
			customCss: ['@charte/tokens/cahut-fonts.css', '@charte/tokens/cahut-tokens.css', './src/styles/cahut.css'],
		}),
	],
	server: { port: 4322 },
});
