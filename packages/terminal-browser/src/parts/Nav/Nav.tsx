import style from "./Nav.module.scss";

import Link from "next/link";
import { cls } from "@/utils/classname";
import { PropsWithClassName } from "@/types/react";
import styles from "./Nav.module.scss";
import { Space } from "@/components/Layouts/Space";
import AuthRegion from "../AuthRegion";
import { authModalControllerID, useOverlayLayer } from "@/libs/react/modal";
import Modal from "@/components/Modal";
import Logo, { LogoThemeOption } from "@/components/Logo";
import AuthModal from "../AuthModal";

export type NavProps = {
	theme?: "transparent" | "dark";
} & PropsWithClassName;

const Nav = ({ theme = "transparent", className }: NavProps) => {
	const logoTheme: LogoThemeOption = theme === "dark" ? "dark" : "light";
	const controller = useOverlayLayer(authModalControllerID)(
		<AuthModal
			onCloseBtnClick={() => {
				controller?.close();
			}}
		/>
	);

	return (
		<div {...cls(style.root, className)}>
			<Logo theme={logoTheme} {...cls(styles.logo, "Logo")} />
			<Link href={"/"}>
				<a>首页</a>
			</Link>
			<Link href={"/contact"}>
				<a>联系我们</a>
			</Link>
			<Space />
			<button
				onClick={() => {
					controller?.open();
				}}
			>
				open
			</button>
			<div className="auth-region">
				<AuthRegion fold={false} name="hell" />
			</div>
		</div>
	);
};

export default Nav;
