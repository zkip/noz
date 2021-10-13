import LogoDark from "$images/logo-dark.png";
import LogoLight from "$images/logo-light.png";
import { PropsWithClassName } from "@/types/react";
import { cls } from "@/utils/classname";
import Image from "next/image";
import styles from "./Logo.module.scss";

export type LogoThemeOption = "light" | "dark";
export type LogoProps = {
	theme: LogoThemeOption;
} & PropsWithClassName;
const Logo = ({ theme = "light", className }: LogoProps) => {
	const which = theme === "dark" ? LogoDark.src : LogoLight.src;
	return <img {...cls(className, styles.root)} src={which} alt="logo" />;
};

export default Logo;
