import { lastTimeForSentVcodeState } from "@/store/auth";
import { restoreLastTimeForSentVCode } from "@/store/storage";
import { useSetRecoilState } from "recoil";

export function setup() {
	const setLastTime = useSetRecoilState(lastTimeForSentVcodeState);
	return () => {
		restoreLastTimeForSentVCode(setLastTime);
	};
}
