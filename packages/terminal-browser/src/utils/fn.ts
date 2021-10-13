import { DOMEventType, getTopObject } from "./constants";

export function listen(target: EventTarget = getTopObject()) {
	return {
		on(type: DOMEventType, option?: AddEventListenerOptions) {
			return (fn: EventListener) => {
				target.addEventListener(type, fn, option);
				return () => target.removeEventListener(type, fn, option);
			};
		},
	};
}

export const fallback =
	<T, K>(value: T) =>
	(v: T | K = value) =>
		v;

export const runAll = (...fns: Function[]) => {
	let result;
	for (const fn of fns) {
		result = fn(result);
	}
	return result;
};
