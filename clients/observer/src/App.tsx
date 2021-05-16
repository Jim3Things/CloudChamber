import React from 'react'

import './App.css'
import {getErrorDetails, logon, logout} from "./proxies/Session"
import {MainPage} from "./MainPage/Main"
import {Login} from "./MainPage/Login"
import {LogEntry, LogProxy} from "./proxies/LogProxy"
import {ErrorSnackbar} from "./common/Snackbar"
import {WatchProxy} from "./proxies/WatchProxy"
import {
    hasSession,
    logonSlice,
    logSlice,
    snackbarSelector,
    snackbarSlice,
    stepperSlice,
    useAppDispatch,
    useAppSelector
} from "./store/Store"
import {RenderIf} from "./common/If"

const logProxy = new LogProxy()
const watchProxy = new WatchProxy()

function App() {
    const dispatch = useAppDispatch()

    const snackText = useAppSelector(snackbarSelector)
    const activeSession = useAppSelector(hasSession)

    // Initiate a login to a session
    const onLogon = (name: string, password: string) => {
        logon(name, password)
            .then(value => {
                // It worked, so record the session state, and start the
                // background calls to get the next tick
                dispatch(logonSlice.actions.logon(value))

                logProxy.start((toHold: number, events: LogEntry[]) =>
                    dispatch(logSlice.actions.append(toHold, events)))

                watchProxy.start((cur) =>
                    dispatch(stepperSlice.actions.update(cur)))
            })
            .catch(msg => getErrorDetails(msg, details => dispatch(logonSlice.actions.loginFailure(details))))
    }

    // Initiate a logout from the active session
    const onLogoutEvent = (name: string) => {
        logout(name)
            .then(() => {
                // We're logged out.  Set the state and cancel the
                // background calls for the next tick
                dispatch(logonSlice.actions.logout())

                logProxy.cancelUpdates()
                watchProxy.cancel()
            })
            .catch(() => {
                dispatch(logonSlice.actions.logout())

                logProxy.cancelUpdates()
                watchProxy.cancel()
            })
    }

    return <div className="App">
        <link rel="stylesheet"
              href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap"/>

        <RenderIf cond={activeSession}>
            <MainPage onLogout={onLogoutEvent}/>
        </RenderIf>

        <RenderIf cond={!activeSession}>
            <Login onClose={onLogon}/>
        </RenderIf>

        <ErrorSnackbar
            open={snackText !== ""}
            onClose={() => dispatch(snackbarSlice.actions.clear())}
            autoHideDuration={4000}
            message={snackText}
        />

    </div>
}

export default App
