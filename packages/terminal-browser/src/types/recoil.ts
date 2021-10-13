import { RecoilState } from "recoil";

export type UnpackedRecoilState<T> = T extends RecoilState<infer U> ? U : T;
