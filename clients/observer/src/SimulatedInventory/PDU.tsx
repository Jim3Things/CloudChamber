// Draw a PDU component, with the currently defined connections.
//
// Connections that are 'true' indicate that the equivalent blade
// should be getting power; those that are 'false' indicate that
// the equivalent blade should also be either powered off or faulted.
// (powered off, if the 'false' state is intentional, faulted if it
// is not)

import React, {FunctionComponent} from "react";
import {PduDetails} from "../proxies/InventoryProxy";
import {Colors} from "./SimulatedInventory";
import {Opacity, PhysicalBox} from "./PhysicalBox";
import {Connectors} from "./Connectors";

export const PDU: FunctionComponent<{
    x: number,
    y: number,
    width: number,
    height: number,
    details: PduDetails,
    palette: Colors
}> = ({x, y, width, height, details, palette}) => {
    const offset = 60
    const connectionWidth = width - offset

    return (
        <React.Fragment>
            <PhysicalBox
                x={x}
                y={y}
                width={width}
                height={height}
                state={details.state}
                palette={palette}/>

            <text
                x={10}
                y={height + y - 2}
                textAnchor={"center"}>PDU
            </text>

            <Connectors
                x={offset}
                y={y}
                width={connectionWidth}
                height={height}
                state={details.powerTo}
                onColor={palette.runningColor}
                offColor={palette.faultedColor}
                opacity={Opacity(details.state)} />

        </React.Fragment>
    )
}