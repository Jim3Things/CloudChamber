import React from 'react';
import {IconButton} from "@material-ui/core";
import {FastForward, Pause, PlayArrow, SkipNextOutlined} from '@material-ui/icons';

import './App.css';
import {SetStepperPolicy} from "./proxies/StepperProxy";

export function Stepper(props: {
    disabled: boolean,
    onPolicyEvent?: (policy: SetStepperPolicy) => any})  {
    const notify = (policy: SetStepperPolicy) => {
        if (props.onPolicyEvent) {
            props.onPolicyEvent(policy);
        }
    }

    return (
        <div>
            <IconButton disabled={props.disabled} color="inherit" onClick={() => notify(SetStepperPolicy.Pause)}><Pause/></IconButton>
            <IconButton disabled={props.disabled} color="inherit" onClick={() => notify(SetStepperPolicy.Step)}><SkipNextOutlined/></IconButton>
            <IconButton disabled={props.disabled} color="inherit" onClick={() => notify(SetStepperPolicy.Run)}><PlayArrow/></IconButton>
            <IconButton disabled={props.disabled} color="inherit" onClick={() => notify(SetStepperPolicy.Faster)}><FastForward/></IconButton>
        </div>
    );
}
