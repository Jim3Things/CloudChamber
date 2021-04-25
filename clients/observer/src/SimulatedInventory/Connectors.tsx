// Connectors is a support visual component that draws connection points,
// and colors them, based on a boolean area indicating if the connector is
// enabled or not

import React from "react";
import {grey} from "@material-ui/core/colors";

export function Connectors(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    state: boolean[],
    onColor: string,
    offColor: string,
    opacity: number
}) {
    const radius = Math.min((props.height / 4) - 2, (props.width / props.state.length))
    const cy = props.height - radius - 1 + props.y
    const dx = props.width / props.state.length

    const linkColor = (flag: boolean): string => {
        if (flag) return props.onColor
        return props.offColor
    }

    let leftX = props.x + (dx / 2)

    return (
        <React.Fragment>

            {props.state.map((value: boolean) => {
                const thisX = leftX
                leftX += dx

                return <circle
                    cx={thisX}
                    cy={cy}
                    r={radius}
                    fill={linkColor(value)}
                    strokeWidth={1}
                    stroke={grey[700]}
                    fillOpacity={props.opacity}
                />
            })}

        </React.Fragment>
    )
}
