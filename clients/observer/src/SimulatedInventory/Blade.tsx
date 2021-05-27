import React from "react"
import {grey} from "@material-ui/core/colors"
import {createStyles, Popover, Tooltip} from "@material-ui/core"
import {makeStyles} from "@material-ui/core/styles"

import {BladeDescription, InstanceDetails, InstanceState, PhysicalState} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {BladeUsageDetails} from "./BladeUsageDetails"
import {BladeCapacity} from "../pkg/protos/inventory/capacity"
import {Computer} from "@material-ui/icons"
import {BladeSmState} from "../pkg/protos/inventory/common"

const useStyles = makeStyles((theme) =>
    createStyles({
        popover: {
            pointerEvents: 'none'
        },
        paper: {
            padding: theme.spacing(1),
        },
    }),
)

// +++ Detail SVGRect boundary calculations

interface detailBox {
    left: number
    width: number
    color: string
    opacity: number
}

// Determine the color for the workload instance, based on its running
// state
function statusToColor(state: InstanceState, palette: Colors): string {
    switch (state) {
        case InstanceState.escrow:
            return palette.escrowColor
        case InstanceState.running:
            return palette.runningColor
        case InstanceState.faulted:
            return palette.faultedColor
        default:
            return palette.illegal
    }
}

// toPhysical is a transitional conversion function to handle the
// partial removal of PhysicalState. It converts a blade SM state
// into the closest physical state approximation.
function toPhysical(s: BladeSmState): PhysicalState {
    switch (s) {
        case BladeSmState.faulted:
            return PhysicalState.faulted

        case BladeSmState.off_disconnected:
        case BladeSmState.off_connected:
            return PhysicalState.off

        default:
            return PhysicalState.healthy
    }
}

// Construct the details needed to place the usage rectangles
function formBladeDetailBoxes(
    instances: InstanceDetails[],
    capacity: BladeCapacity,
    bladeWidth: number,
    boundingState: BladeSmState,
    palette: Colors): detailBox[] {

    let set: detailBox[] = []
    let left = 0

    // Construct the left, width, and fill details for each current workload
    for (const item of instances) {
        const percent = item.usage / capacity.cores
        const pixelsUsed = Math.round(bladeWidth * percent)
        const width = Math.min(pixelsUsed + left, bladeWidth) - left

        set.push({
            left: left,
            width: width,
            color: statusToColor(item.state, palette),
            opacity: Opacity(toPhysical(boundingState))
        })

        left += width
    }

    if (left !== bladeWidth) {
        // Not all the capacity is used, fill out the rest as unused space
        set.push({
            left: left,
            width: bladeWidth - left,
            color: palette.freeColor,
            opacity: Opacity(toPhysical(boundingState))
        })
    }

    return set
}

// --- Detail SVGRect boundary calculations

// This method draws a blade, accounting for its state and usage
export function Blade(props: {
    x: number,
    y: number,
    width: number,
    height: number,
    index: number,
    details: BladeDescription,
    limits: BladeCapacity,
    palette: Colors
}) {
    const classes = useStyles()

    const [anchorEl, setAnchorEl] = React.useState<SVGSVGElement | null>(null)

    const handlePopoverOpen = (event: React.MouseEvent<SVGSVGElement, MouseEvent>): void => {
        setAnchorEl(event.currentTarget)
    }

    const handlePopoverClose = () => {
        setAnchorEl(null)
    }

    const open = Boolean(anchorEl)

    const iconWidth = Math.min(props.height, 50)
    const offset = iconWidth + 5

    const bladeWidth = props.width - offset

    const frameWidth = bladeWidth * props.details.blade.capacity.cores / props.limits.cores

    // Construct the inner box width boundaries
    const boxes = formBladeDetailBoxes(
        props.details.usage,
        props.details.blade.capacity,
        frameWidth - 4,
        props.details.blade.observed.smState,
        props.palette)

    // Draw the blade, filling in the instance usage and state
    return (
        <React.Fragment>
            <Tooltip title={"Blade " + props.index}>
                <Computer
                    x={props.x}
                    y={props.y}
                    width={iconWidth}
                    height={props.height}/>
            </Tooltip>

            <PhysicalBox
                x={props.x + offset}
                y={props.y}
                width={frameWidth}
                height={props.height}
                state={toPhysical(props.details.blade.observed.smState)}
                palette={props.palette}
                pointerEvents="all"
                aria-owns={open ? 'mouse-over-popover' : undefined}
                aria-haspopup="true"
                onMouseEnter={handlePopoverOpen}
                onMouseLeave={handlePopoverClose}
            >
                {boxes.map((value) => {
                    return <rect
                        x={2 + value.left}
                        y={2}
                        height={props.height - 4}
                        width={value.width}
                        fill={value.color}
                        strokeWidth={1}
                        stroke={grey[700]}
                        fillOpacity={value.opacity}
                    />
                })}
            </PhysicalBox>

            <Popover
                id="mouse-over-popover"
                className={classes.popover}
                classes={{
                    paper: classes.paper,
                }}
                open={open}
                anchorEl={anchorEl}
                anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                }}
                transformOrigin={{
                    vertical: 'top',
                    horizontal: 'left',
                }}
                onClose={handlePopoverClose}
                disableRestoreFocus
            >
                <BladeUsageDetails
                    index={props.index}
                    details={props.details}/>
            </Popover>

        </React.Fragment>
    )
}
