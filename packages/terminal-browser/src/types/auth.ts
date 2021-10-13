export type AuthByPwd = {
	phone: string;
	passwd: string;
};

export type AuthByVC = {
	phone: string;
	vcode: string;
};

export type AuthStyle = "vcode" | "passwd";
export type AuthModalMode = AuthStyle | "resetPwd";
