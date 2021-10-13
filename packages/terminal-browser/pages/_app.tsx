import "styles/globals.css";
import "styles/vditor_fix.css";
import "vditor/src/assets/scss/index.scss";
import type { AppProps } from "next/app";
import { RecoilRoot } from "recoil";
import { getProvider } from "@/libs/react/layout";
import { PageComponentType } from "@/types/react";
import { ModalLayer } from "@/libs/react/modal";

import "$icons/icm-v1.0/style.css";
import "$icons/fontello-v1.0/css/fontello.css";
import "$icons/fontello-v1.0/css/animation.css";
import { useEffect } from "react";
import "@/libs/setup";
import { setup } from "@/libs/setup";

function Init() {
    const initHook = setup();

    useEffect(() => {
        initHook();
    }, []);

    return <></>;
}

function MyApp({ Component, pageProps, router }: AppProps) {
    const Page = Component as PageComponentType;
    const provide = getProvider(Page);
    const content = provide(<Page {...pageProps} />);

    return (
        <RecoilRoot>
            <Init />
            <ModalLayer />
            {content}
        </RecoilRoot>
    );
}

export default MyApp;
