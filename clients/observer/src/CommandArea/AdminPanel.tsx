// This module contains the Administration sub-panel and associated dialogs

import React, {Component} from "react";

import {JsonUserList, JsonUserListEntry, UserDetails, UsersProxy} from "../proxies/UsersProxy";
import {getErrorDetails} from "../proxies/Session";
import {Container, Item} from "../common/Cells";
import {ListUsers} from "./UsersList";
import {UserDetailsCard} from "./UserDetailsCard";
import {RenderIf} from "../common/If";
import {UserAddCard} from "./UserAddCard";
import {SuccessSnackbar} from "../common/SuccessSnackbar";
import {ErrorSnackbar} from "../common/ErrorSnackbar";

interface Props {
    height: number
    usersProxy: UsersProxy
    sessionUser: string
}

enum MessageMode {
    None = 0,                   // Show no snackbar
    Success,                // Show the success snackbar
    Error                   // Show the error snackbar
}

interface State {
    users: JsonUserListEntry[]; // Set of known user names

    addInProgress: boolean      // True, if we're in the middle of an add user
    newUserName: string;        // What is the new user name (add dialog)

    selectedUser: string;       // Which username is currently selected
    cleanUser: UserDetails      // The user details, unchanged, for the selected user
    editUser: UserDetails;      // User details, either update or add ops
    detailsLoaded: boolean

    snackMode: MessageMode      // Which snackbar to display, if any
    snackText: string           // ... and the text to supply
}

export class AdminPanel extends Component<Props & any, State> {
    state: State = {
        users: [],

        addInProgress: false,
        newUserName: "",

        selectedUser: "",
        cleanUser: new UserDetails(),
        editUser: new UserDetails(),
        detailsLoaded: false,

        snackMode: MessageMode.None,
        snackText: ""
    }

    private UpdateErrorState(msg: any) {
        getErrorDetails(msg, (details) => this.setState({
            snackMode: MessageMode.Error,
            snackText: details
        }));
    }

    private findAfter(name: string, list: JsonUserListEntry[]) : string | undefined {
        const result = list.find(item => item.name > name)
        return result === undefined ? undefined : result.name
    }

    private findNeighbor(name: string) : string {
        const after = this.findAfter(name, this.state.users)
        if (after === undefined) {
            let reversed = [...this.state.users]
            reversed.reverse()

            const before = this.findAfter(name, reversed)
            return before === undefined ? this.state.users[0].name : before
        }

        return after
    }

    // +++ Functions to access the users store

    // These all follow a pattern of invoking the proxy and then modifying
    // the state when the asynchronous event completes.  That state change
    // then triggers a re-execution of the render() method.

    // Get the list of users
    refreshUserList = (name: string | undefined) => {
        this.props.usersProxy.list()
            .then((list: JsonUserList) =>
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

                this.setState({
                    users: newList,
                    selectedUser: target,
                    detailsLoaded: false
                })

                this.onFetchUser(target)
            })
            .catch(() => {
                this.setState({
                    users: [this.props.sessionUser],
                    selectedUser: this.props.sessionUser,
                    detailsLoaded: false
                })

                this.onFetchUser(this.props.sessionUser)
            })
    }

    // Get the details for a specific user
    onFetchUser = (name: string) => {
        this.props.usersProxy.get(name)
            .then((item: UserDetails) => {
                this.setState({
                    cleanUser: item,
                    editUser: item,
                    detailsLoaded: true
                });
            })
            .catch((msg: any) => {
                if (this.state.users.find(item => item.name === name) !== undefined) {
                    this.UpdateErrorState(msg)
                }
            });
    }

    onAdd = () => {
        this.props.usersProxy.add(this.state.newUserName, this.state.editUser)
            .then(() => {
                this.setState({
                    addInProgress: false,
                    selectedUser: this.state.newUserName,
                    detailsLoaded: false,

                    snackMode: MessageMode.Success,
                    snackText: "User " + this.state.newUserName + " was successfully created"
                })

                this.refreshUserList(this.state.newUserName)
            })
            .catch((msg: any) => this.UpdateErrorState(msg))
    }

    onSaveEdit = () => {
        const name = this.state.selectedUser
        this.props.usersProxy.set(name, this.state.editUser)
            .then((item: UserDetails) => {
                if (name === this.state.selectedUser) {
                    this.setState({
                        cleanUser: item,
                        editUser: item,
                        detailsLoaded: true,

                        snackMode: MessageMode.Success,
                        snackText: "User " + name + " was successfully updated"
                    })
                } else {
                    // Selection moved on while the save was processing
                    this.setState({
                        snackMode: MessageMode.Success,
                        snackText: "User " + name + " was successfully updated"
                    })
                }
            })
            .catch((msg: any) => {
                if (this.state.users.find(item => item.name === name) !== undefined) {
                    this.UpdateErrorState(msg)
                }
            });
    }

    onDelete = (name: string) => {
        const pick = (name === this.state.selectedUser) ? this.findNeighbor(name) : name

        const newList = this.state.users.filter(item => item.name !== name)

        this.setState({
            users: newList,
            selectedUser: pick,
            detailsLoaded: pick !== name
        })

        this.props.usersProxy.remove(name)
            .then(() => {
                this.setState({
                    snackMode: MessageMode.Success,
                    snackText: "User " + name + " was successfully deleted"
                })

                this.refreshUserList(pick);
            })
            .catch((msg: any) => {
                this.UpdateErrorState(msg)
                this.onFetchUser(name)
            });
    }

    onSetPassword = (old: string, newPass: string) => {
        const name = this.state.selectedUser

        this.props.usersProxy.setPassword(name, this.state.editUser, old, newPass)
            .then((tag: number) => {
                if (name === this.state.selectedUser) {
                    const details = { ...this.state.editUser, eTag: tag }
                    this.setState({
                        cleanUser: details,
                        editUser: details,
                        detailsLoaded: true,
                    })
                }

                this.setState({
                    snackMode: MessageMode.Success,
                    snackText: "Password for user " + name + " has been successfully changed"})
            })
            .catch((msg: any) => this.UpdateErrorState(msg))
    }

    // --- Functions to access the users store

    // Prepare for a new user operation.  This is a little different from the
    // edit case, inasmuch as there is no data to fetch.  This just resets the
    // user details and new user name to their default values and opens the
    // add dialog.
    onPrepAdd = () => {
        const user = new UserDetails()

        this.setState({
            editUser: user,
            cleanUser: user,
            newUserName: "",
            addInProgress: true
        });
    }

    // On component load, capture the initial list of users
    componentWillMount() {
        this.refreshUserList(undefined);
    }

    selectUser = (key: string) => {
        if (this.state.selectedUser !== key &&
            (this.state.users.find(item => item.name === key) !== undefined)) {
            this.setState({
                selectedUser: key,
                detailsLoaded: false
            })

            this.onFetchUser(key)
        }
    }

    render() {
        return <Container >
            <Item xs={6}>
                <ListUsers
                    height={this.props.height - 5}
                    users={this.state.users}
                    selectedUser={this.state.selectedUser}
                    onSelectUser={this.selectUser}
                    onNewUser={this.onPrepAdd}
                    onDeleteUser={this.onDelete}
                />
            </Item>

            <Item xs={6}>
                <RenderIf cond={this.state.detailsLoaded && !this.state.addInProgress}>
                    <UserDetailsCard
                        name={this.state.selectedUser}
                        user={this.state.editUser}
                        onModify={(newVal) => this.setState({editUser: newVal}) }
                        onSave={this.onSaveEdit}
                        onReset={() => this.setState({editUser: this.state.cleanUser}) }
                        onSetPassword={this.onSetPassword}
                    />
                </RenderIf>

                <RenderIf cond={this.state.addInProgress}>
                    <UserAddCard
                        name={this.state.newUserName}
                        user={this.state.editUser}
                        onModify={(name, newVal) => this.setState({
                            newUserName: name,
                            editUser: newVal
                        }) }
                        onAdd={this.onAdd}
                        onCancel={() => this.setState({ addInProgress: false } )} />
                </RenderIf>

                <SuccessSnackbar
                    open={this.state.snackMode === MessageMode.Success}
                    onClose={() => this.setState({snackMode: MessageMode.None})}
                    autoHideDuration={3000}
                    message={this.state.snackText} />

                <ErrorSnackbar
                    open={this.state.snackMode === MessageMode.Error}
                    onClose={() => this.setState({snackMode: MessageMode.None})}
                    autoHideDuration={4000}
                    message={this.state.snackText} />
            </Item>
        </Container>;
    }
}
