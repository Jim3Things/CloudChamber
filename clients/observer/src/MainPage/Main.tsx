import React from 'react'
import {CommandBar} from "../CommandBar"
import {SetStepperPolicy, TimeContext} from "../proxies/StepperProxy"
import {InventoryProxy} from "../proxies/InventoryProxy"
import {Paper} from "@material-ui/core"
import ControllerDetails from "../ControllerDetails"
import {SimulatedInventory} from "../SimulatedInventory/SimulatedInventory"
import {ExpansionHandler, LogDisplay} from "../Log/LogDisplay"
import {StatusBar} from "../StatusBar"
import {makeStyles} from "@material-ui/core/styles"
import {Container, Item} from "../common/Cells"
import {Organizer} from "../Log/Organizer"
import {SettingsState} from "../Settings"
import {SessionUser} from "../proxies/Session"

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
    activeSession: boolean,
    sessionUser: SessionUser,
    settings: SettingsState,
    onPolicyEvent: (policy: SetStepperPolicy) => void,
    onSettingsChange: (settings: SettingsState) => void,
    onTrackChange: ExpansionHandler,
    onLogout: () => void,
    proxy: InventoryProxy,
    cur: TimeContext,
    organizer: Organizer
}) {
    const classes = useStyles()

    return <div className={classes.root}>
        <Container>
            <Item xs={12}>
                <CommandBar
                    sessionUser={props.sessionUser}
                    settings={props.settings}
                    onPolicyEvent={props.onPolicyEvent}
                    onSettingsChange={props.onSettingsChange}
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
                            <SimulatedInventory proxy={props.proxy}/>
                        </Paper>
                    </Item>
                </Container>
            </Item>
            <Item xs={3}>
                <LogDisplay
                    height={getElementHeight("left-pane")}
                    settings={props.settings}
                    organizer={props.organizer}
                    onTrackChange={props.onTrackChange}
                />
            </Item>
            <Item xs={12}>
                <Paper variant="outlined">
                    <StatusBar cur={props.cur}/>
                </Paper>
            </Item>
        </Container>
    </div>
}
