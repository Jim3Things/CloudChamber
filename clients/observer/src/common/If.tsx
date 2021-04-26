import React from 'react';
import {ExpandLess, ExpandMore} from "@material-ui/icons";

// Render the children only if the supplied condition is true
export function RenderIf(props: {cond: boolean, children?: React.ReactNode}) {
    if (props.cond) {
        return (
            <div>
                {props.children}
            </div>
        )
    }

    return null
}

// Hide the children whenever the supplied condition is true
export function HideIf(props: {cond: boolean, children?: React.ReactNode}) {
    return (
        <div hidden={props.cond}>
            {props.children}
        </div>
    )
}

export function MoreOrLess(props: { cond: boolean }) {
    if (props.cond) {
        return <ExpandMore/>
    } else {
        return <ExpandLess/>
    }
}
