import React, {useState} from "react";
import {
    AppBar,
    IconButton,
    List,
    ListItem,
    ListItemIcon,
    ListItemText,
    Popover,
    Toolbar,
    Tooltip,
    Typography
} from "@material-ui/core";
import {makeStyles, Theme} from "@material-ui/core/styles";

import {Stepper} from "./Stepper";
import {SetStepperPolicy} from "./proxies/StepperProxy";
import {CheckBox, CheckBoxOutlineBlank, ExitToApp} from "@material-ui/icons";
import {Settings, SettingsState} from "./Settings";
import {SessionUser} from "./proxies/Session";
import {MoreOrLess} from "./common/If";

const useStyles = makeStyles((theme: Theme) => ({
    root: {
        flexGrow: 1
    },
    paper: {
        padding: theme.spacing(1),
    },
    rightIcon: {
        MarginLeft: theme.spacing(1)
    },
    nested: {
        paddingLeft: theme.spacing(4),
        pt: 0,
        pb: 0
    }
}));

function CheckIf(props : {cond: boolean}) {
    if (props.cond) {
        return <CheckBox />
    } else {
        return <CheckBoxOutlineBlank />
    }
}

function ListRight(props: {
    cond: boolean | undefined,
    text: string
}) {
    const classes = useStyles()

    const test = Boolean(props.cond)

    return <ListItem dense className={classes.nested}>
        <ListItemIcon>
            <CheckIf cond={test} />
        </ListItemIcon>
        <ListItemText primary={props.text} />
    </ListItem>

}

function UserDetails(props: {user: SessionUser}) {
    const classes = useStyles();

    const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null);

    const onClick = (e: React.MouseEvent<HTMLButtonElement | undefined>) => {
        if (open) {
            setAnchorEl(null)
        } else {
            setAnchorEl(e.currentTarget)
        }
    }

    const onClose = () => setAnchorEl(null)

    const open = Boolean(anchorEl)

    return <React.Fragment>
        <IconButton
            color="inherit"
            onClick={onClick}
            aria-owns={open ? 'detail-popover' : undefined}
            aria-haspopup={true}
            >
            <MoreOrLess cond={open} />
        </IconButton>

        <Popover
            id="detail-popover"
            classes={{paper:classes.paper }}
            anchorEl={anchorEl}
            anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'left'
            }}
            transformOrigin={{
                vertical: 'top',
                horizontal: 'left'
            }}
            onClose={onClose}
            open={open}
            disableRestoreFocus
            >
            <List>
                <ListItem>
                    <ListItemIcon>
                        <CheckIf cond={props.user.details.enabled} />
                    </ListItemIcon>
                    <ListItemText primary="Enabled" />
                </ListItem>
                <ListItem>
                    <ListItemIcon>
                        <CheckIf cond={props.user.details.neverDelete} />
                    </ListItemIcon>
                    <ListItemText primary="Protected" />
                </ListItem>
                <ListItem />

                <ListItem>
                    <ListItemText primary="Rights:" />
                </ListItem>
                <ListRight
                    cond={props.user.details.rights?.canStepTime}
                    text="Can Step Time" />
                <ListRight
                    cond={props.user.details.rights?.canInjectFaults}
                    text="Can Inject Faults" />
                <ListRight
                    cond={props.user.details.rights?.canManageAccounts}
                    text="Can Manage Accounts" />
                <ListRight
                    cond={props.user.details.rights?.canModifyInventory}
                    text="Can Modify Inventory" />
                <ListRight
                    cond={props.user.details.rights?.canModifyWorkloads}
                    text="Can Modify Workloads" />
                <ListRight
                    cond={props.user.details.rights?.canPerformRepairs}
                    text="Can Perform Repairs" />
            </List>
        </Popover>
    </React.Fragment>
}
export function CommandBar(props: {
            sessionUser: SessionUser,
            settings: SettingsState,
            onPolicyEvent: (policy: SetStepperPolicy) => void,
            onSettingsChange: (settings: SettingsState) => void,
            onLogout: () => void
}) {
    const classes = useStyles();

    const rights = props.sessionUser.details.rights
    const disableStepTime = rights !== undefined ? !rights.canStepTime : true

    return (
        <div className={classes.root}>
            <AppBar position="static">
                <Toolbar variant="dense">
                    <Typography variant="subtitle2">
                        {props.sessionUser.name}
                    </Typography>
                    <UserDetails user={props.sessionUser} />
                    <Tooltip title="log out">
                        <IconButton
                            color="inherit"
                            className={classes.rightIcon}
                            onClick={() => props.onLogout()}>
                                <ExitToApp/>
                        </IconButton>
                    </Tooltip>
                    <div className={classes.root}/>
                        <Stepper disabled={disableStepTime} onPolicyEvent={props.onPolicyEvent}/>
                    <Settings
                        settings={props.settings}
                        onChange={props.onSettingsChange}
                    />
                </Toolbar>
            </AppBar>
        </div>
    );
}
