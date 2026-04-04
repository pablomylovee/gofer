import type { Setter } from "solid-js";
import {SeeDir} from "./App";

type Props = {
    authFunc: Setter<boolean>,
}

export default function AuthWidget(props: Props) {
    console.log("sigma");
    let userInput, passInput;
    const tryAuth = (e: Event) => {
        if (!userInput || !passInput) return;
        const username: string = (userInput as HTMLInputElement).value;
        const password: string = (passInput as HTMLInputElement).value;
        fetch("/session/create", {method: "POST", body: JSON.stringify({
            Username: username, Password: password,
        })}).then(r => {
            if (r.ok) {
                props.authFunc(true);
                SeeDir("/");
            }
            else {
                (e.target as HTMLButtonElement).textContent = "Failed!";
                setTimeout(() => (e.target as HTMLButtonElement).textContent = "Login", 1500);
            }
        });
    }

    return <div id="auth">
        Login with your username and password.
        <input ref={userInput} type="text" placeholder="Username here..."/>
        <input ref={passInput} type="password" placeholder="Password here..."/>
        <button onClick={tryAuth}>Login</button>
    </div>;
}
