import { genCSDPairFromStyles } from "@/utils/classname";
import { useEffect, useRef } from "react";
import Vditor from "vditor";
import styles from "./Editor.module.scss";

const { clsS, diodeS } = genCSDPairFromStyles(styles);

export default function Editor({}) {
    const vditorRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (!vditorRef.current) return;

        const text = `# fkdf
+ sdf
+ sdf
		`;
        const vditor = new Vditor(vditorRef.current, {
            cache: { enable: false },
			mode: "wysiwyg",
            value: text,
        });
        console.log(vditor, "+++++");

        addEventListener("keypress", (e) => {
            if (e.shiftKey && e.key === "G") {
                console.log(vditor.getValue());
            }
        });
    });

    return <div {...clsS("root")} ref={vditorRef}></div>;
}
