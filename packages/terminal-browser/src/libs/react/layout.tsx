import {
	LayoutComponentType,
	LayoutElementProvider,
	PageComponentType,
} from "@/types/react";

const layoutMap = new Map<PageComponentType<any>, LayoutComponentType>();
const providerMap = new Map<LayoutComponentType, LayoutElementProvider>();

const EmptyLayout: LayoutComponentType = ({ children }) => <>{children}</>;
const fallbackProvider: LayoutElementProvider = (page) => (
	<EmptyLayout>{page}</EmptyLayout>
);
providerMap.set(EmptyLayout, fallbackProvider);

export function mapLayout<P>(
	PageComponent: PageComponentType<P>,
	Layout: LayoutComponentType
) {
	layoutMap.set(PageComponent, Layout);
	return PageComponent;
}

export function registProvider(
	Layout: LayoutComponentType,
	NestBy?: LayoutComponentType
) {
	const Provider: LayoutElementProvider = NestBy
		? (page) => getProviderByLayout(NestBy)(<Layout>{page}</Layout>)
		: (page) => <Layout>{page}</Layout>;

	providerMap.set(Layout, Provider);
	return Layout;
}

export function getProvider(Component: PageComponentType) {
	const Layout = layoutMap.get(Component) ?? EmptyLayout;
	return providerMap.get(Layout) ?? fallbackProvider;
}

export function getProviderByLayout(Layout: LayoutComponentType) {
	return providerMap.get(Layout) ?? fallbackProvider;
}
