import {
	authStyleState,
	lastTimeForSentVcodeState,
	passwdState,
	phoneState,
	vcodeState,
} from "@/store/auth";
import { diode, cls, genCSDPairFromStyles } from "@/utils/classname";
import { useRecoilState, useResetRecoilState } from "recoil";
import styles from "./AuthModal.module.scss";
import Modal from "@/components/Modal";
import Image from "@/components/Image";
import PhoneNumber from "@/components/Input/PhoneNumber";
import Passwd from "@/components/Input/Passwd";
import VCodeInput from "@/components/Input/VCode";
import { persistLastTimeForSentVCode } from "@/store/storage";
import { useEffect, useState } from "react";
import { timeout } from "@/utils/async";
import { useVisitation } from "@/libs/react/useVisitation";
import { authVisitationState } from "@/store/form";
import { noop } from "@/utils/constants";
import { authModalControllerID, getController } from "@/libs/react/modal";

export type AuthModalProps = {
	onCloseBtnClick?: () => void;
};

const { clsS, diodeS } = genCSDPairFromStyles(styles);

export default function AuthModal({ onCloseBtnClick = noop }: AuthModalProps) {
	const [authStyle, setAuthStyle] = useRecoilState(authStyleState);
	const [phone, setPhone] = useRecoilState(phoneState);
	const [passwd, setPasswd] = useRecoilState(passwdState);
	const [vcode, setVCode] = useRecoilState(vcodeState);
	const [lastTime, setLastTime] = useRecoilState(lastTimeForSentVcodeState);
	const [isWaiting, setWaiting] = useState(false);
	const [visitation, visit, visitAll] = useVisitation(authVisitationState);
	const resetAuthVisitation = useResetRecoilState(authVisitationState);

	const toggleAuthStyle = () =>
		setAuthStyle(authStyle === "passwd" ? "vcode" : "passwd");

	const isPwdStyle = authStyle === "passwd";

	const onSentCode = async () => {
		setWaiting(true);
		await timeout(3000);
		const now = new Date();

		persistLastTimeForSentVCode(now);
		setLastTime(now);
		setWaiting(false);
	};

	const onModalCloseBtnClick = () => {
		onCloseBtnClick();
	};

	const onAuthBtnClick = () => {
		visitAll();
	};

	useEffect(() => {
		const controller = getController(authModalControllerID);
		controller?.on("closed", () => {
			resetAuthVisitation();
		});
	}, []);

	return (
		<Modal
			{...clsS("prefer", "modal")}
			onCloseBtnClick={onModalCloseBtnClick}
		>
			<div {...clsS("root")}>
				<Image
					{...clsS("logo")}
					source={import("$images/logo-dark.png")}
				/>
				<div {...clsS("slogan")}>????????????,??????????????????</div>

				<PhoneNumber number={phone} onNumberChanged={setPhone} />

				{diode(
					<Passwd passwd={passwd} onChanged={setPasswd} />,
					<VCodeInput
						visited={visitation.vcode}
						lastTimeForSent={lastTime}
						code={vcode}
						isWaiting={isWaiting}
						onCodeChanged={setVCode}
						onCodeSent={onSentCode}
						onVisited={() => visit("vcode")}
					/>
				)(isPwdStyle)}

				<div {...clsS("style-toggle-btn")} onClick={toggleAuthStyle}>
					{diode("?????????????????????", "????????????")(isPwdStyle)}
				</div>
				<button onClick={onAuthBtnClick}>
					{diode("??????", "??????/??????")(isPwdStyle)}
				</button>

				{diode(
					<span>
						?????????????????????<a href="#">????????????????????????</a>
					</span>
				)(!isPwdStyle)}
			</div>
		</Modal>
	);
}
