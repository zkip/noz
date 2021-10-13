import { PropsWithClassName } from "@/types/react";
import { cls, diode } from "@/utils/classname";
import { PropsWithChildren, ReactNode } from "react";
import Icon, { ICMIconTypes } from "../Icon";
import styles from "./Card.module.scss";

export type CardProps = {
	title: string;

	hint?: ICMIconTypes;
	shadow?: "always" | "hover" | "never";

	header?: ReactNode;
	content?: ReactNode;
	footer?: ReactNode;
} & PropsWithClassName &
	PropsWithChildren<{}>;
export default function Card({
	title,
	hint = "icm-close",
	shadow = "always",
	header,
	content,
	footer,
	children,
}: CardProps) {
	const headerDefaultSeg = (
		<div {...cls(styles.title)}>
			<Icon code={hint} />
			<span>{title}</span>
		</div>
	);
	const contentDefaultSeg = <div></div>;
	const footerDefaultSeg = <div></div>;
	return (
		<div {...cls(styles.root, `shadow-${shadow}`)}>
			<div {...cls(styles.header)}>{header ?? headerDefaultSeg}</div>
			<div {...cls(styles.content)}>
				{content ?? contentDefaultSeg}
				{children}
			</div>
			<div {...cls(styles.footer)}>{footer ?? footerDefaultSeg}</div>
		</div>
	);
}
