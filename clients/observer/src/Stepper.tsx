import React from 'react'
import {IconButton} from "@material-ui/core"
import {FastForward, Pause, PlayArrow, SkipNextOutlined} from '@material-ui/icons'

import './App.css'
import {changeStepperPolicy, SetStepperPolicy} from "./proxies/StepperProxy"
import {curSelector, snackbarSlice, useAppDispatch} from "./store/Store"
import {useSelector} from "react-redux"

export function Stepper(props: {
    disabled: boolean,
}) {
    const dispatch = useAppDispatch()

    const cur = useSelector(curSelector)

    const notify = (policy: SetStepperPolicy) => {
        changeStepperPolicy(
            (msg: string) => dispatch(snackbarSlice.actions.update(msg)),
            policy,
            cur)
    }

    return (
        <div>
            <IconButton disabled={props.disabled} color="inherit"
                        onClick={() => notify(SetStepperPolicy.Pause)}><Pause/></IconButton>
            <IconButton disabled={props.disabled} color="inherit"
                        onClick={() => notify(SetStepperPolicy.Step)}><SkipNextOutlined/></IconButton>
            <IconButton disabled={props.disabled} color="inherit"
                        onClick={() => notify(SetStepperPolicy.Run)}><PlayArrow/></IconButton>
            <IconButton disabled={props.disabled} color="inherit"
                        onClick={() => notify(SetStepperPolicy.Faster)}><FastForward/></IconButton>
        </div>
    )
}
