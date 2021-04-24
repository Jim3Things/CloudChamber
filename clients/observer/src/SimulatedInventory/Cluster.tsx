// Cluster takes the full cluster definition and emits the outer table with the
// name as the first row The second row holds the rack names, and the third row
// holds the details for each rack

import React from "react";
import {ClusterDetails} from "../proxies/InventoryProxy";
import {makeStyles} from "@material-ui/core/styles";
import {Colors} from "./SimulatedInventory";
import {Rack} from "./Rack";

const useStyles = makeStyles(() => ({
    th: {
        textAlign: "center",
        alignContent: "center",
        border: "0px",
        padding: "1px",
        fontSize: "medium"
    },
    tr: {
        border: "0px",
        padding: "1px"
    },
    td: {
        border: "3px solid darkgrey",
        padding: "1px",
        alignContent: "start",
        verticalAlign: "top",
        background: "lightgrey"
    },
    tdClusterName: {
        border: "0px",
        padding: "1px",
        alignContent: "start",
        verticalAlign: "top"
    },
    rackName: {
        textAlign: "center",
        alignContent: "center",
        fontSize: "medium",
        border: "3px solid lightgrey",
        padding: "0px",
        verticalAlign: "top"
    }
}));

export function Cluster(props: {
            cluster: ClusterDetails,
            palette: Colors
        }) {
    const classes = useStyles();

    return (
        <table>
            <tbody>
                <tr className={classes.th}>
                    <td className={classes.tdClusterName} colSpan={Math.max(props.cluster.racks.size, 1)}>{props.cluster.name}</td>
                </tr>
                <tr className={classes.tr}>
                    {Array.from(props.cluster.racks.keys()).map((name) =>
                        (
                            <td className={classes.rackName}>{name}</td>
                        ))}
                </tr>
                <tr className={classes.tr}>
                    {Array.from(props.cluster.racks.values()).map((rack) =>
                        (
                            <td className={classes.td}>
                                <Rack
                                    bladeLimit={props.cluster.maxBladeCount}
                                    capacityLimit={props.cluster.maxCapacity}
                                    rack={rack}
                                    palette={props.palette}
                                />
                            </td>
                        ))}
                </tr>
            </tbody>
        </table>
    )
}
