// Draw a TOR box, with the currently defined connections.  Currently, these
// connections are to the blades in that rack.

// TODO: Figure out how/if to show connections to the individual instances via
//       an SDN setup

import React from "react"
import {TorDetails} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {Connectors} from "./Connectors"
import {RssFeed} from "@material-ui/icons"
import {Tooltip} from "@material-ui/core"

export function Tor(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: TorDetails,
    palette: Colors
}) {
    const iconWidth = Math.min(props.height, 50)
    const offset = iconWidth + 5
    const connectionWidth = props.width - offset

    return (
        <React.Fragment>
            <Tooltip title="TOR 0">
                <RssFeed
                    x={props.x}
                    y={props.y}
                    width={iconWidth}
                    height={props.height} />
            </Tooltip>

            <PhysicalBox
                x={props.x + offset}
                y={props.y}
                width={connectionWidth}
                height={props.height}
                state={props.details.state}
                palette={props.palette}>

                <Connectors
                    x={0}
                    y={0}
                    width={connectionWidth}
                    height={props.height}
                    state={props.details.linkTo}
                    onColor={props.palette.runningColor}
                    offColor={props.palette.faultedColor}
                    opacity={Opacity(props.details.state)}/>

            </PhysicalBox>

        </React.Fragment>
    )
}
