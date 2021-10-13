import { isBrowser } from "@/utils/env";
import { listen, runAll } from "@/utils/fn";
import { cls, diode } from "@/utils/classname";
import {
	ComponentType,
	Dispatch,
	ReactNode,
	SetStateAction,
	useEffect,
	useRef,
	useState,
} from "react";
import events from "events";
import { createPortal } from "react-dom";
import { atom, selector, useRecoilValue, useSetRecoilState } from "recoil";
import { noop } from "@/utils/constants";

export class ModalController extends events.EventEmitter {
	visible = false;
	private setVisible: Dispatch<SetStateAction<boolean>> = noop;
	constructor() {
		super();
	}
	setup(visible: boolean, setVisible: Dispatch<SetStateAction<boolean>>) {
		this.visible = visible;
		this.setVisible = setVisible;
	}
	toggle() {
		if (this.visible) {
			this.close();
		} else {
			this.open();
		}
	}
	close() {
		this.setVisible(false);
		this.emit("closed");
	}
	open() {
		this.setVisible(true);
		this.emit("opened");
	}
}

const controllers = new Map<symbol, ModalController>();

export const authModalControllerID = registController();

export function getController(id: symbol) {
	return controllers.get(id);
}

export function registController(id: symbol = Symbol()) {
	controllers.set(id, new ModalController());
	return id;
}

export type ModalComponentType = ComponentType<ModalProps>;

const globalModalState = atom({
	key: "globalModal",
	default: new Map<ReactNode, boolean>(),
});

const modalLayerVisibleState = selector({
	key: "modalLayerVisible",
	get({ get }) {
		const d = Array.from(get(globalModalState)).some(
			([_, visible]) => visible
		);
		return d;
	},
});

export function closeAllModal() {
	Array.from(controllers).map(([id, controller]) => controller.close());
}

export type ModalProps = {
	visible?: boolean;
	target: Element;
	children: ReactNode;
};
const ModalItem = ({ visible = false, target, children }: ModalProps) => {
	if (isBrowser()) {
		return createPortal(visible ? children : null, target);
	} else {
		return null;
	}
};

export function useOverlayLayer(id: symbol) {
	return (segment: ReactNode) => {
		const updateModals = useSetRecoilState(globalModalState);
		const [visible, setVisible] = useState(false);

		const controller = controllers.get(id);

		controller?.setup(visible, setVisible);

		useEffect(() => {
			if (isBrowser()) {
				updateModals(
					(modals) => (modals.set(segment, visible), new Map(modals))
				);
				return () => {
					updateModals(
						(modals) => (modals.delete(segment), new Map(modals))
					);
				};
			}
		}, [visible]);

		return controller;
	};
}

export const ModalLayer = () => {
	const modals = useRecoilValue(globalModalState);
	const modalLayerVisible = useRecoilValue(modalLayerVisibleState);
	const modalLayerRef = useRef<HTMLDivElement>(null);

	useEffect(() => {
		const node = modalLayerRef.current;
		if (!node) return;
		const cleanWheel = listen(window).on("wheel", {
			capture: true,
			passive: false,
		})((e) => e.preventDefault());
		const cleanClick = listen(node).on("click")((e) => {
			const depth = e.composedPath().length;
			if (depth === 6) {
				closeAllModal();
			}
		});
		return () => runAll(cleanWheel, cleanClick);
	}, [modalLayerRef]);

	const segments = modalLayerRef.current
		? Array.from(modals).map(([segment, visible], i) => (
				<ModalItem
					children={segment}
					visible={visible}
					target={modalLayerRef.current as HTMLDivElement}
					key={i}
				/>
		  ))
		: null;
	return (
		<div
			{...cls("modal-layer", diode("visible")(modalLayerVisible))}
			ref={modalLayerRef}
		>
			{segments}
		</div>
	);
};
