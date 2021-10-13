export function isBrowser() {
	return process.browser;
}
export function isServer() {
	return !process.browser;
}
