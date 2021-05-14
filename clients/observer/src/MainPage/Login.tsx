import React, {useState} from 'react'
import {Button, Dialog, DialogActions, DialogContent, DialogTitle, FormGroup, TextField} from "@material-ui/core"

import {PasswordTextField} from "../common/PasswordTextField"
import {AlertIf} from "../common/AlertIf"
import {useSelector} from "react-redux"
import {logonErrorSelector} from "../store/Store"

export function Login(props: {
    onClose: (name: string, password: string) => void
}) {
    const [userName, setUserName] = useState<string>("")
    const [password, setPassword] = useState<string>("")

    const logonError = useSelector(logonErrorSelector)

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
                    value={userName}
                    onChange={(e) => setUserName(e.target.value)}
                    id="name"
                />

                <PasswordTextField
                    value={password}
                    label="Password"
                    onChange={(e) => setPassword(e.target.value)}/>
            </FormGroup>

            <AlertIf title="Login Failed" text={logonError}/>

        </DialogContent>
        <DialogActions>
            <Button onClick={() => props.onClose(userName, password)}>
                Submit
            </Button>
        </DialogActions>
    </Dialog>
}
