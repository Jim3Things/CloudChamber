// Draw a TOR box, with the currently defined connections.  Currently, these
// connections are to the blades in that rack.

// TODO: Figure out how/if to show connections to the individual instances via
//       an SDN setup

import React from "react"
import {TorDetails} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {Connectors} from "./Connectors"

export function Tor(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: TorDetails,
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
                textAnchor={"center"}>TOR
            </text>

            <Connectors
                x={offset}
                y={props.y}
                width={connectionWidth}
                height={props.height}
                state={props.details.linkTo}
                onColor={props.palette.runningColor}
                offColor={props.palette.faultedColor}
                opacity={Opacity(props.details.state)}/>

        </React.Fragment>
    )
}
