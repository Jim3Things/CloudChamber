// PhysicalBox is a support visual component that draws the outer box and
// background for a component that has a physical health state.
//
// Note that this module also exports a utility function which determines
// opacity based on the physical state.  This supports the ability to
// 'grey out' physical boxes that are turned off.

import React, {FunctionComponent} from "react";
import {PhysicalState} from "../proxies/InventoryProxy";
import {Colors} from "./SimulatedInventory";

export const Opacity = (state: PhysicalState): number => {
    switch (state) {
        case PhysicalState.healthy: return 1.0
        case PhysicalState.faulted: return 1.0
        case PhysicalState.off: return .3
        default: return 1.0
    }
}

export const PhysicalBox: FunctionComponent<{
    x: number,
    y: number,
    width: number,
    height: number,
    state: PhysicalState,
    fillOpacity?: number,
    palette: Colors,
    pointerEvents?: string | number,
    onMouseEnter?: React.MouseEventHandler<SVGSVGElement>,
    onMouseLeave?: React.MouseEventHandler<SVGSVGElement>,
}> = (props) => {
    const borderColor = (state: PhysicalState): string => {
        switch (state) {
            case PhysicalState.healthy: return props.palette.runningColor
            case PhysicalState.faulted: return props.palette.faultedColor
            case PhysicalState.off: return props.palette.offColor
            default: return props.palette.illegal
        }
    }

    return (
        <svg
            x={props.x}
            y={props.y}
            width={props.width}
            height={props.height}
            pointerEvents={props.pointerEvents}
            onMouseEnter={props.onMouseEnter}
            onMouseLeave={props.onMouseLeave}
        >
            <rect
                x={0}
                y={0}
                width={props.width}
                height={props.height}
                fill={props.palette.backgroundColor}
                fillOpacity={props.fillOpacity}
                strokeWidth={2}
                stroke={borderColor(props.state)}
                opacity={Opacity(props.state)}
            />

            {props.children}
        </svg>
    )
}
