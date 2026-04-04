import { For, createSignal, type Accessor } from "solid-js";
import {path, SeeDir} from "../App";
import Compressed from "./compressed";

export const click = (index: number, to?: string) => {
    const pathList: string[] = path().split("/");
    let path1: string = "/"+pathList.slice(0, index+1).join("/");
    SeeDir(to ?? path1);
}

export default function Pathbar() {
    const [visible, setVisible] = createSignal<boolean>(false);
    const compressed = () => {
        const p: string[] = path().split("/");
        return p.slice(0, p.length-2);
    }
    const segments = () => {
        const p: string[] = path().split("/");
        if (p.length >= 4)
            return ["...", ...p.slice(p.length-2)];
        else return p;
    }

    return <div id="pathbar">
        <button onClick={() => click(-1, "/")} class={path() == ""? "selected":""}>~</button>
        <For each={segments()}>{(v: string, i: Accessor<number>) => {
            if (v === "...") return <>
                <img width="20" src="/gui/arrow.svg"/>
                <button onClick={() => setVisible(v => !v)} class={path().endsWith(v)? "selected":""}>
                    {v}<Compressed visible={visible} compressed={compressed()}/>
                </button>
            </>
            else if (v !== "") return <>
                <img width="20" src="/gui/arrow.svg"/>
                <button onClick={() => click(i())} class={path().endsWith(v)? "selected":""}>{v}</button>
            </>
        }}</For>
    </div>;
}

