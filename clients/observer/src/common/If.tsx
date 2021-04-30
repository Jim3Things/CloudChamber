import React from 'react'
import {CheckBox, CheckBoxOutlineBlank, ExpandLess, ExpandMore} from "@material-ui/icons"

// This module contains simple helper functions that perform a basic if-else
// task.

// Render the children only if the supplied condition is true
export function RenderIf(props: { cond: boolean, children?: React.ReactNode }) {
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
export function HideIf(props: { cond: boolean, children?: React.ReactNode }) {
    return (
        <div hidden={props.cond}>
            {props.children}
        </div>
    )
}

// Show either expand or retract (aka more or less) icons, depending on the
// supplied condition.
export function MoreOrLess(props: { cond: boolean }) {
    if (props.cond) {
        return <ExpandMore/>
    } else {
        return <ExpandLess/>
    }
}

// Show a checked or empty display-only checkbox, depending on the supplied
// condition.
export function CheckIf(props: { cond: boolean }) {
    if (props.cond) {
        return <CheckBox/>
    } else {
        return <CheckBoxOutlineBlank/>
    }
}
