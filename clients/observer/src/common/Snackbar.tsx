// This module provides a snackbar formatted to indicate success

import {IconButton, Snackbar} from "@material-ui/core";
import CloseIcon from "@material-ui/icons/Close";
import React from "react";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles(theme => ({
        success: {
            backgroundColor: theme.palette.success.main,
            color: theme.palette.success.contrastText
        },
    error: {
        backgroundColor: theme.palette.error.main,
        color: theme.palette.error.contrastText
    }

}))

export enum MessageMode {
    None = 0,                   // Show no snackbar
    Success = 1,                // Show the success snackbar
    Error = 2                   // Show the error snackbar
}

export interface SnackData {
    message: string
    mode: MessageMode
}

export function ErrorSnackbar(props: {
    open: boolean,
    autoHideDuration: number,
    message: string
    onClose: () => void
}) {
    const classes = useStyles()

    return <Snackbar
        open={props.open}
        onClose={props.onClose}
        autoHideDuration={props.autoHideDuration}
        action={[
            <IconButton
                color="inherit"
                onClick={props.onClose} >
                <CloseIcon />
            </IconButton>
        ]}
        anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center'
        }}
        message={props.message}
        ContentProps={{
            classes: { root: classes.error }
        }}
    />
}

export function SuccessSnackbar(props: {
    open: boolean,
    autoHideDuration: number,
    message: string
    onClose: () => void
}) {
    const classes = useStyles()

    return <Snackbar
        open={props.open}
        onClose={props.onClose}
        autoHideDuration={props.autoHideDuration}
        action={[
            <IconButton
                color="inherit"
                onClick={props.onClose} >
                <CloseIcon />
            </IconButton>
        ]}
        anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'center'
        }}
        message={props.message}
        ContentProps={{
            classes: { root: classes.success }
        }}
    />
}
