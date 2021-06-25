// This module contains the redux store definitions and access functions.

import { combineReducers, configureStore, createSlice, PayloadAction } from "@reduxjs/toolkit"
import { TypedUseSelectorHook, useDispatch, useSelector } from "react-redux"
import { Organizer } from "../Log/Organizer"
import { LogEntry } from "../proxies/LogProxy"
import { CreateSessionUser, SessionUser } from "../proxies/Session"
import { StepperMode, TimeContext } from "../proxies/StepperProxy"
import { SettingsState } from "../Settings"
import { Impacts, NoImpacts } from "../SimulatedController/Constants"


// The store consists of slices associated with:
//   - simulated time (stepper),
//   - the current server session (logon),
//   - the display settings (settings),
//   - the error alert bar (snackbar),
//   - the event log and display,
//   - the controller element impact map,
//
// Each section has the definition for the slice, and the retrieval functions
// used to select specific information from that slice.
//
// Note that the actual store schema is no used outside of this module.

// +++ Simulated time slice
export const stepperSlice = createSlice({
    name: "stepper",
    initialState: {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    },
    reducers: {
        // Update the simulated time
        update: {
            reducer: (state, action: PayloadAction<TimeContext>) => {
                state.now = action.payload.now
                state.mode = action.payload.mode
                state.rate = action.payload.rate
            },
            prepare: (newTime: TimeContext) => {
                return {
                    payload: newTime
                }
            }
        }
    }
})

// Get the simulated time
export const curSelector = (state: StoreSchema) => state.cur

// --- Simulated time slice

// +++ Current user session slice

export const logonSlice = createSlice({
    name: "logon",
    initialState: {
        hasUser: false,
        user: CreateSessionUser({}, ""),
        error: ""
    },
    reducers: {
        // Record a successful login
        logon: {
            reducer: (state, action: PayloadAction<SessionUser>) => {
                state.user = action.payload
                state.hasUser = true
                state.error = ""
            },
            prepare: (user: SessionUser) => {
                return {
                    payload: user
                }
            }
        },

        // Record a failed login
        loginFailure: {
            reducer: (state, action: PayloadAction<string>) => {
                state.hasUser = false
                state.error = action.payload
            },
            prepare: (msg: string) => {
                return {
                    payload: msg
                }
            }
        },

        // Record a logout
        logout: (state) => {
            state.hasUser = false
            state.user = CreateSessionUser({}, "")
            state.error = ""
        }
    }
})

// Get the logged in user details, or undefined if not logged in
export const sessionUserSelector = (state: StoreSchema) => state.session.hasUser ? state.session.user : undefined

// true, if there is an active logged in user
export const hasSession = (state: StoreSchema) => state.session.hasUser

// Get the last login failure
export const logonErrorSelector = (state: StoreSchema) => state.session.error

// --- Current user session slice

// +++ Display option settings slice

export const settingsSlice = createSlice({
    name: "settings",
    initialState: {
        logSettings: {
            showDebug: true,
            showInfra: true,
        }
    },
    reducers: {
        // Update the display options
        update: {
            reducer: (state, action: PayloadAction<SettingsState>) => {
                state.logSettings = action.payload.logSettings
            },
            prepare: (newSetting: SettingsState) => {
                return {
                    payload: newSetting
                }
            }
        }
    }
})

// Get the display options
export const settingsSelector = (state: StoreSchema) => state.settings

// --- Display option settings

// +++ Error alert bar slice

export const snackbarSlice = createSlice({
    name: "snack",
    initialState: {
        msg: ""
    },
    reducers: {
        // set an alert message
        update: {
            reducer: (state, action: PayloadAction<string>) => {
                state.msg = action.payload
            },
            prepare: (msg: string) => {
                return {
                    payload: msg
                }
            }
        },

        // Remove an alert message
        clear: (state) => {
            state.msg = ""
        }
    }
})

// Get the alert message text, if any
export const snackbarSelector = (state: StoreSchema) => state.snackText.msg

// --- Error alert bar slice

// +++ Simulation log tracking slice

// The ogranizer is not held in the redux store itself, but is generated when
// the log slice is selected.  Therefore, the form of the data in the store is
// in logStoreData, and the type returned is logStore.

interface logStoreData {
    entries: LogEntry[],
}

interface logStore {
    entries: LogEntry[],
    organizer: Organizer
}

const initialState: logStoreData = {
    entries: [],
}

export const logSlice = createSlice({
    name: "log",
    initialState: initialState,
    reducers: {
        // append new log entries and update the organizer indices
        append: {
            reducer: (state, action: PayloadAction<{
                toHold: number,
                entries: LogEntry[]
            }>) => {
                const newEntries = state.entries.concat(action.payload.entries)
                const start = Math.max(newEntries.length - action.payload.toHold, 0)
                const slice = newEntries.slice(start)

                state.entries = slice
            },
            prepare: (toHold: number, entries: LogEntry[]) => {
                return {
                    payload: {
                        toHold: toHold,
                        entries: entries
                    }
                }
            }
        },

        // flip the expansion flag for a specific span
        flip: {
            reducer: (state, action: PayloadAction<string>) => {
                state.entries = state.entries.map((v) => {
                    if (v.entry.spanID === action.payload) {
                        v.expanded = !v.expanded
                    }

                    return v
                })
            },
            prepare: (key: string) => {
                return {payload: key}
            }
        }
    }
})

// get the current set of log entries
export const logSelector = (state: StoreSchema): logStore => {
    return {
        entries: state.log.entries,
        organizer: new Organizer(state.log.entries),
    }
}

// --- Simulation log tracking slice

// +++ Controller impacts slice

export const impactsSlice = createSlice({
    name: "impacts",
    initialState: {
        controllerImpacts: NoImpacts
    },
    reducers: {
        // replace the impact claims with the newly provided ones
        update: {
            reducer: (state, action: PayloadAction<Impacts>) => {
                state.controllerImpacts = action.payload
            },
            prepare: (impacts: Impacts) => {
                return { payload: impacts }
            }
        },

        // clear all impact claims
        clear: (state) => {
            state.controllerImpacts = NoImpacts
        }
    }
})

export const impactsSelector = (state: StoreSchema) => state.impacts

// --- controller impacts slice

const rootReducer = combineReducers({
    cur: stepperSlice.reducer,
    settings: settingsSlice.reducer,
    snackText: snackbarSlice.reducer,
    session: logonSlice.reducer,
    log: logSlice.reducer,
    impacts: impactsSlice.reducer
})

export const store = configureStore({
    reducer: rootReducer
})

type StoreSchema = ReturnType<typeof rootReducer>
type AppDispatch = typeof store.dispatch

export const useAppDispatch = () => useDispatch<AppDispatch>()
export const useAppSelector: TypedUseSelectorHook<StoreSchema> = useSelector
