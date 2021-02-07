import React from 'react';
import {makeStyles} from "@material-ui/core/styles";

import {CommandTab} from "../CommandBar";
import {UsersProxy} from "../proxies/UsersProxy";
import {AdminPanel} from "./AdminPanel";
import {WorkloadsPanel} from "./WorkloadsPanel";
import {InventoryPanel} from "./InventoryPanel";
import {FaultInjectionPanel} from "./FaultInjectionPanel";
import {HelpPanel} from "./HelpPanel";
import {Paper} from "@material-ui/core";
import {HideIf} from "../common/If";

const useStyles = makeStyles((theme) => ({
    root: {
        maxHeight: 450,
        minHeight: 300
    }
}));

// Command subset panel details

// This is the holding area for the command subset panels.  Since it has no
// state and no complexity, it does not use a full component class.  It does
// define a fixed size for the command area that all panels must respect...
export function CommandArea(props: {
        tab: CommandTab,
        usersProxy: UsersProxy,
        sessionUser: string
    }) {
    const classes = useStyles();

    return (
        <Paper className={classes.root}>
            <HideIf cond={props.tab !== CommandTab.Admin}>
                 <AdminPanel
                     height={300}
                     usersProxy={props.usersProxy}
                     sessionUser={props.sessionUser}/>
            </HideIf>

            <HideIf cond={props.tab !== CommandTab.Workloads}>
                 <WorkloadsPanel />
            </HideIf>

            <HideIf cond={props.tab !== CommandTab.Inventory}>
                 <InventoryPanel />
            </HideIf>

            <HideIf cond={props.tab !== CommandTab.Faults}>
                 <FaultInjectionPanel />
            </HideIf>

            <HideIf cond={props.tab !== CommandTab.Help}>
                <HelpPanel />
            </HideIf>
        </Paper>
    );
}
