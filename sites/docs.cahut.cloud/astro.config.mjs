// @ts-check
import starlight from '@astrojs/starlight';
import brand from '@cahut/brand';
import { defineConfig } from 'astro/config';

export default defineConfig({
	site: 'https://docs.cahut.cloud',
	integrations: [
		brand(),
		starlight({
			title: 'Cahut docs',
			description: 'Install Cahut on your server, connect your AIs over MCP, and share what they make by link.',
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
			expressiveCode: {
				styleOverrides: { borderRadius: 'var(--radius-md)' },
			},
			sidebar: [
				{
					label: 'Start here',
					items: [{ label: 'Overview', slug: 'index' }, 'start/how-it-works', 'start/status'],
				},
				{
					label: 'Install',
					items: ['install/server', 'install/local'],
				},
				{
					label: 'Connect an AI',
					items: ['connect/ai-clients', 'connect/claude-code', 'connect/claude-plugin', 'connect/tokens', 'connect/publish'],
				},
				{
					label: 'Concepts',
					items: [
						'concepts/vocabulary',
						'concepts/cahuts',
						'concepts/rooms-and-windows',
						'concepts/access',
						'concepts/house-style',
						'concepts/knock-first-and-logbook',
					],
				},
				{
					label: 'Administration',
					items: ['admin/console', 'admin/commands', 'admin/service'],
				},
				{
					// Written from the code by make docs-reference.
					label: 'Reference',
					items: [{ autogenerate: { directory: 'reference' } }],
				},
			],
		}),
	],
	server: { port: 4322 },
});
