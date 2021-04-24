import React from 'react';
import {IconButton} from "@material-ui/core";
import {FastForward, Pause, PlayArrow, SkipNextOutlined} from '@material-ui/icons';

import './App.css';
import {SetStepperPolicy} from "./proxies/StepperProxy";

export function Stepper(props: {onPolicyEvent?: (policy: SetStepperPolicy) => any})  {
    const notify = (policy: SetStepperPolicy) => {
        if (props.onPolicyEvent) {
            props.onPolicyEvent(policy);
        }
    }

    return (
        <div>
            <IconButton color="inherit" onClick={() => notify(SetStepperPolicy.Pause)}><Pause/></IconButton>
            <IconButton color="inherit" onClick={() => notify(SetStepperPolicy.Step)}><SkipNextOutlined/></IconButton>
            <IconButton color="inherit" onClick={() => notify(SetStepperPolicy.Run)}><PlayArrow/></IconButton>
            <IconButton color="inherit" onClick={() => notify(SetStepperPolicy.Faster)}><FastForward/></IconButton>
        </div>
    );
}
