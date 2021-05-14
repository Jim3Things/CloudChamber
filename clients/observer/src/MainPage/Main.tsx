import React from 'react'
import {CommandBar} from "../CommandBar"
import {Paper} from "@material-ui/core"
import ControllerDetails from "../ControllerDetails"
import {SimulatedInventory} from "../SimulatedInventory/SimulatedInventory"
import {ExpansionHandler, LogDisplay} from "../Log/LogDisplay"
import {StatusBar} from "../StatusBar"
import {makeStyles} from "@material-ui/core/styles"
import {Container, Item} from "../common/Cells"
import {Organizer} from "../Log/Organizer"

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1
    }
}))

function getElementHeight(id: string): number {
    const elem = document.getElementById(id)
    if (elem === null) {
        return 100
    }

    return elem.offsetHeight
}

export function MainPage(props: {
    onTrackChange: ExpansionHandler,
    onLogout: (name: string) => void,
    organizer: Organizer
}) {
    const classes = useStyles()

    return <div className={classes.root}>
        <Container>
            <Item xs={12}>
                <CommandBar
                    onLogout={props.onLogout}
                />
            </Item>
            <Item xs={9}>
                <Container id="left-pane" direction="column">
                    <Item xs={12}>
                        <Paper variant="outlined" style={{maxHeight: 150, minHeight: 150, overflow: "auto"}}>
                            <ControllerDetails/>
                        </Paper>
                    </Item>
                    <Item xs={12}>
                        <Paper variant="outlined" style={{minHeight: 250, overflow: "auto"}}>
                            <SimulatedInventory />
                        </Paper>
                    </Item>
                </Container>
            </Item>
            <Item xs={3}>
                <LogDisplay
                    height={getElementHeight("left-pane")}
                    organizer={props.organizer}
                    onTrackChange={props.onTrackChange}
                />
            </Item>
            <Item xs={12}>
                <Paper variant="outlined">
                    <StatusBar />
                </Paper>
            </Item>
        </Container>
    </div>
}
