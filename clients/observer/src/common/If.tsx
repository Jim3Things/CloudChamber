import React from 'react';

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
