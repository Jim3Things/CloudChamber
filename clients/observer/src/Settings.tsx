// Settings manages the transient view settings for the UI via a defined state
// structure held by the app, and a dialog for manipulating it.

import React from 'react'
import {createStyles, makeStyles, Theme} from '@material-ui/core/styles'
import SettingsIcon from '@material-ui/icons/Settings'
import {
    Button,
    Checkbox,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    FormControlLabel,
    IconButton
} from "@material-ui/core"
import {settingsSelector, settingsSlice, useAppDispatch, useAppSelector} from "./store/Store"

// LogSettings are the filtering options that determine which log entries to
// show in the display.
export interface LogSettings {
    showDebug: boolean          // True to show debug traces
    showInfra: boolean          // True to show simulation infrastructure entries
}

// SettingsState holds the collected set of display settings.
export interface SettingsState {
    logSettings: LogSettings
}

const useStyles = makeStyles((theme: Theme) =>
    createStyles({
        dialog: {
            padding: theme.spacing(1),
            backgroundColor: theme.palette.background.paper,
            minHeight: 300,
            maxHeight: 450,
            overflow: 'auto',
            margin: theme.spacing(2, 2, 2)
        },
        controls: {
            display: 'flex',
            paddingLeft: theme.spacing(1),
            paddingBottom: theme.spacing(1),
            backgroundColor: theme.palette.background.paper,
        }
    })
)

// Settings manages a settings icon button that opens a dialog when clicked.
// It also provides a pass-through for the settings state update handlers.
export function Settings() {
    const dispatch = useAppDispatch()
    const settings = useAppSelector(settingsSelector)

    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null)
    const [working, setWorking] = React.useState<SettingsState>(settings)

    const open = Boolean(anchorEl)

    const handleOpenSettings = (event: React.MouseEvent<HTMLButtonElement>) => {
        setWorking({...settings})
        setAnchorEl(event.currentTarget)
    }

    const handleClose = () => {
        setAnchorEl(null)
        dispatch(settingsSlice.actions.update(working))
    }

    const handleCancel = () => {
        setAnchorEl(null)
    }

    return (
        <div>
            <IconButton
                color="inherit"
                aria-owns={open ? 'settings-dialog' : undefined}
                aria-haspopup="true"
                onClick={handleOpenSettings}
            >
                <SettingsIcon/>
            </IconButton>
            <SettingsDialog
                open={open}
                settings={working}
                onChange={setWorking}
                onCancel={handleCancel}
                onSave={handleClose}
            />
        </div>
    )
}

// SettingsDialog displays and manages the dialog that supports editing of the
// current display settings.
export function SettingsDialog(props: {
    open: boolean,
    settings: SettingsState,
    onChange: (settings: SettingsState) => void,
    onCancel: () => void,
    onSave: () => void
}) {
    const classes = useStyles()

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const logSettings = {...props.settings.logSettings, [event.target.name]: event.target.checked}
        const newSettings = {...props.settings, logSettings: logSettings}
        props.onChange(newSettings)
    }

    return <Dialog
        id="settings-dialog"
        open={props.open}
        className={classes.dialog}
    >

        <DialogTitle title="Cloud Chamber Settings"/>
        <DialogContent>
            <FormControl component="fieldset" margin="dense">
                <FormControlLabel
                    labelPlacement="end"
                    label="Display Debug Traces"
                    control={
                        <Checkbox
                            disabled={false}
                            name="showDebug"
                            checked={props.settings.logSettings.showDebug}
                            onChange={handleChange}
                        />}
                />
                <FormControlLabel
                    labelPlacement="end"
                    label="Display Internal Simulation Detail Traces"
                    control={
                        <Checkbox
                            disabled={false}
                            name="showInfra"
                            checked={props.settings.logSettings.showInfra}
                            onChange={handleChange}
                        />}
                />
            </FormControl>
        </DialogContent>

        <DialogActions className={classes.controls} disableSpacing={true}>
            <Button
                disabled={false}
                size="small"
                color="primary"
                onClick={props.onSave}>
                Save
            </Button>
            <Button
                disabled={false}
                size="small"
                color="primary"
                onClick={props.onCancel}>
                Cancel
            </Button>
        </DialogActions>
    </Dialog>
}
