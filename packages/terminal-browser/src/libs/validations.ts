import { Validation, Validator } from "@/types/form";
import { isInRange } from "@/utils/array";

export function genValidation(msg: string) {
	const valid = msg === "";
	return { valid, msg };
}

export function genValidator<T>(validate: (v: T) => Validation): Validator<T> {
	return { validate };
}

export function mergeValidator<T>(...validators: Validator<T>[]): Validator<T> {
	const validate = (value: any) =>
		validators.reduce(
			(validation, validator) =>
				validation.valid ? validator.validate(value) : validation,
			genValidation("")
		);
	return {
		validate,
	};
}

export const Validators = {
	whatever: () => genValidator(() => genValidation("")),
	notEmpty: (name = "") =>
		genValidator((v: string) =>
			genValidation(v === "" ? `${name}不能为空` : "")
		),
	lengthIn: (min: number, max: number, name: string) =>
		genValidator(({ length }: string) =>
			genValidation(
				isInRange(min, max + 1)(length)
					? ""
					: `${name}长度在${min}和${max}之间`
			)
		),
	lengthLess: (max: number, name: string) =>
		genValidator(({ length }: string) =>
			genValidation(length < max ? "" : `${name}长度少于${max}`)
		),
	lengthMore: (min: number, name: string) =>
		genValidator(({ length }: string) =>
			genValidation(length >= min ? "" : `${name}长度不少于${min}`)
		),
	length: (n: number, name: string) =>
		genValidator(({ length }: string) =>
			genValidation(length === n ? "" : `${name}长度为${n}位`)
		),
};
