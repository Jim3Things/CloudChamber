import React from "react";
import {Check, CheckCircleOutline, Error, HighlightOff} from "@material-ui/icons";
import {createStyles, Table, TableBody, TableCell, TableHead, TableRow} from "@material-ui/core";

import {BladeDetails, InstanceState, PhysicalState} from "../proxies/InventoryProxy";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles((theme) =>
    createStyles({
        cell: {
            backgroundColor: theme.palette.action.hover
        },
    }),
);

// Return the icon that matches the instance state
function statusIcon(state: InstanceState) {
    switch (state) {
        case InstanceState.escrow:
            return <CheckCircleOutline/>

        case InstanceState.faulted:
            return <Error/>

        case InstanceState.running:
            return <Check/>

        default:
    }
}

// Return the icon that matches the physical state of a blade
function bladeStatusIcon(state: PhysicalState) {
    switch (state) {
        case PhysicalState.faulted:
            return <Error/>

        case PhysicalState.healthy:
            return <Check/>

        case PhysicalState.off:
            return <HighlightOff/>
    }
}

// Construct the details display showing the usage of a blade, its overall
// capacity, and what remains available.
export function BladeUsageDetails(props: { index: number, details: BladeDetails }) {
    const classes = useStyles();

    const totalUsed = props.details.usage.reduce((sum: number, item) => sum + item.usage, 0)
    const unused = props.details.capacity.cores - totalUsed

    return <Table size="small">
        <TableHead>
            <TableRow>
                <TableCell align="center" colSpan={6}>
                    Details for blade {props.index}
                </TableCell>
            </TableRow>
            <TableRow>
                <TableCell/>
                <TableCell>Status</TableCell>
                <TableCell>Cores</TableCell>
                <TableCell>Memory</TableCell>
                <TableCell>Disk</TableCell>
                <TableCell>NIC</TableCell>
            </TableRow>
        </TableHead>
        <TableBody>
            <TableRow>
                <TableCell className={classes.cell}>Blade Capacity</TableCell>
                <TableCell className={classes.cell}>{bladeStatusIcon(props.details.state)}</TableCell>
                <TableCell className={classes.cell}>{props.details.capacity.cores}</TableCell>
                <TableCell className={classes.cell}>{props.details.capacity.memoryInMb}</TableCell>
                <TableCell className={classes.cell}>{props.details.capacity.diskInGb}</TableCell>
                <TableCell className={classes.cell}>{props.details.capacity.networkBandwidthInMbps}</TableCell>
            </TableRow>

            {props.details.usage.map((v, k) => {
                return <TableRow>
                    <TableCell>Instance {k}</TableCell>
                    <TableCell>{statusIcon(v.state)}</TableCell>
                    <TableCell>{v.usage}</TableCell>
                    <TableCell>?</TableCell>
                    <TableCell>?</TableCell>
                    <TableCell>?</TableCell>
                </TableRow>
            })}

            <TableRow>
                <TableCell className={classes.cell}>Unused</TableCell>
                <TableCell className={classes.cell} />
                <TableCell className={classes.cell}>{unused}</TableCell>
                <TableCell className={classes.cell}>?</TableCell>
                <TableCell className={classes.cell}>?</TableCell>
                <TableCell className={classes.cell}>?</TableCell>
            </TableRow>
        </TableBody>
    </Table>
}