import { createEffect, For, type Accessor } from "solid-js";
import { click } from "./pathbar";

type Props = {
    visible: Accessor<boolean>,
    compressed: string[],
}

export default function Compressed(props: Props) {
    let div: HTMLDivElement;
    createEffect(() => {
        if (props.visible())
            div.style.display = "flex";
        else
            div.style.display = "none";
    })

    return <div ref={e=>div=e}>
        <For each={props.compressed}>{(v: string, i: Accessor<number>) =>
            <button onClick={(_: MouseEvent) => click(i())}>{v}</button>
        }</For>
    </div>
}
