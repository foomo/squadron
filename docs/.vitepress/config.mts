import { defineConfig } from "vitepress";

export default defineConfig({
	title: "Squadron",
	description: "Docker Compose for Kubernetes",
	lang: "en-US",
	cleanUrls: true,
	lastUpdated: true,
	appearance: "dark",
	ignoreDeadLinks: false,
	base: "/squadron/",
	sitemap: {
		hostname: "https://foomo.github.io/squadron",
	},
	themeConfig: {
		logo: "/logo.png",
		outline: [2, 4],
		nav: [
			{ text: "Guide", link: "/guide/introduction" },
			{ text: "Reference", link: "/reference/" },
		],
		sidebar: [
			{
				text: "Guide",
				items: [
					{ text: "Introduction", link: "/guide/introduction" },
					{ text: "Installation", link: "/guide/installation" },
					{ text: "Quick Start", link: "/guide/quickstart" },
					{ text: "Core Concepts", link: "/guide/concepts" },
					{ text: "Configuration", link: "/guide/configuration" },
				],
			},
			{
				text: "Reference",
				items: [
					{ text: "Overview", link: "/reference/" },
					{
						text: "CLI",
						link: "/reference/cli/squadron",
						collapsed: true,
						items: [
							{ text: "squadron", link: "/reference/cli/squadron" },
							{ text: "up", link: "/reference/cli/squadron_up" },
							{ text: "down", link: "/reference/cli/squadron_down" },
							{ text: "diff", link: "/reference/cli/squadron_diff" },
							{ text: "status", link: "/reference/cli/squadron_status" },
							{ text: "rollback", link: "/reference/cli/squadron_rollback" },
							{ text: "bake", link: "/reference/cli/squadron_bake" },
							{ text: "build", link: "/reference/cli/squadron_build" },
							{ text: "push", link: "/reference/cli/squadron_push" },
							{ text: "list", link: "/reference/cli/squadron_list" },
							{ text: "config", link: "/reference/cli/squadron_config" },
							{ text: "template", link: "/reference/cli/squadron_template" },
							{ text: "schema", link: "/reference/cli/squadron_schema" },
							{
								text: "completion",
								link: "/reference/cli/squadron_completion",
							},
							{ text: "version", link: "/reference/cli/squadron_version" },
						],
					},
				],
			},
			{
				text: "Contributing",
				collapsed: true,
				items: [
					{
						text: "Guideline",
						link: "/CONTRIBUTING.md",
					},
					{
						text: "Code of conduct",
						link: "/CODE_OF_CONDUCT.md",
					},
					{
						text: "Security guidelines",
						link: "/SECURITY.md",
					},
				],
			},
		],
		socialLinks: [
			{ icon: "github", link: "https://github.com/foomo/squadron" },
		],
		editLink: {
			pattern: "https://github.com/foomo/squadron/edit/main/docs/:path",
		},
		search: {
			provider: "local",
		},
		footer: {
			message:
				'Made with ♥ <a href="https://www.foomo.org">foomo</a> by <a href="https://www.bestbytes.com">bestbytes</a>',
		},
	},
	markdown: {
		// https://github.com/vuejs/vitepress/discussions/3724
		theme: {
			light: "catppuccin-latte",
			dark: "catppuccin-frappe",
		},
	},
	head: [
		["meta", { name: "theme-color", content: "#ffffff" }],
		["link", { rel: "icon", href: "/logo.png" }],
		["meta", { name: "author", content: "foomo by bestbytes" }],
		// OpenGraph
		["meta", { property: "og:title", content: "foomo/squadron" }],
		[
			"meta",
			{
				property: "og:image",
				content:
					"https://github.com/foomo/squadron/blob/main/docs/public/banner.png?raw=true",
			},
		],
		[
			"meta",
			{
				property: "og:description",
				content: "Docker Compose for Kubernetes",
			},
		],
		["meta", { name: "twitter:card", content: "summary_large_image" }],
		[
			"meta",
			{
				name: "twitter:image",
				content:
					"https://github.com/foomo/squadron/blob/main/docs/public/banner.png?raw=true",
			},
		],
		[
			"meta",
			{
				name: "viewport",
				content: "width=device-width, initial-scale=1.0, viewport-fit=cover",
			},
		],
	],
});
