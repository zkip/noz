module.exports = {
	reactStrictMode: true,
	trailingSlash: false,
	pageExtensions: ["tsx", "page.ts","api.ts"],
	webpack: (config) => {
		return config;
	},
};
