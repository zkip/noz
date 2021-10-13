import { PropsWithChildren, PropsWithClassName } from "@/types/react";
import { cls, diode } from "@/utils/classname";
import { noop } from "@/utils/constants";
import Icon from "../Icon";
import styles from "./Modal.module.scss";

export type ModalProps = {
	title?: string;
	onCloseBtnClick?: () => void;
} & PropsWithChildren &
	PropsWithClassName;

export default function Modal({
	onCloseBtnClick = noop,
	children,
	className,
}: ModalProps) {
	return (
		<div {...cls(styles.root, className)}>
			<Icon
				code="icm-close"
				{...cls(styles["close-btn"])}
				onClick={() => onCloseBtnClick()}
			/>
			{children}
		</div>
	);
}
