import { defineConfig } from "vitepress";
import {
	groupIconMdPlugin,
	groupIconVitePlugin,
} from "vitepress-plugin-group-icons";
import llmstxt, {
	copyOrDownloadAsMarkdownButtons,
} from "vitepress-plugin-llms";
import { withSidebar } from "vitepress-sidebar";

const version = "latest";

const vitepressConfig = {
	title: "Squadron",
	description: "Docker Compose for Kubernetes",
	themeConfig: {
		logo: "/logo.png",
		nav: [
			{
				text: "Guide",
				link: "/guide/",
			},
			{
				text: `${version}`,
				items: [
					{
						text: "Release Notes",
						link: "https://github.com/foomo/squadron/releases",
					},
				],
			},
		],
		outline: {
			level: [2, 3],
		},
		editLink: {
			pattern: "https://github.com/foomo/squadron/edit/main/docs/:path",
			text: "Suggest changes to this page",
		},
		search: {
			provider: "local",
		},
		footer: {
			message: "Released under the MIT License.",
		},
		socialLinks: [
			{
				icon: "github",
				link: "https://github.com/foomo/squadron",
			},
		],
	},
	head: [
		["meta", { name: "theme-color", content: "#ffffff" }],
		["link", { rel: "icon", href: "/logo.png" }],
		["meta", { name: "author", content: "foomo by bestbytes" }],
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
	markdown: {
		theme: {
			dark: "one-dark-pro",
			light: "github-light",
		},
		config(md) {
			md.use(groupIconMdPlugin);
			md.use(copyOrDownloadAsMarkdownButtons);
		},
	},
	vite: {
		plugins: [
			groupIconVitePlugin(),
			llmstxt({
				excludeIndexPage: false,
			}),
		],
	},
	sitemap: {
		hostname: "https://foomo.github.io/squadron",
	},
	ignoreDeadLinks: true,
};

export default defineConfig(
	withSidebar(vitepressConfig, {
		useTitleFromFrontmatter: true,
		frontmatterOrderDefaultValue: 10,
		useFolderTitleFromIndexFile: true,
		sortMenusByFrontmatterOrder: true,
	}),
);
