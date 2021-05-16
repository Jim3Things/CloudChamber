import React from 'react'
import {CommandBar} from "../CommandBar"
import {Paper} from "@material-ui/core"
import ControllerDetails from "../ControllerDetails"
import {SimulatedInventory} from "../SimulatedInventory/SimulatedInventory"
import {LogDisplay} from "../Log/LogDisplay"
import {StatusBar} from "../StatusBar"
import {makeStyles} from "@material-ui/core/styles"
import {Container, Item} from "../common/Cells"

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1
    }
}))

export function MainPage(props: {
    onLogout: (name: string) => void
}) {
    const classes = useStyles()

    return <div className={classes.root}>
        <Container>
            <Item xs={12}>
                <CommandBar onLogout={props.onLogout}/>
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
                            <SimulatedInventory/>
                        </Paper>
                    </Item>
                </Container>
            </Item>
            <Item xs={3}>
                <LogDisplay matchId="left-pane"/>
            </Item>
            <Item xs={12}>
                <Paper variant="outlined">
                    <StatusBar/>
                </Paper>
            </Item>
        </Container>
    </div>
}
