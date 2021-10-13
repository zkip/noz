import { ComponentType, ReactPortal } from "react";
import { createPortal } from "react-dom";

export type PortalComponentType = () => ReactPortal;

export function withPortal<P>(Component: ComponentType, element: Element) {
	return (props: P) => createPortal(<Component {...props} />, element);
}
