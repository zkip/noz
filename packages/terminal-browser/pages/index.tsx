import Editor from "@/parts/Editor";
import CommonLayout from "@/layouts/CommonLayout";
import { mapLayout } from "@/libs/react/layout";
import { useEffect } from "react";

export default function Home() {
    useEffect(() => {
        async function start() {
            const resp = await fetch("http://127.0.0.1:7703", {
                method: "SDF",
            });
            console.log(resp, "+++++");
        }
        start();
    }, []);
    return <Editor></Editor>;
}

mapLayout(Home, CommonLayout);
