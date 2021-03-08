// This module provides the visual component that lists the set of known users.
// This includes controls to add a new user, and to delete one in the list

import React from "react";
import {Box, IconButton, List, ListItem, ListItemIcon, ListItemText, Paper, Typography} from "@material-ui/core";
import PersonAddIcon from "@material-ui/icons/PersonAdd";
import DeleteIcon from "@material-ui/icons/Delete";
import DeleteOutlineOutlinedIcon from '@material-ui/icons/DeleteOutlineOutlined';
import {makeStyles} from "@material-ui/core/styles";
import {UserList_Entry} from "../../../../pkg/protos/admin/users"

const useStyles = makeStyles(theme => ({
    title: {
        margin: theme.spacing(0, 0, 0),
    },
    filler: {
        flexGrow: 1
    },
    area: (props: {height: number}) => ({
        height: props.height,
        overflow: 'auto'
    })
}));

function DeleteAnnotation(props: {
    protected: boolean,
    onClick: () => void
}) {
    if (props.protected) {
        return <ListItemIcon>
            <DeleteOutlineOutlinedIcon color="disabled" />
        </ListItemIcon>
    }

    return <ListItemIcon
        onClick={props.onClick}>
        <DeleteIcon />
    </ListItemIcon>
}
export function ListUsers(props: {
        height: number,
        users: UserList_Entry[],
        selectedUser: string,
        onSelectUser: (key: string) => void,
        onNewUser: (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => void,
        onDeleteUser: (key: string) => void
    }) {
    const classes = useStyles(props)

    return <Paper className={classes.area}>
        <div style={{width: '100%'}}>
            <Box bgcolor="secondary" display="flex">
                <Box bgcolor="secondary" pt={1} pl={1} className={classes.filler}>
                    <Typography variant="subtitle1" className={classes.title}>
                        Users
                    </Typography>
                </Box>
                <Box bgcolor="secondary" pt={1} pr={1}>
                    <IconButton size="small" onClick={props.onNewUser}>
                        <PersonAddIcon />
                    </IconButton>
                </Box>
            </Box>
        </div>
        <Paper variant="outlined">
            <List dense disablePadding>
                {props.users.map(val =>
                    <ListItem
                        button dense
                        selected={props.selectedUser === val.name}
                        onClick={() => props.onSelectUser(val.name)}>
                        <ListItemText primary={val.name}/>
                        <DeleteAnnotation
                            protected={val.protected}
                            onClick={() => props.onDeleteUser(val.name)}/>
                    </ListItem>
                )}
            </List>
        </Paper>
    </Paper>
}
