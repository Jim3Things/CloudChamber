// Cluster takes the full cluster definition and emits the outer table with the
// name as the first row The second row holds the rack names, and the third row
// holds the details for each rack

import React from "react"
import {ClusterDetails} from "../proxies/InventoryProxy"
import {makeStyles} from "@material-ui/core/styles"
import {Colors} from "./SimulatedInventory"
import {Rack} from "./Rack"
import {Container, Item} from "../common/Cells"
import {Typography} from "@material-ui/core"

const useStyles = makeStyles((theme) => ({
    th: {
        borderBottom: "5px solid darkgrey",
        padding: "1px"
    },
    tr: {
        border: "0px",
        padding: theme.spacing(1),
        background: theme.palette.background.paper,
    },
    td: {
        border: "3px",
        borderColor: theme.palette.background.paper,
        paddingRight: theme.spacing(1),
        paddingLeft: theme.spacing(1),
        background: theme.palette.background.paper,
    },
    tdClusterName: {
        border: "0px",
        alignContent: "start",
        verticalAlign: "top",
        textAlign: "center",
    },
    tdClusterLocation: {
        border: "0px",
        alignContent: "start",
        verticalAlign: "top",
        textAlign: "right",
        fontStyle: "italic",
        paddingRight: theme.spacing(1),
    },
    rackName: {
        textAlign: "center",
        alignContent: "center",
        align: "center",
        backgroundColor: theme.palette.background.paper,
    }
}))

export function Cluster(props: {
    cluster: ClusterDetails,
    palette: Colors
}) {
    const classes = useStyles()

    return (
        <Container xs={12}>
            <Container xs={12} className={classes.th}>
                <Item xs={12} className={classes.tdClusterName}>
                    <Typography variant="h4">
                        {props.cluster.name}
                    </Typography>
                </Item>
                <Item xs={12} className={classes.tdClusterLocation}>
                    <Typography variant="subtitle2">
                        At: {props.cluster.location}
                    </Typography>
                </Item>
            </Container>
            <Container xs={12} className={classes.tr}>
                {Array.from(props.cluster.racks.entries()).map(([name, value]) => (
                    <Item className={classes.td}>
                        <div>
                            <Typography variant="h5" className={classes.rackName}>
                                {name}
                            </Typography>
                        </div>

                        <Rack
                            bladeLimit={props.cluster.maxBladeCount}
                            capacityLimit={props.cluster.maxCapacity}
                            rack={value}
                            palette={props.palette}
                        />
                    </Item>
                ))}
            </Container>
        </Container>
    )
}
