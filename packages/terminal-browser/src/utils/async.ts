export const timeout = (ms: number) => new Promise((rv) => setTimeout(rv, ms));

export const interval = (fn: () => void) => {
	return (ms: number) => {
		const interval = setInterval(fn, ms);
		return () => {
			clearInterval(interval);
		};
	};
};

export const makeChannel = <T>(inherit?: Channel<T>): Channel<T> => {
	let resolve: (v: T) => void;
	let reject: (v: T) => void;
	const channel = (
		inherit
			? inherit.then()
			: new Promise((rv, rj) => {
					resolve = rv;
					reject = rj;
			  })
	) as Channel<T>;

	channel.put = inherit
		? inherit.put
		: (value) => {
				if (resolve !== undefined) {
					resolve(value);
				}
		  };
	channel.break = inherit
		? inherit.break
		: (value) => {
				if (reject !== undefined) {
					reject(value);
				}
		  };

	return channel;
};

export type Channel<T = {}> = Promise<T> & {
	put(value: T): void;
	break(value: T): void;
};
