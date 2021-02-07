// This module implements a visual card containing the public detail data
// about as user.  The non-immutable values about a user can be changed via
// this card.

import React, {useState} from "react";
import {
    Avatar,
    Button,
    Card,
    CardActions,
    CardContent,
    CardHeader,
    Checkbox, Collapse,
    FormControl,
    FormControlLabel, FormGroup
} from "@material-ui/core";
import PersonIcon from "@material-ui/icons/Person";
import ExpandLessIcon from "@material-ui/icons/ExpandLess"
import ExpandMoreIcon from "@material-ui/icons/ExpandMore"
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
        paddingLeft: theme.spacing(1),
        paddingBottom: theme.spacing(1),
    },
    expand: {
        marginLeft: 'auto'
    }
}));

const ExpandIcon = (props: {expanded: boolean}) =>
    props.expanded ? <ExpandLessIcon /> : <ExpandMoreIcon />

export function UserDetailsCard(props: {
    name: string,
    user: UserDetails,
    onModify: (user: UserDetails) => void,
    onSave: () => void,
    onReset: () => void,
    onSetPassword: (oldPass: string, newPass: string) => void
}) {
    const classes = useStyles()

    const [expanded, setExpanded] = useState(false)
    const [oldPass, setOldPass] = useState("")
    const [newPass, setNewPass] = useState("")

    const toggleExpanded = () => setExpanded(!expanded)

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        props.onModify({ ...props.user, [event.target.name]: event.target.checked });
    }

    return <div className={classes.area}>
        <Card variant="elevation" elevation={3} className={classes.card}>
            <CardHeader
                title={props.name}
                subheader={props.user.neverDelete ? "Protected" : undefined}
                avatar={
                    <Avatar>
                        <PersonIcon/>
                    </Avatar>
                }
            />

            <CardContent>
                <FormControl component="fieldset" margin="dense">
                    <FormControlLabel
                        labelPlacement="end"
                        label="Enabled"
                        control={
                            <Checkbox
                                disabled={props.user.neverDelete}
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
                                disabled={props.user.neverDelete}
                                name="canManageAccounts"
                                checked={props.user.canManageAccounts}
                                onChange={handleChange}
                            />}
                    />
                </FormControl>
            </CardContent>

            <CardActions
                className={classes.controls} disableSpacing={true}>
                <Button
                    disabled={props.user.neverDelete}
                    size="small"
                    color="primary"
                    onClick={props.onSave}>
                    Save
                </Button>
                <Button
                    disabled={props.user.neverDelete}
                    size="small"
                    color="primary"
                    onClick={props.onReset}>
                    Reset
                </Button>
                <Button
                    className={classes.expand}
                    size="small"
                    color="primary"
                    onClick={toggleExpanded}>
                    Change Password
                    <ExpandIcon expanded={expanded} />
                </Button>
            </CardActions>
            <Collapse in={expanded}>
                <FormGroup>
                    <PasswordTextField
                        value={oldPass}
                        label="Current password"
                        onChange={(event) => setOldPass(event.target.value)}
                    />
                    <PasswordTextField
                        value={newPass}
                        label="New password"
                        onChange={(event) => setNewPass(event.target.value)}
                    />
                </FormGroup>
                <Button
                    className={classes.expand}
                    size="small"
                    color="primary"
                    onClick={() => props.onSetPassword(oldPass, newPass)}>
                    Set Password
                </Button>
            </Collapse>
        </Card>
    </div>
}