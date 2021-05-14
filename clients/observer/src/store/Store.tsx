// Define the redux store here.

import {
    combineReducers,
    configureStore,
    createSelector,
    createSlice, PayloadAction
} from '@reduxjs/toolkit'

import {StepperMode, TimeContext} from "../proxies/StepperProxy"
import {useDispatch} from "react-redux"
import {SettingsState} from "../Settings"
import {CreateSessionUser, SessionUser} from "../proxies/Session"

export const stepperSlice = createSlice({
    name: "stepper",
    initialState: {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    },
    reducers: {
        updatePolicy: {
            reducer: (state, action: PayloadAction<TimeContext>) => {
                state.now = action.payload.now
                state.mode = action.payload.mode
                state.rate = action.payload.rate
            },
            prepare: (newTime: TimeContext) => {
                return {
                    payload: {
                        now: newTime.now,
                        rate: newTime.rate,
                        mode: newTime.mode
                    }
                }
            }
        }
    }
})

export const logonSlice = createSlice({
    name: "logon",
    initialState: {
        hasUser: false,
        user: CreateSessionUser({}, "", ""),
        error: ""
    },
    reducers: {
        logon: {
            reducer: (state, action: PayloadAction<SessionUser>) => {
                state.user = action.payload
                state.hasUser = true
                state.error = ""
            },
            prepare: (user: SessionUser) => {
                return {
                    payload:  user
                }
            }
        },
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
        logout: (state) => {
            state.hasUser = false
            state.user = CreateSessionUser({}, "", "")
            state.error = ""
        }
    }
})

export const settingsSlice = createSlice ({
    name: "settings",
    initialState: {
        logSettings: {
            showDebug: true,
            showInfra: true,
        }
    },
    reducers: {
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

export const snackbarSlice = createSlice({
    name: "snack",
    initialState: {
        msg: ""
    },
    reducers: {
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
        clear: (state) => {
            state.msg = ""
        }
    }
})

const rootReducer = combineReducers({
    cur: stepperSlice.reducer,
    settings: settingsSlice.reducer,
    snackText: snackbarSlice.reducer,
    session: logonSlice.reducer
})

export type StoreSchema = ReturnType<typeof rootReducer>

export const store = configureStore({
    reducer: rootReducer
})

export const curSelector = createSelector(
    (state: StoreSchema) => state.cur,
    (cur) => cur
)

export const settingsSelector = createSelector(
    (state: StoreSchema) => state.settings,
    (settings) => settings
)

export const snackbarSelector = createSelector(
    (state: StoreSchema) => state.snackText.msg,
    (msg) => msg
)

export const sessionUserSelector = createSelector(
    (state: StoreSchema) => state.session,
    (session) => session.hasUser ? session.user : undefined
)

export const logonErrorSelector = createSelector(
    (state: StoreSchema) => state.session,
    (session) => session.error
)

export const hasSession = createSelector(
    (state: StoreSchema) => state.session.hasUser,
    (hasUser) => hasUser
)

export type AppDispatch = typeof store.dispatch
export const useAppDispatch = () => useDispatch<AppDispatch>()
