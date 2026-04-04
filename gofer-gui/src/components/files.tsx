import { createSignal, createEffect, type Setter } from "solid-js";
import {SeeDir, type FileEnt} from "../App";
import prettyBytes from "pretty-bytes";

type Props = {
    f: FileEnt,
}
type File = {
    el: HTMLDivElement,
    path: string,
    selector: Setter<boolean>,
}

export const [selected, setSelected] = createSignal<string[]>([]);
export let fileDivs: File[] = [];
const formatter: Intl.DateTimeFormat = new Intl.DateTimeFormat("en-NL", {
    dateStyle: "medium",
    timeStyle: "short",
});

export default function FileEntry(props: Props) {
    const [s, setS] = createSignal<boolean>(false);
    let button: HTMLButtonElement;
    const click = (e: MouseEvent) => {
        const element: HTMLElement = (e.target as HTMLElement);
        if (element.classList.contains("select")) return;
        if (!props.f.file) SeeDir(`/${props.f.path}`);
    }
    createEffect(() => {
        if (s()) {
            setSelected(v => [...v, props.f.path]);
            button.classList.add("selected");
            button.parentElement?.classList.add("selected");
        } else {
            setSelected(v => v.filter(item => item !== props.f.path));
            button.classList.remove("selected");
            button.parentElement?.classList.remove("selected");
        }
    });
    const select = (e: MouseEvent) => {
        e.stopPropagation();
        setS(v => !v);
    }

    return <div onClick={click} ref={(el) => fileDivs.push({el: el, path: props.f.path, selector: setS})} class="entry">
        <button class="select" ref={e=>button=e} onClick={select}>
            <img src="/gui/check.svg" width="12" />
        </button>
        <span class="name">{props.f.name}</span>
        <span class="mod">{formatter.format(new Date(props.f["mod-time"] * 1000))}</span>
        <span class="size">{props.f.size > 0? prettyBytes(props.f.size):"-"}</span>
    </div>;
}
