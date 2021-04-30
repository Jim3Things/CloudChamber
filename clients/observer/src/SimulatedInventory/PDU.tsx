// Draw a PDU component, with the currently defined connections.
//
// Connections that are 'true' indicate that the equivalent blade
// should be getting power; those that are 'false' indicate that
// the equivalent blade should also be either powered off or faulted.
// (powered off, if the 'false' state is intentional, faulted if it
// is not)

import React from "react"
import {PduDetails} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {Connectors} from "./Connectors"

export function PDU(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: PduDetails,
    palette: Colors
}) {
    const offset = 60
    const connectionWidth = props.width - offset

    return (
        <React.Fragment>
            <PhysicalBox
                x={props.x}
                y={props.y}
                width={props.width}
                height={props.height}
                state={props.details.state}
                palette={props.palette}/>

            <text
                x={10}
                y={props.height + props.y - 2}
                textAnchor={"center"}>PDU
            </text>

            <Connectors
                x={offset}
                y={props.y}
                width={connectionWidth}
                height={props.height}
                state={props.details.powerTo}
                onColor={props.palette.runningColor}
                offColor={props.palette.faultedColor}
                opacity={Opacity(props.details.state)}/>

        </React.Fragment>
    )
}
