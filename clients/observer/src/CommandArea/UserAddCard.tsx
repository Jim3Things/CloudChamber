// This module implements a visual card to add a new user to Cloud Chamber

import React from "react";
import {
    Avatar,
    Button,
    Card,
    CardActions,
    CardContent,
    CardHeader,
    Checkbox,
    FormControl,
    FormControlLabel, FormGroup, TextField
} from "@material-ui/core";
import PersonIcon from "@material-ui/icons/Person";
import {makeStyles} from "@material-ui/core/styles";

import {UserDetails} from "../proxies/UsersProxy";
import {PasswordTextField} from "../common/PasswordTextField";

const useStyles = makeStyles(theme => ({
    area: {
        minHeight: 300,
        maxHeight: 450,
        overflow: 'auto'
    },
    card: {
        margin: theme.spacing(2, 2, 2)
    },
    controls: {
        display: 'flex',
        alignItems: 'right',
        paddingLeft: theme.spacing(1),
        paddingBottom: theme.spacing(1),
    },
}));

export function UserAddCard(props: {
    name: string,
    user: UserDetails,
    onModify: (name: string, user: UserDetails) => void,
    onAdd: () => void,
    onCancel: () => void
}) {
    const classes = useStyles()

    const handleNameChange = (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
        props.onModify(event.target.value, props.user)

    const handlePasswordChange = (event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
        props.onModify(props.name,{ ...props.user, password: event.target.value })

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) =>
        props.onModify(props.name,{ ...props.user, [event.target.name]: event.target.checked })

    return <div className={classes.area}>
        <Card variant="elevation" elevation={3} className={classes.card}>
            <CardHeader
                title="Create New User"
                avatar={
                    <Avatar>
                        <PersonIcon/>
                    </Avatar>
                }
            />

            <CardContent>
                <FormGroup>
                    <TextField
                        autoFocus
                        margin="dense"
                        label="User name"
                        variant="outlined"
                        value={props.name}
                        onChange={handleNameChange}
                        id="name"
                    />

                    <PasswordTextField
                        value={props.user.password}
                        label="Password"
                        onChange={handlePasswordChange}
                    />

                    <FormControl component="fieldset" margin="normal">
                        <FormControlLabel
                            labelPlacement="end"
                            label="Enabled"
                            control={
                                <Checkbox
                                    name="enabled"
                                    checked={props.user.enabled}
                                    onChange={handleChange}
                                />}
                        />
                        <FormControlLabel
                            labelPlacement="end"
                            label="Can Manage Accounts"
                            control={
                                <Checkbox
                                    name="canManageAccounts"
                                    checked={props.user.canManageAccounts}
                                    onChange={handleChange}
                                />}
                        />
                    </FormControl>
                </FormGroup>
            </CardContent>

            <CardActions className={classes.controls}>
                <Button
                    size="small"
                    color="primary"
                    onClick={props.onAdd}>
                    Add
                </Button>
                <Button
                    size="small"
                    color="primary"
                    onClick={props.onCancel}>
                    Cancel
                </Button>
            </CardActions>
        </Card>
    </div>
}