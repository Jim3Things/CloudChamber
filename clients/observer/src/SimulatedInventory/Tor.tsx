// Draw a TOR box, with the currently defined connections.  Currently, these
// connections are to the blades in that rack.

// TODO: Figure out how/if to show connections to the individual instances via
//       an SDN setup

import React, {FunctionComponent} from "react";
import {TorDetails} from "../proxies/InventoryProxy";
import {Colors} from "./SimulatedInventory";
import {Opacity, PhysicalBox} from "./PhysicalBox";
import {Connectors} from "./Connectors";

export const Tor: FunctionComponent<{
    x: number,
    y: number,
    width: number,
    height: number,
    details: TorDetails,
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
                textAnchor={"center"}>TOR</text>

            <Connectors
                x={offset}
                y={y}
                width={connectionWidth}
                height={height}
                state={details.linkTo}
                onColor={palette.runningColor}
                offColor={palette.faultedColor} opacity={Opacity(details.state)} />

        </React.Fragment>
    )
}