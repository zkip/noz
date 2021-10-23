import Editor from "@/parts/Editor";
import CommonLayout from "@/layouts/CommonLayout";
import { mapLayout } from "@/libs/react/layout";
import { useEffect, useState } from "react";
import { isNotFound } from "@/utils/constants";
import { lastOne } from "@/utils/array";

interface HierarchyRecord {
    Size: number;
    Order: number;
    ID: string;
    Name: string;
    Path: Array<string>;
}

function genTentBlock(data: object) {
    const m = data as { [key: string]: HierarchyRecord };
    // record children map, { ID: Children }
    const om = {} as { [key: string]: HierarchyRecord[] };
    const omr = [] as HierarchyRecord[];

    const omm = {} as { [key: string]: number };

    const put = (record: HierarchyRecord) => {
        if (record.Path.length === 0) {
            omr[record.Order] = record;
            return;
        }

        const parentID = lastOne(record.Path);

        if (!(parentID in om)) {
            const parentRecord = m[parentID];
            om[parentID] = new Array(parentRecord.Size);
            put(parentRecord);
        }

        om[parentID][record.Order] = record;
    };

    const maxOrder = (record: HierarchyRecord): number => {
        if (record.ID in omm) {
            return omm[record.ID];
        }

        if (record.Size > 0) {
            return (omm[record.ID] =
                om[record.ID].reduce((c, r) => maxOrder(r) + c, 0) + 1);
        }
        return (omm[record.ID] = 1);
    };

    Object.values(m).map(put);
    Object.values(m).map(maxOrder);

    console.log("++++++++++++++");

    console.log(om);

    const orderCaches = new Map<string, number>();

    const ds = Object.values(m);
    const resolveOrder = (record: HierarchyRecord): number => {
        const orderChached = orderCaches.get(record.ID);

        if (!isNotFound(orderChached)) {
            return orderChached;
        }

        const isRoot = record.Path.length === 0;
        const parentID = lastOne(record.Path);
        const prevRecord =
            record.Order > 0 ? om[parentID][record.Order - 1] : m[parentID];

        const prevRange = record.Order > 0 ? omm[prevRecord.ID] : 0;
        const prevOrder = isRoot
            ? -1
            : resolveOrder(prevRecord) + (record.Order > 0 ? prevRange - 1 : 0);

        const order = prevOrder + 1;
        orderCaches.set(record.ID, order);
        return order;
    };

    ds.map(resolveOrder);

    const debug = (_m: Map<string, number>) => {
        const rs = new Map<string, number>();

        for (const [i, n] of _m.entries()) {
            rs.set(m[i].Name, n);
        }
        return rs;
    };
    console.log(debug(orderCaches), ">>>>>>>>>>>>>");

    return ds
        .sort((a, b: HierarchyRecord) => {
            return resolveOrder(a) - resolveOrder(b);
        })
        .map((record) => (
            <div
                key={record.ID}
                style={{ left: `${20 * record.Path.length}px` }}
            >
                <span>{record.Name}</span>
                <span>{orderCaches.get(record.ID)}</span>
            </div>
        ));
}

export default function Home() {
    const [blocks, setBlocks] = useState<JSX.Element[]>();
    useEffect(() => {
        async function start() {
            const resp = await fetch("/hierarchy_record/list", {
                method: "ACTION",
                headers: {
                    Authorization:
                        "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImMwMjdjNzY2LTA2OWYtNDRlNC1hZGRlLTRjY2I2OGQwMTEzYSIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYzNzI5OTMwOSwidXNlcl9pZCI6MX0.2PURiNGehNl-h2Wz8gl9er_wu367hhNlsi12Cwe-lpg",
                },
            });
            const { Data } = await resp.json();
            console.log(Data, "@@@@@@@@@@@");

            const blocks = genTentBlock(Data);
            setBlocks(blocks);
        }
        start();
    }, []);
    // return <Editor></Editor>;

    return <div className="SFD">{blocks}</div>;
}

mapLayout(Home, CommonLayout);
