import Image from "@/components/Image";
import { ImageDataType } from "@/types/image";
import { timeout } from "@/utils/async";
import { cls } from "@/utils/classname";
import { useEffect, useState } from "react";
import styles from "./Avatar.module.scss";

export type AvatarProps = {
	source: Promise<ImageDataType>;
};
export default function Avatar({ source }: AvatarProps) {
	return (
		<div {...cls(styles.root)}>
			<Image source={source} {...cls(styles.img)} />
		</div>
	);
}
