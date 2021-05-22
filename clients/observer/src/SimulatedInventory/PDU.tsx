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
import {Power} from '@material-ui/icons'
import {Tooltip} from "@material-ui/core"

export function PDU(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: PduDetails,
    palette: Colors
}) {
    const iconWidth = Math.min(props.height, 50)
    const offset = iconWidth + 5
    const connectionWidth = props.width - offset

    return (
        <React.Fragment>
            <Tooltip title="PDU 0">
                <Power
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
                    state={props.details.powerTo}
                    onColor={props.palette.runningColor}
                    offColor={props.palette.faultedColor}
                    opacity={Opacity(props.details.state)}/>

            </PhysicalBox>
        </React.Fragment>
    )
}
