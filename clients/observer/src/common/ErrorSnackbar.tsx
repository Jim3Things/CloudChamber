// This module provides a snackbar formatted for an error status

import {IconButton, Snackbar} from "@material-ui/core";
import CloseIcon from "@material-ui/icons/Close";
import React from "react";
import {makeStyles} from "@material-ui/core/styles";

const useStyles = makeStyles(theme => ({
    success: {
        backgroundColor: theme.palette.error.main,
        color: theme.palette.error.contrastText
    }
}))

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
            classes: { root: classes.success }
        }}
    />
}