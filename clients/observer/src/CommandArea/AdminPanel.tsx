// This module contains the Administration sub-panel and associated dialogs

import React, {useEffect, useState} from "react";

import {UserDetails, UsersProxy} from "../proxies/UsersProxy";
import {getErrorDetails} from "../proxies/Session";
import {Container, Item} from "../common/Cells";
import {ListUsers} from "./UsersList";
import {UserDetailsCard} from "./UserDetailsCard";
import {RenderIf} from "../common/If";
import {UserAddCard} from "./UserAddCard";
import {ErrorSnackbar, MessageMode, SnackData, SuccessSnackbar} from "../common/Snackbar";

import {UserList, UserList_Entry} from "../../../../pkg/protos/admin/users"

const initialSnackData: SnackData = {
    message: "",
    mode: MessageMode.None
}

const initialUsers: UserList_Entry[] = []

export function AdminPanel(props: {
    height: number,
    usersProxy: UsersProxy,
    sessionUser: string
}) {
    const [snackData, setSnackData] = useState(initialSnackData)

    const [users, setUsers] = useState(initialUsers)
    const [addInProgress, setAddInProgress] = useState(false)
    const [newUserName, setNewUserName] = useState("")
    const [selectedUser, setSelectedUser] = useState("")
    const [cleanUser, setCleanUser] = useState(new UserDetails())
    const [editUser, setEditUser] = useState(new UserDetails())
    const [detailsLoaded, setDetailsLoaded] = useState(false)

    const UpdateErrorState = (msg: any) => {
        getErrorDetails(msg, (details) => setSnackData({
            mode: MessageMode.Error,
            message: details
        }))
    }

    const findAfter = (name: string, list: UserList_Entry[]) : string | undefined => {
        const result = list.find(item => item.name > name)
        return result === undefined ? undefined : result.name
    }

    const findNeighbor = (name: string) : string => {
        const after = findAfter(name, users)
        if (after === undefined) {
            let reversed = [...users]
            reversed.reverse()

            const before = findAfter(name, reversed)
            return before === undefined ? users[0].name : before
        }

        return after
    }

    // +++ Functions to access the users store

    // These all follow a pattern of invoking the proxy and then modifying
    // the state when the asynchronous event completes.  That state change
    // then triggers a re-execution of the render() method.

    // Get the list of users
    const refreshUserList = (name: string | undefined) => {
        props.usersProxy.list()
            .then((list: UserList) =>
            {
                // Order the list by name
                const newList = list.users.sort((n1, n2) => {
                    const left = n1.name.toLowerCase()
                    const right = n2.name.toLowerCase()

                    if (right < left) {
                        return 1
                    }

                    if (right > left) {
                        return -1
                    }

                    return 0
                })

                let target = name

                if (name !== undefined) {
                    const result = newList.find(item => item.name >= name)
                    target = result === undefined ? undefined : result.name
                }

                if (target === undefined) {
                    target = newList[0].name
                }

                setUsers(newList)
                setSelectedUser(target)
                setDetailsLoaded(false)

                onFetchUser(target)
            })
            .catch(() => {
                setUsers([{ name : props.sessionUser, uri: "", protected: false}])
                setSelectedUser(props.sessionUser)
                setDetailsLoaded(false)

                onFetchUser(props.sessionUser)
            })
    }

    // Get the details for a specific user
    const onFetchUser = (name: string) => {
        props.usersProxy.get(name)
            .then((item: UserDetails) => {
                setCleanUser(item)
                setEditUser(item)
                setDetailsLoaded(true)
            })
            .catch((msg: any) => {
                if (users.find(item => item.name === name) !== undefined) {
                    UpdateErrorState(msg)
                }
            })
    }

    const onAdd = () => {
        props.usersProxy.add(newUserName, editUser)
            .then(() => {
                setAddInProgress(false)
                setSelectedUser(newUserName)
                setDetailsLoaded(false)
                setSnackData({
                    mode: MessageMode.Success,
                    message: "User " + newUserName + " was successfully created"
                })

                refreshUserList(newUserName)
            })
            .catch((msg: any) => UpdateErrorState(msg))
    }

    const onSaveEdit = () => {
        const name = selectedUser
        props.usersProxy.set(name, editUser)
            .then((item: UserDetails) => {
                if (name === selectedUser) {
                    // We're still focused on the updated user, so show that.
                    setCleanUser(item)
                    setEditUser(item)
                    setDetailsLoaded(true)
                }

                setSnackData({
                    mode: MessageMode.Success,
                    message: "User " + name + " was successfully updated"
                })
            })
            .catch((msg: any) => {
                if (users.find(item => item.name === name) !== undefined) {
                    UpdateErrorState(msg)
                }
            })
    }

    const onDelete = (name: string) => {
        const pick = (name === selectedUser) ? findNeighbor(name) : name
        const newList = users.filter(item => item.name !== name)

        setUsers(newList)
        setSelectedUser(pick)
        setDetailsLoaded(pick !== name)

        props.usersProxy.remove(name)
            .then(() => {
                setSnackData({
                    mode: MessageMode.Success,
                    message: "User " + name + " was successfully deleted"
                })

                refreshUserList(pick);
            })
            .catch((msg: any) => {
                UpdateErrorState(msg)
                onFetchUser(name)
            })
    }

    const onSetPassword = (old: string, newPass: string) => {
        const name = selectedUser

        props.usersProxy.setPassword(name, editUser, old, newPass)
            .then((tag: number) => {
                if (name === selectedUser) {
                    const details = { ...editUser, eTag: tag }
                    setCleanUser(details)
                    setEditUser(details)
                    setDetailsLoaded(true)
                }

                setSnackData({
                    mode: MessageMode.Success,
                    message: "Password for user " + name + " has been successfully changed"})
            })
            .catch((msg: any) => UpdateErrorState(msg))
    }

    // --- Functions to access the users store

    // Prepare for a new user operation.  This is a little different from the
    // edit case, inasmuch as there is no data to fetch.  This just resets the
    // user details and new user name to their default values and opens the
    // add dialog.
    const onPrepAdd = () => {
        const user = new UserDetails()

        setEditUser(user)
        setCleanUser(user)
        setNewUserName("")
        setAddInProgress(true)
    }

    // On component load, capture the initial list of users
    useEffect(() => {
        refreshUserList(undefined)
    }, [props.usersProxy])

    const selectUser = (key: string) => {
        if (selectedUser !== key && (users.find(item => item.name === key) !== undefined)) {
            setSelectedUser(key)
            setDetailsLoaded(false)

            onFetchUser(key)
        }
    }

        return <Container >
            <Item xs={6}>
                <ListUsers
                    height={props.height - 5}
                    users={users}
                    selectedUser={selectedUser}
                    onSelectUser={selectUser}
                    onNewUser={onPrepAdd}
                    onDeleteUser={onDelete}
                />
            </Item>

            <Item xs={6}>
                <RenderIf cond={detailsLoaded && !addInProgress}>
                    <UserDetailsCard
                        name={selectedUser}
                        user={editUser}
                        onModify={(newVal) => setEditUser(newVal) }
                        onSave={onSaveEdit}
                        onReset={() => setEditUser(cleanUser) }
                        onSetPassword={onSetPassword}
                    />
                </RenderIf>

                <RenderIf cond={addInProgress}>
                    <UserAddCard
                        name={newUserName}
                        user={editUser}
                        onModify={(name, newVal) => {
                            setNewUserName(name)
                            setEditUser(newVal)
                        }}
                        onAdd={onAdd}
                        onCancel={() => setAddInProgress(false)} />
                </RenderIf>

                <SuccessSnackbar
                    open={snackData.mode === MessageMode.Success}
                    onClose={() => setSnackData({mode: MessageMode.None, message: ""})}
                    autoHideDuration={3000}
                    message={snackData.message} />

                <ErrorSnackbar
                    open={snackData.mode === MessageMode.Error}
                    onClose={() => setSnackData({mode: MessageMode.None, message: ""})}
                    autoHideDuration={4000}
                    message={snackData.message} />
            </Item>
        </Container>;
}
