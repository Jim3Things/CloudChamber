import React from 'react';
import {CommandBar, CommandTab} from "../CommandBar";
import {StepperPolicy, Timestamp} from "../proxies/StepperProxy";
import {UsersProxy} from "../proxies/UsersProxy";
import {InventoryProxy} from "../proxies/InventoryProxy";
import {Paper} from "@material-ui/core";
import {CommandArea} from "../CommandArea/CommandArea";
import ControllerDetails from "../ControllerDetails";
import {SimulatedInventory} from "../SimulatedInventory/SimulatedInventory";
import {ExpansionHandler, LogDisplay} from "../Log/LogDisplay";
import {StatusBar} from "../StatusBar";
import {makeStyles} from "@material-ui/core/styles";
import {Container, Item} from "../common/Cells";
import {Organizer} from "../Log/Organizer";
import {SettingsState} from "../Settings";

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1
    }
}));

function getElementHeight(id: string) : number {
    const elem = document.getElementById(id);
    if (elem === null) {
        return 100
    }

    return elem.offsetHeight;
}

export function MainPage(props: {
            tab: CommandTab,
            activeSession: boolean,
            sessionUser: string,
            settings: SettingsState,
            onPolicyEvent: (policy: StepperPolicy) => void,
            onCommandSelect: (tab: CommandTab) => void,
            onSettingsChange: (settings: SettingsState) => void,
            onTrackChange: ExpansionHandler,
            onLogout: () => void,
            usersProxy: UsersProxy,
            proxy: InventoryProxy,
            cur: Timestamp,
            organizer: Organizer}) {
    const classes = useStyles();

    return <div className={classes.root}>
        <Container>
            <Item xs={12}>
                <CommandBar
                    tab={props.tab}
                    sessionUser={props.sessionUser}
                    settings={props.settings}
                    onCommandSelect={props.onCommandSelect}
                    onPolicyEvent={props.onPolicyEvent}
                    onSettingsChange={props.onSettingsChange}
                    onLogout={props.onLogout}
                />
            </Item>
            <Item xs={9}>
                <Container id="left-pane" direction="column">
                    <Item xs={12}>
                        <Paper variant="outlined">
                            <CommandArea
                                sessionUser={props.sessionUser}
                                usersProxy={props.usersProxy}
                                tab={props.tab}/>
                        </Paper>
                    </Item>
                    <Item xs={12}>
                        <Paper variant="outlined" style={{maxHeight: 100, minHeight: 100, overflow: "auto"}}>
                            <ControllerDetails/>
                        </Paper>
                    </Item>
                    <Item xs={12}>
                        <Paper variant="outlined" style={{minHeight: 200, overflow: "auto"}}>
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
    </div>;
}