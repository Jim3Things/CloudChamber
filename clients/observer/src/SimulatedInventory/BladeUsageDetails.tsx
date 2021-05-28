import React from "react"
import {Check, CheckCircleOutline, Error, HighlightOff, Warning} from "@material-ui/icons"
import {
    Card,
    CardContent,
    CardHeader,
    createStyles,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Typography,
} from "@material-ui/core"

import {BladeDescription, InstanceState} from "../proxies/InventoryProxy"
import {makeStyles, Theme} from "@material-ui/core/styles"
import {
    BladeBootInfo_Method,
    BladeSmState,
    bladeSmStateToString,
    conditionToString
} from "../pkg/protos/inventory/common"
import {Accelerator} from "../pkg/protos/inventory/capacity"

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        root: {
            maxWidth: 600,
        },
        cell: {
            backgroundColor: theme.palette.action.hover
        },
        noBorder: {
            borderStyle: "none",
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
function bladeStatusIcon(state: BladeSmState) {
    switch (state) {
        case BladeSmState.faulted:
            return <Error/>

        case BladeSmState.working:
            return <Check/>

        case BladeSmState.off_disconnected:
        case BladeSmState.off_connected:
            return <HighlightOff/>

        default:
            return <Warning/>
    }
}

function acceleratorText(accels: Accelerator[]) {
    const accel = accels.join(", ")

    if (accel.length > 0) {
        return accel
    }

    return "none"
}

function bootSourceText(source: BladeBootInfo_Method): string {
    switch (source) {
        case BladeBootInfo_Method.local:
            return "local"
        case BladeBootInfo_Method.network:
            return "network"
        default:
            return "unknown"
    }
}

// Construct the details display showing the usage of a blade, its overall
// capacity, and what remains available.
export function BladeUsageDetails(props: { index: number, details: BladeDescription }) {
    const classes = useStyles();

    const totalUsed = props.details.usage.reduce((sum: number, item) => sum + item.usage, 0)
    const unused = props.details.blade.capacity.cores - totalUsed
    const blade = props.details.blade

    const accelCount = blade.capacity.accelerators.length
    const accelText = "It has " + (accelCount === 0 ? "no" : accelCount) + " accelerator" +
        (accelCount !== 1 ? "s" : "") +
        (accelCount > 0 ? ", of type " + acceleratorText(blade.capacity.accelerators) : "")

    return <Card className={classes.root}>
        <CardHeader
            title={"Details for blade " + props.index}>
        </CardHeader>
        <CardContent>
            <Typography paragraph>
                This blades uses processor architecture {blade.capacity.arch}. {accelText}.
                It is {blade.details.enabled ? "enabled" : "disabled"} for use, and is {conditionToString(blade.details.condition)}.
            </Typography>
            <Typography paragraph>
                It is configured to {blade.bootOnPowerOn ? "" : "not "}automatically boot when powered on.
                When booting, it uses the '{blade.bootInfo.image}' image at version '{blade.bootInfo.version}
                ' from '{bootSourceText(blade.bootInfo.source)}' storage, with parameters '{blade.bootInfo.parameters}'.
            </Typography>
            <Typography paragraph>
                At simulated time {blade.observed.at},
                it was in the {bladeSmStateToString(blade.observed.smState)} state, which it entered at simulated time {blade.observed.enteredAt}.
            </Typography>

            <Table size="small">
                <TableHead>
                    <TableRow>
                        <TableCell align="center" colSpan={6}>
                            Usage Details
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
                        <TableCell className={classes.cell}>{bladeStatusIcon(blade.observed.smState)}</TableCell>
                        <TableCell className={classes.cell}>{blade.capacity.cores}</TableCell>
                        <TableCell className={classes.cell}>{blade.capacity.memoryInMb}</TableCell>
                        <TableCell className={classes.cell}>{blade.capacity.diskInGb}</TableCell>
                        <TableCell className={classes.cell}>{blade.capacity.networkBandwidthInMbps}</TableCell>
                    </TableRow>

                    {props.details.usage.map((v, k) =>
                         <TableRow>
                            <TableCell>Instance {k}</TableCell>
                            <TableCell>{statusIcon(v.state)}</TableCell>
                            <TableCell>{v.usage}</TableCell>
                            <TableCell>?</TableCell>
                            <TableCell>?</TableCell>
                            <TableCell>?</TableCell>
                        </TableRow>
                    )}

                    <TableRow>
                        <TableCell className={classes.cell}>Unused</TableCell>
                        <TableCell className={classes.cell}/>
                        <TableCell className={classes.cell}>{unused}</TableCell>
                        <TableCell className={classes.cell}>?</TableCell>
                        <TableCell className={classes.cell}>?</TableCell>
                        <TableCell className={classes.cell}>?</TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </CardContent>
    </Card>
}
