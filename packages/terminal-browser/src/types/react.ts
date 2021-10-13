import { NextComponentType, NextPage, NextPageContext } from "next";
import { AppProps } from "next/app";
import React, { ComponentType, ReactElement, ReactNode } from "react";

export interface PropsWithChildren {
	children: ReactNode;
}

export type PropsWithClassName = {
	className?: string;
};

export type LayoutComponentType<
	P extends PropsWithChildren = PropsWithChildren
> = ComponentType<P>;

export type LayoutElementProvider<P = {}> = (
	page: ReactElement<P, PageComponentType<P>>
) => ReactElement<any, LayoutComponentType>;

export type PageComponentType<P = {}> = NextPage<P, {}>;
