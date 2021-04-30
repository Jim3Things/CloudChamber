import React from "react"
import {AppBar, IconButton, List, ListItem, ListItemIcon, ListItemText, Toolbar, Tooltip} from "@material-ui/core"
import {makeStyles, Theme} from "@material-ui/core/styles"

import {Stepper} from "./Stepper"
import {SetStepperPolicy} from "./proxies/StepperProxy"
import {ExitToApp} from "@material-ui/icons"
import {Settings, SettingsState} from "./Settings"
import {SessionUser} from "./proxies/Session"
import {ExpandingLabel} from "./common/ExpandingLabel"
import {CheckIf} from "./common/If"
import {UserPublic} from "./pkg/protos/admin/users"

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        flexGrow: 1
    },
    rightIcon: {
        MarginLeft: theme.spacing(1)
    },
    nested: {
        paddingLeft: theme.spacing(4),
        pt: 0,
        pb: 0
    }
}))

// ListRight shows a single user right, with a checkbox indicating if it is
// enabled or not.  The item is display-only.
function ListRight(props: {
    cond: boolean | undefined,
    text: string
}) {
    const classes = useStyles()

    const test = Boolean(props.cond)

    return <ListItem dense className={classes.nested}>
        <ListItemIcon>
            <CheckIf cond={test}/>
        </ListItemIcon>
        <ListItemText primary={props.text}/>
    </ListItem>

}

// UserDetails displays the supplied public user attributes.
function UserDetails(props: { details: UserPublic }) {
    return (
        <List>
            <ListItem>
                <ListItemIcon>
                    <CheckIf cond={props.details.enabled}/>
                </ListItemIcon>
                <ListItemText primary="Enabled"/>
            </ListItem>
            <ListItem>
                <ListItemIcon>
                    <CheckIf cond={props.details.neverDelete}/>
                </ListItemIcon>
                <ListItemText primary="Protected"/>
            </ListItem>
            <ListItem/>

            <ListItem>
                <ListItemText primary="Rights:"/>
            </ListItem>
            <ListRight
                cond={props.details.rights?.canStepTime}
                text="Can Step Time"/>
            <ListRight
                cond={props.details.rights?.canInjectFaults}
                text="Can Inject Faults"/>
            <ListRight
                cond={props.details.rights?.canManageAccounts}
                text="Can Manage Accounts"/>
            <ListRight
                cond={props.details.rights?.canModifyInventory}
                text="Can Modify Inventory"/>
            <ListRight
                cond={props.details.rights?.canModifyWorkloads}
                text="Can Modify Workloads"/>
            <ListRight
                cond={props.details.rights?.canPerformRepairs}
                text="Can Perform Repairs"/>
        </List>
    )
}

export function CommandBar(props: {
    sessionUser: SessionUser,
    settings: SettingsState,
    onPolicyEvent: (policy: SetStepperPolicy) => void,
    onSettingsChange: (settings: SettingsState) => void,
    onLogout: () => void
}) {
    const classes = useStyles()

    const rights = props.sessionUser.details.rights
    const disableStepTime = rights !== undefined ? !rights.canStepTime : true

    return (
        <div className={classes.root}>
            <AppBar position="static">
                <Toolbar variant="dense">
                    <ExpandingLabel
                        label={props.sessionUser.name}
                        variant="subtitle2"
                    >
                        <UserDetails details={props.sessionUser.details}/>
                    </ExpandingLabel>

                    <Tooltip title="log out">
                        <IconButton
                            color="inherit"
                            className={classes.rightIcon}
                            onClick={() => props.onLogout()}
                        >
                            <ExitToApp/>
                        </IconButton>
                    </Tooltip>

                    <div className={classes.root}/>

                    <Stepper
                        disabled={disableStepTime}
                        onPolicyEvent={props.onPolicyEvent}
                    />

                    <Settings
                        settings={props.settings}
                        onChange={props.onSettingsChange}
                    />
                </Toolbar>
            </AppBar>
        </div>
    )
}
