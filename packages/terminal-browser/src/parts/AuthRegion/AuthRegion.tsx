import { diode, cls } from "@/utils/classname";
import Avatar from "../Avatar";
import styles from "./AuthRegion.module.scss";

export type AuthRegionProps = {
	fold: boolean;
	name: string;
};

export default function AuthRegion({ fold }: AuthRegionProps) {
	return (
		<div {...cls(styles.root, diode(styles.fold)(fold))}>
			<div>
				<Avatar source={import("$images/avatar/avatar001.jpg")} />
				<div className="tray"></div>
			</div>
		</div>
	);
}
