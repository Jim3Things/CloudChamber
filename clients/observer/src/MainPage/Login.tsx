import React from 'react'
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, FormGroup, TextField} from "@material-ui/core"

import {PasswordTextField} from "../common/PasswordTextField"
import {AlertIf} from "../common/AlertIf"

export function Login(props: {
    onClose: () => void,
    userName: string,
    onUserNameChange: (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => void,
    password: string,
    onPasswordChange: (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) => void,
    logonError: string
}) {
    return <Dialog
        aria-labelledby="login-title"
        open={true}>

        <DialogTitle id="login-title">Log in to Cloud Chamber</DialogTitle>
        <DialogContent>
            <FormGroup>
                <TextField
                    autoFocus
                    margin="dense"
                    label="User name"
                    variant="outlined"
                    value={props.userName}
                    onChange={props.onUserNameChange}
                    id="name"
                />

                <PasswordTextField
                    value={props.password}
                    label="Password"
                    onChange={props.onPasswordChange}/>
            </FormGroup>

            <AlertIf title="Login Failed" text={props.logonError}/>

        </DialogContent>
        <DialogActions>
            <Button onClick={props.onClose}>
                Submit
            </Button>
        </DialogActions>
    </Dialog>
}
