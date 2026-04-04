import {path, SeeDir} from "./App";
import { createSignal, For } from "solid-js";

export default function Sidebar() {
    let favourites: Array<{n: string, p: string}> = [];
    cookieStore.get("fav").then(c => {
        if (!c) return;
        if (typeof c.value === "undefined") return;
        favourites = JSON.parse(c.value);
    });
    const [fav, _] = createSignal([
      {n: "Home", p: "/"}, 
      ...favourites
    ]);
    const click = (e: Event, to: string) => {
        SeeDir(to);
        (e.target as HTMLButtonElement).classList.add("selected");
    }
    return <div id="sidebar">
        <img src="/gui/gofer.svg"/>
        <h4>Favourites</h4>
        <For each={fav()}>{(v: {n: string, p: string}) =>
            <button onClick={(e) => click(e, v.p)} class={path() == v.p.slice(1)? "selected":""}>{v.n}</button>
        }</For>
    </div>
}
