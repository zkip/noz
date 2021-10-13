import { atom } from "recoil";

export const authVisitationState = atom({
	key: "authVisitation",
	default: { phone: false, vcode: false, passwd: false },
});
