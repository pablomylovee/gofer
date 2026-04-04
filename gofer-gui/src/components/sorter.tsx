import { createEffect, type Accessor, type Setter } from "solid-js"
import type { FileEnt } from "../App"
import { selected, setSelected, fileDivs } from "./files"

type Props = {
    setFiles: Setter<FileEnt[]>,
    files:    Accessor<FileEnt[]>,
}
type SortThing = {
    up?: HTMLImageElement,
    ad: boolean,
    using: boolean,
}

const saButton = (_: MouseEvent) => {
    if (selected().length > 0) {
        for (const div of fileDivs)
            div.selector(false);
        setSelected([]);
    } else if (selected().length == 0) {
        for (const div of fileDivs)
            div.selector(true);
    }
}

export default function Sorter(props: Props) {
    const sortThings: SortThing[] = [];
    const nameimg: SortThing = {ad: true, using: true};
    sortThings.push(nameimg);

    let sall: HTMLButtonElement;
    const sort = (by: string) => {
        for (const st of sortThings) {
            st.using = false;
            if (st.up)
                st.up.style.display = "none";
        }
        const sortedFiles = [...props.files()].sort((a, b) => {
            switch (by) {
                case "n": {
                    if (nameimg.ad) return a.name < b.name ? 1 : (b.name < a.name ? -1 : 0);
                    else return a.name > b.name ? 1 : (b.name > a.name ? -1 : 0);
                }
                default: return 0;
            }
        });
        props.setFiles(sortedFiles);
    
        switch (by) {
            case "n": {
                if (!nameimg.up) return;
                nameimg.up.style.display = "block";
                if (nameimg.ad) nameimg.up.style.transform = "rotate(180deg)";
                else nameimg.up.style.transform = "rotate(0deg)";
                nameimg.ad = !nameimg.ad;
            }
        }
    };

    createEffect(() => {
        sall.classList.remove("aselected", "sselected");
        const slength: number = selected().length;
        const flength: number = fileDivs.length;
        console.log(slength, flength);

        if (slength === 0 && flength === 0) return;
        if (slength === flength) {
            sall.style.backgroundColor = "#07f";
            sall.classList.add("aselected");
        } else if (slength > 0) {
            sall.style.backgroundColor = "#07f";
            sall.classList.add("sselected");
        } else if (slength === 0)
            sall.style.backgroundColor = "";
    });

    return <div id="sorter">
        <button class="select" style={{
            height: "16px", visibility: "visible",
            "pointer-events": "all", padding: 0,
            "border-style": "solid",
            "margin-right": "6.5px",
        }} onClick={saButton} ref={e=>sall=e}>
            <img src="/gui/check.svg" width="12" class="c" />
            <img src="/gui/dash.svg" width="14" class="d" />
        </button>
        <button class="name" onClick={(_) => sort("n")}>
            Name<img src="/gui/up.svg" width="17" ref={(e) => nameimg.up = e}/>
        </button>
    </div>
}
