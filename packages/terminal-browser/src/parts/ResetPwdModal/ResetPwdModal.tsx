import { diode, cls } from "@/utils/classname";
import Avatar from "../Avatar";
import styles from "./AuthModal.module.scss";

export type AuthModalProps = {};

export default function AuthModal({}: AuthModalProps) {
	return (
		<div {...cls(styles.root)}>
			<div>
				<Avatar source={import("$images/avatar/avatar001.jpg")} />
				<div className="tray"></div>
			</div>
		</div>
	);
}
