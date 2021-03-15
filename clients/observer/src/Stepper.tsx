import React, {FunctionComponent} from 'react';
import {IconButton} from "@material-ui/core";
import {FastForward, Pause, PlayArrow, SkipNextOutlined} from '@material-ui/icons';

import './App.css';
import {SetStepperPolicy} from "./proxies/StepperProxy";

export const Stepper: FunctionComponent<{onPolicyEvent?: (policy: SetStepperPolicy) => any}> = ({onPolicyEvent}) => {
    const notify = (policy: SetStepperPolicy) => {
        if (onPolicyEvent) {
            onPolicyEvent(policy);
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
