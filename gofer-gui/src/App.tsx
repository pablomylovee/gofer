import {createSignal, onMount, Show, For, type Setter} from "solid-js";
import { setSelected, fileDivs } from "./components/files.tsx";
import AuthWidget from "./auth.tsx";
import FileEntry from "./components/files.tsx";
import Sidebar from "./sidebar.tsx";
import './App.css';
import Pathbar from "./components/pathbar.tsx";
import Sorter from "./components/sorter.tsx";

export type FileEnt = {
    name: string,
    path: string,
    "mod-time": number,
    size: number,
    file: boolean,
    type: string
}
export type Props = {
    filesFunc: Setter<FileEnt[]>,
}

const [files, setFiles] = createSignal<FileEnt[]>([]);
export function SeeDir(to: string) {
    setStat("Getting files...");
    setSelected([]); setFiles([]);
    fileDivs.length = 0;
    fetch(`/dir${to}`).then(r => {
        if (r.ok && r.headers.get("Content-Type")?.startsWith("application/json")) {
            setPath(to.slice(1));
            return r.json();
        } else if (!r.ok) {
            setStat(`Error! (Code ${r.status})`);
            return false;
        }
    }).then(f => {
        if (f === false) return;
        else if (f === null || f.length === 0) {
            setStat("No files here!");
            return;
        }
        setFiles(f);
    });
}

export const [path, setPath] = createSignal("/");
const [stat, setStat] = createSignal("Getting files...");
export default function App() {
    console.log("ligma");
    const [authenticated, setAuth] = createSignal<boolean>(false);
    onMount(() => {
        setStat("Getting files...");
        setFiles([]);
        fetch(`/dir`).then(r => {
            if (r.ok && r.headers.get("Content-Type")?.startsWith("application/json")) {
                setPath(""); setAuth(true);
                return r.json();
            } else if (!r.ok) return false;
        }).then(d => {
            if (d === false) return;
            else if (d === null || d.length === 0) {
                setStat("No files here!");
                return;
            }
            setFiles(d);
        });
    });

    return <Show when={authenticated()} fallback={<AuthWidget authFunc={setAuth}/>}>
        <Sidebar/>
        <div id="content">
        <Pathbar/>
        <Sorter files={files} setFiles={setFiles}/>
        <For each={files()} fallback={<span style={{
            position: "absolute", top: "70px",
            left: "50%", transform: "translateX(-50%)",
        }}>{stat()}</span>}>
            {(v: FileEnt) =>
                <FileEntry f={v}/>
            }
        </For>
        </div>
    </Show>
}

