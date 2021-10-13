export type Validator<T = {}> = {
	validate(value: T): Validation;
};

export type Validation = {
	msg: string;
	valid: boolean;
};

export type Sifter<T> = (ctx: T) => T;

export type Visitation = { [key: string]: boolean };
