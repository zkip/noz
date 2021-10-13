import { Visitation } from "@/types/form";
import { atom, RecoilState, useRecoilState } from "recoil";

export function useVisitation<T extends Visitation>(
	atom: RecoilState<T>
): [T, (which: keyof T) => void, () => void] {
	const [visitation, updateVisitation] = useRecoilState(atom);

	return [
		visitation,
		function visit(which) {
			updateVisitation((visitation) => {
				return { ...visitation, [which]: true };
			});
		},
		function visitAll() {
			updateVisitation((visitation) => {
				return Object.keys(visitation).reduce(
					(m, name) => ({
						...m,
						[name]: true,
					}),
					{} as T
				);
			});
		},
	];
}
