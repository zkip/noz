import { PropsWithClassName } from "@/types/react";
import { diode, genCSDPairFromStyles } from "@/utils/classname";

import styles from "./VCode.module.scss";
import { Sifter, Validator } from "@/types/form";
import { mergeValidator, Validators } from "@/libs/validations";
import { noop, Noop } from "@/utils/constants";
import { InputHTMLAttributes, useEffect } from "react";
import { useState } from "react";
import { interval } from "@/utils/async";
import Icon from "@/components/Icon";
import { mergeSifter, Sifters } from "@/libs/sifters";

const { clsS, diodeS } = genCSDPairFromStyles(styles);

export const defaultVCodeValidator = mergeValidator(
	Validators.notEmpty("验证码"),
	Validators.length(4, "验证码")
);

export const defaultVCodeSifter = mergeSifter(
	Sifters.onlyNumber(),
	Sifters.maxLength(4)
);

export type VCodeInputProps = {
	code?: string;
	placeholder?: string;
	visited?: boolean;

	lastTimeForSent: Date;
	timeGap?: number; // ms
	isWaiting?: boolean;

	validator?: Validator<string>;
	filter?: Sifter<string>;

	onCodeChanged?: (code: string) => void | Noop;
	onCodeSent?: (time: Date) => void | Noop | Promise<void>;
	onVisited?: () => void | Noop;
} & PropsWithClassName;
export default function VCodeInput({
	code = "",
	placeholder = "请输入验证码",
	visited = false,

	lastTimeForSent,
	timeGap = 10000, // one minute
	isWaiting = false,

	validator = defaultVCodeValidator,
	filter = defaultVCodeSifter,
	onCodeChanged = noop,
	onCodeSent = noop,
	onVisited = noop,
}: VCodeInputProps) {
	const [now, setNow] = useState(new Date());

	const validation = validator.validate(code);
	const invalid = !visited || validation.valid;

	const gone = now.getTime() - lastTimeForSent.getTime();
	const isAvaliableSend = gone >= timeGap;
	const remainingCount = ((timeGap - gone) / 1000) >> 0;

	const errMegSeg = diode(
		<div {...clsS("error-message")}>{validation.msg}</div>
	)(visited && !validation.valid);

	const onSendBtnClick = () => {
		if (!isAvaliableSend || isWaiting) return;
		onCodeSent(new Date());
	};

	const counterOrSendBtn =
		isAvaliableSend || isWaiting ? (
			<button
				{...clsS("send-btn", diodeS("waiting")(isWaiting))}
				onClick={onSendBtnClick}
			>
				<Icon
					{...clsS("waiting-indicator")}
					code="fto-spin1"
					animate="animate-spin"
				/>
				<span {...clsS("text")}>发送验证码</span>
			</button>
		) : (
			<div {...clsS("counter")}>{remainingCount + 1}s</div>
		);

	useEffect(() => {
		if (!isAvaliableSend && !isWaiting) {
			setNow(new Date());
			return interval(() => setNow(new Date()))(1000);
		}
	}, [isAvaliableSend, isWaiting]);

	const visitationSetup = diode<InputHTMLAttributes<HTMLInputElement>>(
		{
			onBlur: () => !visited && onVisited(),
		},
		{}
	)(!visited);

	return (
		<div {...clsS("root", diodeS("invalid")(!invalid))}>
			<input
				{...clsS("code-input")}
				value={filter(code)}
				type="text"
				placeholder={placeholder}
				onChange={(e) => onCodeChanged(filter(e.target.value))}
				{...visitationSetup}
			/>

			<div
				{...clsS(
					"code-indicator",
					diodeS("avaliable")(isAvaliableSend)
				)}
			>
				{counterOrSendBtn}
			</div>

			<div {...clsS("decoration-line")} />
			{errMegSeg}
		</div>
	);
}
