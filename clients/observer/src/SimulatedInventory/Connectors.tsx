// Connectors is a support visual component that draws connection points,
// and colors them, based on a boolean area indicating if the connector is
// enabled or not

import React, {FunctionComponent} from "react";
import {grey} from "@material-ui/core/colors";

export const Connectors: FunctionComponent<{
    x: number,
    y: number,
    width: number,
    height: number,
    state: boolean[],
    onColor: string,
    offColor: string,
    opacity: number
        }> = ({
            x,
            y,
            width,
            height,
            state,
            onColor,
            offColor,
            opacity}) => {
    const radius = Math.min((height / 4) - 2, (width / state.length))
    const cy = height - radius - 1 + y
    const dx = width / state.length

    const linkColor = (flag: boolean): string => {
        if (flag) return onColor
        return offColor
    }

    let leftX = x + (dx / 2)

    return (
        <React.Fragment>

            {state.map((value) => {
                const thisX = leftX
                leftX += dx

                return <circle
                    cx={thisX}
                    cy={cy}
                    r={radius}
                    fill={linkColor(value)}
                    strokeWidth={1}
                    stroke={grey[700]}
                    fillOpacity={opacity}
                />
            })}

        </React.Fragment>
    )
}