export type ClassNameOption = {
	[name: string]: boolean;
};
export type ClassNamePair = string | undefined | ClassNameOption;
const isOptionClassNamePair = (v: any): v is ClassNameOption =>
	typeof v === "object";
export const pass = <T>(v: T) => v;

export const joinClass = (...args: (string | undefined)[]) =>
	args.filter(Boolean).join(" ");
export function cls(...classnames: (string | undefined)[]) {
	const className = joinClass(...classnames);

	const extra = (...plainNames: string[]) => {
		return {
			className: joinClass(className, ...plainNames),
		};
	};
	return Object.assign(extra, {
		className,
	});
}

// generate Classname Selector and style-based Diode from styles
export function genCSDPairFromStyles<T extends { [name: string]: string }>(
	styles: T
) {
	return {
		clsS: (...classnames: (keyof T | undefined)[]) =>
			cls(...classnames.map((name) => name && styles[name])),
		diodeS: (value: keyof T, another?: keyof T) => diode(value, another),
	};
}

export type Diode<T> = (ok: boolean) => T | undefined;
export const diode = function <T, A = T>(value: T, another?: A): Diode<T | A> {
	return (ok = true) => (ok ? value : another);
};
export const makeStyleDiodes = function <K>(
	style: { [key in keyof K]: string }
) {
	return Object.entries(style).reduce(
		(map, [name, token]) => ({ ...map, [name]: diode(token) }),
		{}
	) as { [key in keyof K]: Diode<string> };
};
