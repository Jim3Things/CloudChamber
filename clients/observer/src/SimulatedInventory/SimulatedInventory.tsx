import React, {Component} from 'react';
import {green, grey, red, yellow} from "@material-ui/core/colors";

import {InventoryProxy, ClusterDetails, RackDetails} from "../proxies/InventoryProxy";
import {Cluster} from "./Cluster";
import {SuccessSnackbar} from "../common/SuccessSnackbar";
import {ErrorSnackbar} from "../common/ErrorSnackbar";

interface Props {
    proxy: InventoryProxy
}

enum MessageMode {
    None = 0,                   // Show no snackbar
    Success = 1,                // Show the success snackbar
    Error = 2                   // Show the error snackbar
}

interface State {
    cluster: ClusterDetails,

    snackMode: MessageMode      // Which snackbar to display, if any
    snackText: string           // ... and the text to supply
}

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
export class SimulatedInventory extends Component<Props, State> {
    state: State = {
        cluster: {
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
        },
        snackMode: MessageMode.None,
        snackText: ""
    }

    // Start by getting a snapshot of the cluster's inventory
    componentDidMount() {
        this.props.proxy.getCluster()
            .then((zone) => {
                this.setState({  cluster: zone })

                // Now start getting each rack
                zone.racks.forEach((rack, name) => {
                    this.props.proxy.getRackDetails(rack)
                        .then(res => {
                            let newZone = {...zone}
                            newZone.racks.set(name, res)

                            let done = true
                            newZone.racks.forEach((rack) => {
                                if (!rack.detailsLoaded) {
                                    done = false
                                }
                            })

                            this.setState({
                                cluster: zone,

                                snackMode: done ? MessageMode.Success : MessageMode.None,
                                snackText: done ? "Inventory successfully loaded" : ""
                            })
                        })
                })
            })
            .catch((err: Error) => {
                this.setState({
                    snackMode: MessageMode.Error,
                    snackText: err.message
                })
            })
    }

    render() {
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
            <Cluster cluster={this.state.cluster} palette={palette}/>

            <SuccessSnackbar
                open={this.state.snackMode === MessageMode.Success}
                onClose={() => this.setState({snackMode: MessageMode.None})}
                autoHideDuration={3000}
                message={this.state.snackText} />

            <ErrorSnackbar
                open={this.state.snackMode === MessageMode.Error}
                onClose={() => this.setState({snackMode: MessageMode.None})}
                autoHideDuration={4000}
                message={this.state.snackText} />

        </React.Fragment>
    }
}
