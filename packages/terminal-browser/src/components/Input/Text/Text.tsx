import { PropsWithClassName } from "@/types/react";
import { cls, diode } from "@/utils/classname";
import styles from "./Text.module.scss";
import { Validator } from "@/types/form";
import { Validators } from "@/libs/validations";
import { noop, Noop } from "@/utils/constants";

export type TextInputProps = {
	text?: string;
	validator?: Validator<string>;

	onChanged?: (text: string) => void | Noop;
} & PropsWithClassName;
export default function TextInput({
	text = "",
	onChanged = noop,
	validator = Validators.whatever(),
}: TextInputProps) {
	const validation = validator.validate(text);
	const errMegSeg = validation.valid ? null : (
		<div {...cls(styles["error-message"])}>{validation.msg}</div>
	);
	return (
		<div {...cls(styles.root, diode(styles.invalid)(!validation.valid))}>
			<input
				{...cls(styles.text)}
				type="text"
				value={text}
				onChange={(e) => onChanged(e.target.value)}
			/>
			<div {...cls(styles["decoration-line"])} />
			{errMegSeg}
		</div>
	);
}
