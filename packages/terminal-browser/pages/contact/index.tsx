import CommonLayout from "@/layouts/CommonLayout";
import { mapLayout } from "@/libs/react/layout";
import { PageComponentType } from "@/types/react";
import { cls } from "@/utils/classname";
import styles from "./Contact.module.scss";

type PropsContact = { name?: string };

const Contact: PageComponentType<PropsContact> = ({ name = "SD" }) => {
	return (
		<div {...cls(styles.root)}>
			Contact
			<input />
		</div>
	);
};

Contact.getInitialProps = async (ctx) => {
	
	const d = await fetch("http://localhost:3000/api/hello");

	console.log(await d.json(), "+========");

	return { name: "SDF" };
};

mapLayout(Contact, CommonLayout);
export default Contact;
