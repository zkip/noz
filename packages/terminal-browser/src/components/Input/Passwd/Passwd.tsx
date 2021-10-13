import { PropsWithClassName } from "@/types/react";
import { cls, diode } from "@/utils/classname";
import styles from "./Passwd.module.scss";
import { Sifter, Validator } from "@/types/form";
import {
	genValidation,
	genValidator,
	mergeValidator,
	Validators,
} from "@/libs/validations";
import { noop, Noop } from "@/utils/constants";
import { mergeSifter, Sifters } from "@/libs/sifters";
import { isInRange } from "@/utils/array";

const atLesatOneLetter = (v: string) => /[a-z|A-Z]+/.test(v);
const atLesatOneNumber = (v: string) => /[0-9]+/.test(v);
const inValidLength = ({ length }: string) => isInRange(6, 13)(length);
const isValidPasswd = (passwd: string) =>
	[inValidLength, atLesatOneLetter, atLesatOneNumber].reduce(
		(valid, fn) => valid && fn(passwd),
		true
	);

export const defaultValidator = mergeValidator(
	Validators.notEmpty("密码"),
	genValidator((passwd: string) =>
		genValidation(isValidPasswd(passwd) ? "" : "密码为6-12位数字+字母组合")
	)
);

export type PasswdInputProps = {
	passwd?: string;
	placeholder?: string;

	validator?: Validator<string>;
	filter?: Sifter<string>;

	onChanged?: (text: string) => void | Noop;
} & PropsWithClassName;
export default function PasswdInput({
	passwd = "",
	placeholder = "请输入密码",

	validator = defaultValidator,
	filter = mergeSifter(Sifters.maxLength(12)),
	onChanged = noop,
}: PasswdInputProps) {
	const validation = validator.validate(passwd);
	const errMegSeg = validation.valid ? null : (
		<div {...cls(styles["error-message"])}>{validation.msg}</div>
	);
	return (
		<div {...cls(styles.root, diode(styles.invalid)(!validation.valid))}>
			<input
				{...cls(styles.passwd)}
				type="password"
				value={filter(passwd)}
				placeholder={placeholder}
				onChange={(e) => onChanged(filter(e.target.value))}
			/>
			<div {...cls(styles["decoration-line"])} />
			{errMegSeg}
		</div>
	);
}
