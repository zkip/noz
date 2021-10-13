import { PropsWithClassName } from "@/types/react";
import { cls, diode, genCSDPairFromStyles } from "@/utils/classname";
import styles from "./PhoneNumber.module.scss";
import { MobilePhoneAreaCode } from "@/types/phoneNumberAreaStandards";
import { Sifter, Validator } from "@/types/form";
import {
	genValidation,
	genValidator,
	mergeValidator,
	Validators,
} from "@/libs/validations";
import { noop, Noop } from "@/utils/constants";
import { mergeSifter, Sifters } from "@/libs/sifters";

const { clsS, diodeS } = genCSDPairFromStyles(styles);

export const phoneNumberValidator = genValidator((phoneNumber: string) =>
	genValidation(/1\d{10}/.test(phoneNumber) ? "" : "请输入正确的手机号")
);

export const defaultPhoneNumberValidator = mergeValidator(
	Validators.notEmpty("手机号码"),
	phoneNumberValidator,
	Validators.lengthIn(11, 11, "手机号码")
);

export const defaultPhoneNumberSifter = mergeSifter(
	Sifters.onlyNumber(),
	Sifters.maxLength(11)
);

export type PhoneNumberInputProps = {
	areaCode?: MobilePhoneAreaCode;
	number?: string;
	placeholder?: string;

	validator?: Validator<string>;
	filter?: Sifter<string>;

	// unimplemented
	onAreaChanged?: (code: MobilePhoneAreaCode) => void | Noop;
	onNumberChanged?: (number: string) => void | Noop;
} & PropsWithClassName;
export default function PhoneNumberInput({
	areaCode = "86",
	number = "",
	placeholder = "请输入手机号",
	validator = defaultPhoneNumberValidator,
	filter = defaultPhoneNumberSifter,
	onNumberChanged = noop,
}: PhoneNumberInputProps) {
	const validation = validator.validate(number);

	const errMegSeg = validation.valid ? null : (
		<div {...clsS("error-message")}>{validation.msg}</div>
	);
	return (
		<div {...clsS("root", diodeS("invalid")(!validation.valid))}>
			<div {...clsS("area-code")}>{areaCode}</div>
			<input
				{...clsS("phone-number")}
				value={filter(number)}
				type="text"
				placeholder={placeholder}
				onChange={(e) => onNumberChanged(filter(e.target.value))}
			/>
			<div {...clsS("decoration-line")} />
			{errMegSeg}
		</div>
	);
}
