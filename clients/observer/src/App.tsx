import React, {useState} from 'react'
import {useSelector} from "react-redux"

import './App.css'
import {getErrorDetails, logon, logout} from "./proxies/Session"
import {MainPage} from "./MainPage/Main"
import {Login} from "./MainPage/Login"
import {LogProxy} from "./proxies/LogProxy"
import {Organizer} from "./Log/Organizer"
import {GetAfterResponse, GetAfterResponse_traceEntry} from "./pkg/protos/services/requests"
import {ErrorSnackbar} from "./common/Snackbar"
import {WatchProxy} from "./proxies/WatchProxy"
import {
    snackbarSlice, snackbarSelector,
    hasSession, logonSlice,
    useAppDispatch, stepperSlice
} from "./store/Store"
import {RenderIf} from "./common/If"

const logProxy = new LogProxy()
const watchProxy = new WatchProxy()
let organizer = new Organizer([])

function App() {
    const [entries, setEntries] = useState<GetAfterResponse_traceEntry[]>([])
    //const [organizer, setOrganizer] = useState<Organizer>(new Organizer([]))

    const dispatch = useAppDispatch()

    const snackText = useSelector(snackbarSelector)
    const activeSession = useSelector(hasSession)

    // Initiate a login to a session
    const onLogon = (name: string, password: string) => {
        logon(name, password)
            .then(value => {
                // It worked, so record the session state, and start the
                // background calls to get the next tick
                dispatch(logonSlice.actions.logon(value))

                logProxy.start(onNewLogEvent)
                watchProxy.start((cur) => dispatch(stepperSlice.actions.updatePolicy(cur)))
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

    const onNewLogEvent = (toHold: number, events: GetAfterResponse) => {
        setEntries((prev) => {

            const newEntries = prev.concat(events.entries)
            const start = Math.max(newEntries.length - toHold, 0)
            const slice = newEntries.slice(start)

            const newOrg = new Organizer(slice, organizer)
            organizer = newOrg

            return slice
        })

        //setOrganizer(newOrg)
    }

    const onExpansionHandler = (id: string): void => {
        const org = organizer
        org.flip(id)
        //setOrganizer(org)
    }

    return <div className="App">
        <link rel="stylesheet"
              href="https://fonts.googleapis.com/css?family=Roboto:300,400,500,700&display=swap"/>

        <RenderIf cond={activeSession}>
            <MainPage
                onLogout={onLogoutEvent}
                organizer={organizer}
                onTrackChange={onExpansionHandler}
            />
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
