import {path, SeeDir} from "./App";
import { createSignal, For } from "solid-js";
import Cookies from "js-cookie";

export default function Sidebar() {
    let favourites: Array<{n: string, p: string}> = [];
    const favCookie: string | undefined = Cookies.get("fav");

    if (favCookie) {
      try {
        favourites = favCookie.split(";;").map(v => {
          if (v === "") return {n: "", p: ""};
          const parsed = JSON.parse(v);
          return (parsed?.n && parsed?.p) ? parsed : {n: "", p: ""};
        });
      } catch {
        console.warn("Invalid fav cookie format, resetting");
        favourites = [];
      }
    }

    const [fav] = createSignal([
      {n: "Home", p: "/"}, 
      ...favourites.filter(v => v.n && v.p)
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
