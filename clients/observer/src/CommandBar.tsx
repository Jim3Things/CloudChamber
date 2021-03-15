import React, {FunctionComponent} from "react";
import {AppBar, IconButton, Tab, Tabs, Toolbar, Tooltip, Typography} from "@material-ui/core";
import {makeStyles} from "@material-ui/core/styles";

import {Stepper} from "./Stepper";
import {SetStepperPolicy} from "./proxies/StepperProxy";
import {ExitToApp} from "@material-ui/icons";
import {Settings, SettingsState} from "./Settings";

export enum CommandTab {
    Admin,
    Workloads,
    Inventory,
    Faults,
    Help
}

const useStyles = makeStyles(() => ({
    root: {
        flexGrow: 1
    }
}));

// This method creates a function to handle the command bar that drives the
// user command choices in Cloud Chamber.  This is a stateless component that
// assumes that the current tab prop will be updated in response to a
// onCommandSelect notification.

export const CommandBar: FunctionComponent<{
            tab: CommandTab,
            sessionUser: string,
            settings: SettingsState,
            onPolicyEvent: (policy: SetStepperPolicy) => void,
            onCommandSelect: (tab: CommandTab) => void,
            onSettingsChange: (settings: SettingsState) => void,
            onLogout: () => void
}> = ({
    tab,
    sessionUser,
    settings,
    onPolicyEvent,
    onCommandSelect,
    onSettingsChange,
    onLogout
}) => {
    const classes = useStyles();

    // Helper to decorate the tabs with unique IDs
    const tabProps = (index: any) => {
        return {
            id: `simple-tab-${index}`,
            'aria-controls': `simple-tabpanel-${index}`,
        };
    }

    // Callback helper to notify that a tab has been selected.
    const notify = (event: React.ChangeEvent<{}>, newValue: number) => {
        // Map the number to the command tab enum
        const tab: CommandTab = newValue;

        if (onCommandSelect) {
            onCommandSelect(tab)
        }
    }

    return (
        <div className={classes.root}>
            <AppBar position="static">
                <Toolbar variant="dense">
                    <Tabs value={tab} onChange={notify}>
                        <Tab wrapped label="Admin" {...tabProps(0)}/>
                        <Tab wrapped label="Workloads" {...tabProps(1)}/>
                        <Tab wrapped label="Inventory" {...tabProps(2)}/>
                        <Tab wrapped label="Faults" {...tabProps(3)}/>
                        <Tab wrapped label="Help" {...tabProps(4)}/>
                    </Tabs>
                    <div className={classes.root}/>
                    <Typography variant="subtitle2">
                        {sessionUser}&nbsp;
                    </Typography>
                    <Tooltip title="log out">
                        <IconButton
                            color="inherit"
                            onClick={() => onLogout()}>
                                <ExitToApp/>
                        </IconButton>
                    </Tooltip>
                    <Stepper onPolicyEvent={onPolicyEvent}/>
                    <Settings
                        settings={settings}
                        onChange={onSettingsChange}
                    />
                </Toolbar>
            </AppBar>
        </div>
    );
}
