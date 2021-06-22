// Draw a TOR box, with the currently defined connections.  Currently, these
// connections are to the blades in that rack.

// TODO: Figure out how/if to show connections to the individual instances via
//       an SDN setup

import React from "react"
import {PhysicalState, TorDescription} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {Connectors} from "./Connectors"
import {
    createStyles,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Tooltip
} from "@material-ui/core"
import {
    CableState_SM,
    cableState_SMToString,
    hardware_HwTypeToString,
    TorState_SM
} from "../pkg/protos/inventory/common"
import {makeStyles} from "@material-ui/core/styles"
import {DetailCard} from "./DetailCard"
import {NetworkOffIcon, NetworkOnIcon, TorIcon} from "../common/Icons"
import {usePopup} from "../common/UsePopup"

const useStyles = makeStyles(() =>
    createStyles({
        list: {
            fontSize: "small",
        },
    }),
)

// toPhysical is a transitional conversion function to handle the
// partial removal of PhysicalState. It converts a TOR SM state
// into the closest physical state approximation.
function toPhysical(s: TorState_SM): PhysicalState {
    switch (s) {
        case TorState_SM.UNRECOGNIZED:
        case TorState_SM.stuck:
        case TorState_SM.invalid:
            return PhysicalState.faulted

        default:
            return PhysicalState.healthy
    }
}

function networkDetail(onOff: boolean) {
    if (onOff) {
        return <NetworkOnIcon />
    } else {
        return <NetworkOffIcon />
    }
}

export function Tor(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: TorDescription,
    palette: Colors,
    index: number
}) {
    const classes = useStyles()

    const [open, anchorEl, handlePopoverOpen, handlePopoverClose] = usePopup<SVGSVGElement>()

    const iconWidth = Math.min(props.height, 50)
    const offset = iconWidth + 5
    const connectionWidth = props.width - offset

    const state = toPhysical(props.details.tor.observed.smState)

    const wires = Array.from(props.details.tor.ports).map(v =>
        (state === PhysicalState.healthy) &&
        (v[1].observed.smState === CableState_SM.on))

    return (
        <>
            <Tooltip title={"TOR " + props.index}>
                <TorIcon
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
                state={state}
                palette={props.palette}
                pointerEvents="all"
                aria-owns={open ? 'mouse-over-popover' : undefined}
                aria-haspopup="true"
                onMouseEnter={handlePopoverOpen}
                onMouseLeave={handlePopoverClose}
            >

                <Connectors
                    x={0}
                    y={0}
                    width={connectionWidth}
                    height={props.height}
                    state={wires}
                    onColor={props.palette.runningColor}
                    offColor={props.palette.faultedColor}
                    opacity={Opacity(state)}
                />

            </PhysicalBox>

            <DetailCard
                id="mouse-over-popover"
                open={open}
                anchorEl={anchorEl}
                elementName={"TOR " + props.index}
                enabled={props.details.tor.details.enabled}
                condition={props.details.tor.details.condition}
                onClose={handlePopoverClose}
            >
                <List dense className={classes.list}>
                    {Array.from(props.details.tor.ports).map((v, k) =>
                        <ListItem dense>
                            <ListItemIcon>
                                {networkDetail(v[1].observed.smState === CableState_SM.on)}
                            </ListItemIcon>
                            <ListItemText
                                primary={
                                    "Port " + k + ": wired to " + hardware_HwTypeToString(v[1].port.item.type) +
                                    " " + v[1].port.item.id + ", port " + v[1].port.item.port + ", currently " +
                                    (v[1].port.wired ? "connected" : "disconnected") + ", with network connection " +
                                    cableState_SMToString(v[1].observed.smState)}
                            />
                        </ListItem>
                    )}
                </List>
            </DetailCard>
        </>
    )
}
