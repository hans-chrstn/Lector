import svelte from 'eslint-plugin-svelte';
import ts from 'typescript-eslint';
import svelteParser from 'svelte-eslint-parser';
import tsParser from '@typescript-eslint/parser';
import globals from 'globals';

export default [
	{
		ignores: [
			'.svelte-kit/',
			'node_modules/',
			'build/',
			'static/',
			'eslint.config.js',
			'svelte.config.js',
			'vite.config.ts'
		]
	},

	...ts.configs.recommended,

	...svelte.configs['flat/recommended'],

	{
		files: ['**/*.{js,ts,svelte}'],
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},

	{
		files: ['**/*.svelte', '**/*.svelte.ts'],
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: tsParser,
				extraFileExtensions: ['.svelte']
			}
		},
		rules: {
			'svelte/valid-compile': 'off',
			'@typescript-eslint/no-explicit-any': 'off',
			'@typescript-eslint/no-unused-vars': 'warn'
		}
	}
];
