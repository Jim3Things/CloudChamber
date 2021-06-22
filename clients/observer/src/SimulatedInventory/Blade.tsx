import React from "react"
import {grey} from "@material-ui/core/colors"
import {Tooltip} from "@material-ui/core"

import {BladeDescription, InstanceDetails, InstanceState, PhysicalState} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Opacity, PhysicalBox} from "./PhysicalBox"
import {BladeUsageDetails} from "./BladeUsageDetails"
import {BladeCapacity} from "../pkg/protos/inventory/capacity"
import {Computer} from "@material-ui/icons"
import {BladeState_SM} from "../pkg/protos/inventory/common"
import {usePopup} from "../common/UsePopup"
import {DetailCard} from "./DetailCard"

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
function toPhysical(s: BladeState_SM): PhysicalState {
    switch (s) {
        case BladeState_SM.faulted:
            return PhysicalState.faulted

        case BladeState_SM.off_disconnected:
        case BladeState_SM.off_connected:
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
    boundingState: BladeState_SM,
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
    const [open, anchorEl, handlePopoverOpen, handlePopoverClose] = usePopup<SVGSVGElement>()

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
        <>
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

            <DetailCard
                id="mouse-over-popover"
                open={open}
                anchorEl={anchorEl}
                onClose={handlePopoverClose}
                elementName={"Blade " + props.index}
                enabled={props.details.blade.details.enabled}
                condition={props.details.blade.details.condition}
            >
                <BladeUsageDetails details={props.details} />
            </DetailCard>

        </>
    )
}
