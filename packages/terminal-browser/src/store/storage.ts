import { SetterOrUpdater } from "recoil";

export function persistLastTimeForSentVCode(time: Date) {
	localStorage.setItem("lastTimeForSentVCode", time.toString());
}

export function restoreLastTimeForSentVCode(
	setLastTime: SetterOrUpdater<Date>
) {
	setLastTime(new Date(localStorage.getItem("lastTimeForSentVCode") ?? 0));
}
