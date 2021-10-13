import { registProvider } from "@/libs/react/layout";
import { PropsWithChildren } from "@/types/react";
import CommonLayout from "./CommonLayout";

export default function LayoutSecond({ children }: PropsWithChildren) {
	return (
		<div>
			<span className="title">Title Title.</span>
			<input type="text" />
			<div className="content">{children}</div>
		</div>
	);
}

registProvider(LayoutSecond, CommonLayout);
