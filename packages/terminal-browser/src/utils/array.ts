export const inBound = (a: number, b: number) => (v: number) =>
	Math.max(Math.min(a, b), Math.min(v, Math.max(a, b)));

export const inArrayBound = <T>({ length }: ArrayLike<T>) =>
	inBound(0, length - 1);

export const isInRange = (min: number, max: number) => (n: number) =>
	n >= min && n < max;
