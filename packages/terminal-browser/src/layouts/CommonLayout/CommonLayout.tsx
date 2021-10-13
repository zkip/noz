import { ValueOf } from "@/types/tools";
import { cls } from "@/utils/classname";
import { PropsWithChildren, PropsWithClassName } from "@/types/react";
import { registProvider } from "@/libs/react/layout";
import styles from "./CommonLayout.module.scss";
import Nav, { NavProps } from "@/parts/Nav";

type CommonLayoutProps = {
	theme?: ValueOf<NavProps>;
} & NavProps &
	PropsWithChildren &
	PropsWithClassName;

const CommonLayout = ({
	theme = "transparent",
	className,
	children,
}: CommonLayoutProps) => {
	return (
		<div {...cls(styles.root, className)}>
			<div {...cls(styles.tray)}></div>
			<Nav theme={theme} {...cls(styles.nav)} />
			<div {...cls(styles.content)}>{children}</div>
		</div>
	);
};

registProvider(CommonLayout);

export default CommonLayout;
