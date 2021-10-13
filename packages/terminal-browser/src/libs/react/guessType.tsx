import { ReactElement, ReactFragment, ReactNode } from "react";

export function isReactElement(node: ReactNode): node is ReactElement {
	return "key" in (node ?? {}) && "props" in (node ?? {});
}
