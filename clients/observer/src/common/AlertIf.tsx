import {Alert, AlertTitle} from "@material-ui/lab";
import React from "react";

import {HideIf} from "./If";

// AlertIf displays an error alert with the supplied text and title, if there
// is some text to show.  it is otherwise hidden.
export function AlertIf(props: {
        text: string,
        title: string
    }) {
    return <HideIf cond={props.text === ""}>
        <Alert severity="error">
            <AlertTitle>{props.title}</AlertTitle>
            {props.text}
        </Alert>
    </HideIf>
}