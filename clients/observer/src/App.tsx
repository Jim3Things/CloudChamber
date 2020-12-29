import React, {Component} from 'react';

import './App.css';
import {StepperMode, StepperPolicy, StepperProxy, Timestamp} from "./proxies/StepperProxy";
import {CommandTab} from "./CommandBar";
import {UsersProxy} from "./proxies/UsersProxy";
import {InventoryProxy} from "./proxies/InventoryProxy";
import {getErrorDetails, Session, SessionUser} from "./proxies/Session";
import {RenderIf} from "./common/If";
import {MainPage} from "./MainPage/Main";
import {Login} from "./MainPage/Login";
import {PingProxy} from "./proxies/PingProxy";
import {LogEntries, LogEntry, LogProxy} from "./proxies/LogProxy";
import {Organizer} from "./Log/Organizer";
import {SettingsState} from "./Settings";

interface Props {

}

interface State {
    StepperPolicy: StepperPolicy,
    stepperProxy: StepperProxy,
    usersProxy: UsersProxy,
    inventoryProxy: InventoryProxy,
    pingProxy: PingProxy,
    logProxy: LogProxy,
    session: Session,
    organizer: Organizer,
    entries: LogEntry[],
    cur: Timestamp,
    tab: CommandTab
    activeSession: boolean
    sessionUser: SessionUser
    logonUser: string
    logonPassword: string
    logonError: string
    settings: SettingsState
}

// Format and display the logon dialog box if we do not have an active

export class App extends Component<Props & any, State> {

    // Initiate a login to a session
    onLogon = () => {
        this.state.session.logon(this.state.logonUser, this.state.logonPassword)
            .then(value => {
                // It worked, so record the session state, and start the
                // background calls to get the next tick
                this.setState(
                    {
                        sessionUser: value,
                        activeSession: true,
                        logonPassword: "",
                        logonError: ""
                    })
                this.state.stepperProxy.getStatus()
                this.state.pingProxy.start()
                this.state.logProxy.start()
            })
            .catch(msg => getErrorDetails(msg, details =>
                this.setState({
                    activeSession: false,
                    logonError: details,
                    logonPassword: ""
                }))
            )
    }

    stepperPolicyEvent = (policy: StepperPolicy) => {
        this.setState({StepperPolicy: policy});
        this.state.stepperProxy.changePolicy(policy);
    }

    settingsChangeEvent = (settings: SettingsState) => {
        this.setState({settings: settings} )
    }

    onTimeEvent = (cur: Timestamp) => {
        this.setState({ cur: cur });
    }

    onNewLogEvent = (toHold: number, events: LogEntries) => {
        const newEntries = [...this.state.entries, ...events.entries]
        const start = Math.max(newEntries.length - toHold, 0)
        const slice = newEntries.slice(start)
        const organizer = new Organizer(slice, this.state.organizer)

        this.setState({
            entries: slice,
            organizer: organizer
        })
    }

    onExpansionHandler = (id: string) : void => {
        const org = this.state.organizer
        org.flip(id)
        this.setState({organizer: org})
    }

    // Initiate a logout from the active session
    onLogoutEvent = () => {
        this.state.session.logout(this.state.sessionUser.name)
            .then(() => {
                // We're logged out.  Set the state and cancel the
                // background calls for the next tick
                this.setState({activeSession: false})
                this.state.stepperProxy.cancelUpdates()
                this.state.pingProxy.cancel()
                this.state.logProxy.cancelUpdates()
            })
    }

    render() {
        return <div className="App">
            <link rel="stylesheet"
                  href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap"/>
            <RenderIf cond={this.state.activeSession}>
                <MainPage
                    tab={this.state.tab}
                    activeSession={this.state.activeSession}
                    sessionUser={this.state.sessionUser.name}
                    settings={this.state.settings}
                    onCommandSelect={(tab: CommandTab) => this.setState({tab: tab})}
                    onPolicyEvent={this.stepperPolicyEvent}
                    onSettingsChange={this.settingsChangeEvent}
                    onLogout={this.onLogoutEvent}
                    usersProxy={this.state.usersProxy}
                    proxy={this.state.inventoryProxy}
                    cur={this.state.cur}
                    organizer={this.state.organizer}
                    onTrackChange={this.onExpansionHandler}
                />
            </RenderIf>

            <RenderIf cond={!this.state.activeSession}>
                <Login
                    onClose={this.onLogon}
                    userName={this.state.logonUser}
                    onUserNameChange={(event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                        this.setState({logonUser: event.target.value})}
                    password={this.state.logonPassword}
                    onPasswordChange={(event: React.ChangeEvent<HTMLTextAreaElement | HTMLInputElement>) =>
                        this.setState({logonPassword: event.target.value})}
                    logonError={this.state.logonError}
                />
            </RenderIf>

        </div>;
    }

    state: State = {
        StepperPolicy: StepperPolicy.Pause,
        stepperProxy: new StepperProxy(this.onTimeEvent),
        usersProxy: new UsersProxy(),
        inventoryProxy: new InventoryProxy(),
        pingProxy: new PingProxy(),
        logProxy: new LogProxy(this.onNewLogEvent),
        session: new Session(),
        activeSession: false,
        sessionUser: {
            name: "",
            enabled: false,
            accountManager: false,
            neverDelete: false
        },
        logonUser: "",
        logonPassword: "",
        logonError: "",
        cur: {
            mode: StepperMode.Paused,
            now: 0,
            rate: 0
        },
        entries:[],
        organizer: new Organizer([]),
        tab: CommandTab.Admin,
        settings: {
            logSettings: {
                showDebug: true,
                showInfra: true
            }
        }
    };
}

export default App;