import { User } from "@/model/User";
import { AuthStyle } from "@/types/auth";
import { atom } from "recoil";

export const authStyleState = atom<AuthStyle>({
	key: "authStyle",
	default: "vcode",
});

export const phoneState = atom<string>({
	key: "phone",
	default: "",
});

export const passwdState = atom<string>({
	key: "passwd",
	default: "",
});

export const vcodeState = atom<string>({
	key: "vcode",
	default: "",
});

export const lastTimeForSentVcodeState = atom<Date>({
	key: "lastTimeBySentVcode",
	default: new Date(),
});

export const tokenState = atom({
	key: "token",
	default: "",
});

export const userInfoState = atom<User>({
	key: "userId",
	default: new User(),
});
