// Draw a PDU component, with the currently defined connections.
//
// Connections that are 'true' indicate that the equivalent blade
// should be getting power; those that are 'false' indicate that
// the equivalent blade should also be either powered off or faulted.
// (powered off, if the 'false' state is intentional, faulted if it
// is not)

import React from "react"
import {PduDescription, PhysicalState} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {Connectors} from "./Connectors"
import {Power} from '@material-ui/icons'
import {
    createStyles, List, ListItem, ListItemIcon, ListItemText,
    Tooltip
} from "@material-ui/core"
import {
    CableState_SM, cableState_SMToString,
    hardware_HwTypeToString,
    PduState_SM
} from "../pkg/protos/inventory/common"
import {makeStyles} from "@material-ui/core/styles"
import {usePopup} from "../common/UsePopup"
import {DetailCard} from "./DetailCard"
import {PowerOffIcon, PowerOnIcon} from "../common/Icons"

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
function toPhysical(s: PduState_SM): PhysicalState {
    switch (s) {
        case PduState_SM.UNRECOGNIZED:
        case PduState_SM.stuck:
        case PduState_SM.invalid:
            return PhysicalState.faulted

        default:
            return PhysicalState.healthy
    }
}

function powerDetail(onOff: boolean) {
    if (onOff) {
        return <PowerOnIcon />
    } else {
        return <PowerOffIcon />
    }
}

export function PDU(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    details: PduDescription,
    palette: Colors,
    index: number
}) {
    const classes = useStyles()

    const [open, anchorEl, handlePopoverOpen, handlePopoverClose] = usePopup<SVGSVGElement>()

    const iconWidth = Math.min(props.height, 50)
    const offset = iconWidth + 5
    const connectionWidth = props.width - offset

    const state = toPhysical(props.details.pdu.observed.smState)

    const wires = Array.from(props.details.pdu.ports).map(v =>
        (state === PhysicalState.healthy) &&
        (v[1].observed.smState === CableState_SM.on))

    return (
        <>
            <Tooltip title={"PDU " + props.index}>
                <Power
                    x={props.x}
                    y={props.y}
                    width={iconWidth}
                    height={props.height}
                />
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
                elementName={"PDU " + props.index}
                enabled={props.details.pdu.details.enabled}
                condition={props.details.pdu.details.condition}
                onClose={handlePopoverClose}
            >
                <List dense className={classes.list}>
                    {Array.from(props.details.pdu.ports).map((v, k) =>
                        <ListItem dense>
                            <ListItemIcon>
                                {powerDetail(v[1].observed.smState === CableState_SM.on)}
                            </ListItemIcon>
                            <ListItemText
                                primary={
                                    "Port " + k + ": wired to " + hardware_HwTypeToString(v[1].port.item.type) +
                                    " " + v[1].port.item.id + ", port " + v[1].port.item.port + ", currently " +
                                    (v[1].port.wired ? "connected" : "disconnected") + ", with power " +
                                    cableState_SMToString(v[1].observed.smState)}
                            />
                        </ListItem>
                    )}
                </List>
            </DetailCard>
        </>
    )
}
