import React from "react"
import {AppBar, IconButton, List, ListItem, ListItemIcon, ListItemText, Toolbar, Tooltip} from "@material-ui/core"
import {makeStyles, Theme} from "@material-ui/core/styles"

import {Stepper} from "./Stepper"
import {ExitToApp} from "@material-ui/icons"
import {Settings} from "./Settings"
import {ExpandingLabel} from "./common/ExpandingLabel"
import {CheckIf} from "./common/If"
import {UserPublic} from "./pkg/protos/admin/users"
import {useSelector} from "react-redux"
import {sessionUserSelector} from "./store/Store"

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
function UserDetails(props: { details?: UserPublic }) {
    return (
        <List>
            <ListItem>
                <ListItemIcon>
                    <CheckIf cond={props.details?.enabled}/>
                </ListItemIcon>
                <ListItemText primary="Enabled"/>
            </ListItem>
            <ListItem>
                <ListItemIcon>
                    <CheckIf cond={props.details?.neverDelete}/>
                </ListItemIcon>
                <ListItemText primary="Protected"/>
            </ListItem>
            <ListItem/>

            <ListItem>
                <ListItemText primary="Rights:"/>
            </ListItem>
            <ListRight
                cond={props.details?.rights?.canStepTime}
                text="Can Step Time"/>
            <ListRight
                cond={props.details?.rights?.canInjectFaults}
                text="Can Inject Faults"/>
            <ListRight
                cond={props.details?.rights?.canManageAccounts}
                text="Can Manage Accounts"/>
            <ListRight
                cond={props.details?.rights?.canModifyInventory}
                text="Can Modify Inventory"/>
            <ListRight
                cond={props.details?.rights?.canModifyWorkloads}
                text="Can Modify Workloads"/>
            <ListRight
                cond={props.details?.rights?.canPerformRepairs}
                text="Can Perform Repairs"/>
        </List>
    )
}

export function CommandBar(props: {
    onLogout: (name: string) => void
}) {
    const classes = useStyles()

    const sessionUser = useSelector(sessionUserSelector)

    const rights = sessionUser?.details.rights
    const disableStepTime = rights !== undefined ? !rights.canStepTime : true

    const name = String(sessionUser?.name)

    return (
        <div className={classes.root}>
            <AppBar position="static">
                <Toolbar variant="dense">
                    <ExpandingLabel
                        label={name}
                        variant="subtitle2"
                    >
                        <UserDetails details={sessionUser?.details} />
                    </ExpandingLabel>

                    <Tooltip title="log out">
                        <IconButton
                            color="inherit"
                            className={classes.rightIcon}
                            onClick={() => props.onLogout(name)}
                        >
                            <ExitToApp />
                        </IconButton>
                    </Tooltip>

                    <div className={classes.root} />

                    <Stepper disabled={disableStepTime} />
                    <Settings />
                </Toolbar>
            </AppBar>
        </div>
    )
}
