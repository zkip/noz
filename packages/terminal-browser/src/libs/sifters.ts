import { Sifter } from "@/types/form";

export const Sifters = {
	pass:
		() =>
		<T>(ctx: T) =>
			ctx,
	onlyNumber: () => (ctx: string) => ctx.replaceAll(/[^\d]*/g, ""),
	noNumber: () => (ctx: string) => ctx.replaceAll(/\d*/g, ""),
	maxLength: (length: number) => (ctx: string) => ctx.slice(0, length),
};

export function mergeSifter<T>(...sifters: Sifter<T>[]) {
	return (ctx: T) => sifters.reduce((c: T, s) => s(c), ctx);
}
