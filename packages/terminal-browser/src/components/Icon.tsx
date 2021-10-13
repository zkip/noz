import icmStyles from "$icons/icm-v1.0/style.css";
import ftoStyles from "$icons/fontello-v1.0/css/fontello.css";
import animationStyles from "$icons/fontello-v1.0/css/animation.css";
import { PropsWithClassName } from "@/types/react";
import { cls } from "@/utils/classname";
import { DOMAttributes } from "react";
export type ICMIconTypes = keyof typeof icmStyles;
export type FTOIconTypes = keyof typeof ftoStyles;
export type AnimationTypes = keyof typeof animationStyles;
export type IconFamily = "icm" | "fto";

export type IconProps = {
	code: ICMIconTypes | FTOIconTypes;
	animate?: AnimationTypes;
} & PropsWithClassName &
	DOMAttributes<HTMLElement>;

export default function Icon({ code, animate, className, ...rest }: IconProps) {
	return <i {...cls(code, animate, className)} {...rest} />;
}
