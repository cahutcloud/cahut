// @ts-check
import brand from '@cahut/brand';
import { defineConfig } from 'astro/config';

export default defineConfig({
	site: 'https://cahut.cloud',
	integrations: [brand()],
	server: { port: 4321 },
});
