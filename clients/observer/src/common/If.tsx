import React, {FunctionComponent} from 'react';

// Render the children only if the supplied condition is true
export const RenderIf: FunctionComponent<{cond: boolean}> = (props) => {
    if (props.cond) {
        return (
            <div>
                {props.children}
            </div>
        )
    }

    return null
};

// Hide the children whenever the supplied condition is true
export const HideIf: FunctionComponent<{cond: boolean}> = (props) => {
    return (
        <div hidden={props.cond}>
            {props.children}
        </div>
    )
}
