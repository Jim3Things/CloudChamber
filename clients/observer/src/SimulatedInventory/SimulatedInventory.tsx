import React, {useEffect, useState} from 'react';
import {green, grey, red, yellow} from "@material-ui/core/colors";

import {ClusterDetails, InventoryProxy, RackDetails} from "../proxies/InventoryProxy";
import {Cluster} from "./Cluster";
import {ErrorSnackbar, MessageMode, SnackData, SuccessSnackbar} from "../common/Snackbar";


// This is the palette of colors used by the various parts of the
// cluster, based on physical state, as well as instance state
export interface Colors {
    backgroundColor: string,
    escrowColor: string,
    runningColor: string,
    faultedColor: string,
    offColor: string,
    freeColor: string,
    illegal: string
}

// Draw the simulated inventory
export function SimulatedInventory(props: {proxy: InventoryProxy}) {
    const [cluster, setCluster] = useState<ClusterDetails>({
        name: "Loading...",
        maxCapacity: {
            cores: 0,
            diskInGb: 0,
            memoryInMb: 0,
            networkBandwidthInMbps: 0,
            arch: "",
            accelerators: [],
        },
        maxBladeCount: 1,
        racks: new Map<string, RackDetails>()
    })

    const [snackData, setSnackData] = useState<SnackData>({
        message: "",
        mode: MessageMode.None
    })

    // Start by getting a snapshot of the cluster's inventory
    useEffect(() =>{
        props.proxy.getCluster()
            .then((zone) => {
                setCluster(zone)

                // Now start getting each rack
                zone.racks.forEach((rack, name) => {
                    props.proxy.getRackDetails(rack)
                        .then(res => {
                            let newZone = {...zone}
                            newZone.racks.set(name, res)

                            let done = true
                            newZone.racks.forEach((rack) => {
                                if (!rack.detailsLoaded) {
                                    done = false
                                }
                            })

                            setCluster(zone)
                            setSnackData({
                                message: done ? "Inventory successfully loaded" : "",
                                mode: done ? MessageMode.Success : MessageMode.None})
                        })
                })
            })
            .catch((err: Error) => {
                setSnackData({ message: err.message, mode: MessageMode.Error })
            })
    }, [props.proxy])

    const palette: Colors = {
        backgroundColor: grey[100],
        escrowColor: yellow[800],
        faultedColor: red[900],
        freeColor: grey[100],
        illegal: red.A400,
        offColor: grey[400],
        runningColor: green[300]
    }

    return <React.Fragment>
        <Cluster cluster={cluster} palette={palette}/>

        <SuccessSnackbar
            open={snackData.mode === MessageMode.Success}
            onClose={() => setSnackData({ message: "", mode: MessageMode.None})}
            autoHideDuration={3000}
            message={snackData.message} />

        <ErrorSnackbar
            open={snackData.mode === MessageMode.Error}
            onClose={() => setSnackData({ message: "", mode: MessageMode.None})}
            autoHideDuration={4000}
            message={snackData.message} />

    </React.Fragment>
}
