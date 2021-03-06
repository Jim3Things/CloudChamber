// This module contains custom icons used by the Observer display

import { SvgIcon } from "@material-ui/core"
import React from "react"

export function TorIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <path fill="currentColor"
            d="M17,3A2,2 0 0,1 19,5V15A2,2 0 0,1 17,17H13V19H14A1,1 0 0,1 15,20H22V22H15A1,1 0 0,1 14,23H10A1,1 0 0,1 9,22H2V20H9A1,1 0 0,1 10,19H11V17H7C5.89,17 5,16.1 5,15V5A2,2 0 0,1 7,3H17Z" />
    </SvgIcon>
}

export function NetworkOnIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <path
            fill="currentColor"
            d="M15,20A1,1 0 0,0 14,19H13V17H17A2,2 0 0,0 19,15V5A2,2 0 0,0 17,3H7A2,2 0 0,0 5,5V15A2,2 0 0,0 7,17H11V19H10A1,1 0 0,0 9,20H2V22H9A1,1 0 0,0 10,23H14A1,1 0 0,0 15,22H22V20H15M7,15V5H17V15H7Z" />
    </SvgIcon>
}

export function NetworkOffIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <path
            fill="currentColor"
            d="M1.04,5.27L5,9.23V15A2,2 0 0,0 7,17H11V19H10A1,1 0 0,0 9,20H2V22H9A1,1 0 0,0 10,23H14A1,1 0 0,0 15,22H17.77L19.77,24L21.04,22.72L2.32,4L1.04,5.27M7,11.23L10.77,15H7V11.23M15,20A1,1 0 0,0 14,19H13V17.23L15.77,20H15M22,20V21.14L20.86,20H22M7,6.14L5.14,4.28C5.43,3.53 6.16,3 7,3H17A2,2 0 0,1 19,5V15C19,15.85 18.47,16.57 17.72,16.86L15.86,15H17V5H7V6.14Z" />
    </SvgIcon>
}

export function PowerOnIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <path
            fill="currentColor"
            d="M16 7V3H14V7H10V3H8V7C7 7 6 8 6 9V14.5L9.5 18V21H14.5V18L18 14.5V9C18 8 17 7 16 7M16 13.67L13.09 16.59L12.67 17H11.33L10.92 16.59L8 13.67V9.09C8 9.06 8.06 9 8.09 9H15.92C15.95 9 16 9.06 16 9.09V13.67Z" />
    </SvgIcon>
}

export function PowerOffIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <path
            fill="currentColor"
            d="M22.11 21.46L2.39 1.73L1.11 3L6.25 8.14C6.1 8.41 6 8.7 6 9V14.5L9.5 18V21H14.5V18L15.31 17.2L20.84 22.73L22.11 21.46M13.09 16.59L12.67 17H11.33L10.92 16.59L8 13.67V9.89L13.89 15.78L13.09 16.59M12.2 9L10.2 7H14V3H16V7C17 7 18 8 18 9V14.5L17.85 14.65L16 12.8V9.09C16 9.06 15.95 9 15.92 9H12.2M10 6.8L8 4.8V3H10V6.8Z" />
    </SvgIcon>
}

export function LogicIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
        <   path
            fill="currentColor"
            d="M6.91 5.5L9.21 7.79L7.79 9.21L5.5 6.91L3.21 9.21L1.79 7.79L4.09 5.5L1.79 3.21L3.21 1.79L5.5 4.09L7.79 1.79L9.21 3.21M22.21 16.21L20.79 14.79L18.5 17.09L16.21 14.79L14.79 16.21L17.09 18.5L14.79 20.79L16.21 22.21L18.5 19.91L20.79 22.21L22.21 20.79L19.91 18.5M20.4 6.83L17.18 11L15.6 9.73L16.77 8.23A9.08 9.08 0 0 0 10.11 13.85A4.5 4.5 0 1 1 7.5 13A4 4 0 0 1 8.28 13.08A11.27 11.27 0 0 1 16.43 6.26L15 5.18L16.27 3.6M10 17.5A2.5 2.5 0 1 0 7.5 20A2.5 2.5 0 0 0 10 17.5Z" />
    </SvgIcon>
}

export function StorageIcon(props: {
    x?: number,
    y?: number,
    width?: number,
    height?: number
} = { x: 0, y: 0, width: 24, height: 24 }) {
    return <SvgIcon
        x={props.x}
        y={props.y}
        width={props.width}
        height={props.height}
        viewBox="0 0 24 24">
<path fill="currentColor" d="M12,3C7.58,3 4,4.79 4,7C4,9.21 7.58,11 12,11C16.42,11 20,9.21 20,7C20,4.79 16.42,3 12,3M4,9V12C4,14.21 7.58,16 12,16C16.42,16 20,14.21 20,12V9C20,11.21 16.42,13 12,13C7.58,13 4,11.21 4,9M4,14V17C4,19.21 7.58,21 12,21C16.42,21 20,19.21 20,17V14C20,16.21 16.42,18 12,18C7.58,18 4,16.21 4,14Z" />
</SvgIcon>
}
