import { ImageDataType } from "@/types/image";
import { PropsWithClassName } from "@/types/react";
import {
	ComponentType,
	ReactComponentElement,
	useEffect,
	useState,
} from "react";

export type ImageProps = {
	source: Promise<ImageDataType>;
	holder?: ReactComponentElement<ComponentType>;
} & PropsWithClassName;

const ImageHolderDefault = () => <div>loading...</div>;

export default function Image({
	className,
	source,
	holder = <ImageHolderDefault />,
}: ImageProps) {
	const [src, setSrc] = useState("");
	useEffect(() => {
		async function fetch() {
			const exports = await source;
			setSrc(exports.default.src);
		}
		fetch();
	}, [source]);
	return src ? <img src={src} className={className} /> : holder;
}
