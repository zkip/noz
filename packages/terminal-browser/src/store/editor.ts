import { atom } from "recoil";

export const editorValueState = atom<string>({
    key: "editorValue",
    default: "",
});
