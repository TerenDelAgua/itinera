import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './tests',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	// The SvelteKit dev server (Vite) is single-threaded under the hood;
	// running 8+ Playwright workers against it saturates the request
	// pipeline and Svelte's hydration races with our fill() calls. We
	// cap at 2 workers which gives enough parallelism for fast suites
	// without the saturation. CI runs with 1 worker for determinism.
	workers: process.env.CI ? 1 : 2,
	reporter: 'html',
	use: {
		baseURL: 'http://localhost:5173',
		trace: 'on-first-retry',
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] },
		},
	],
	webServer: {
		command: process.env.CI
			? 'pnpm preview --port 5173 --strictPort'
			: 'pnpm dev',
		url: 'http://localhost:5173',
		reuseExistingServer: !process.env.CI,
		timeout: 120_000,
	},
});
