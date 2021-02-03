import React, {FunctionComponent} from 'react';
import {IconButton} from "@material-ui/core";
import {FastForward, Pause, PlayArrow, SkipNextOutlined} from '@material-ui/icons';

import './App.css';
import {StepperPolicy} from "./proxies/StepperProxy";

export const Stepper: FunctionComponent<{onPolicyEvent?: (policy: StepperPolicy) => any}> = ({onPolicyEvent}) => {
    const notify = (policy: StepperPolicy) => {
        if (onPolicyEvent) {
            onPolicyEvent(policy);
        }
    }

    return (
        <div>
            <IconButton color="inherit" onClick={() => notify(StepperPolicy.Pause)}><Pause/></IconButton>
            <IconButton color="inherit" onClick={() => notify(StepperPolicy.Step)}><SkipNextOutlined/></IconButton>
            <IconButton color="inherit" onClick={() => notify(StepperPolicy.Run)}><PlayArrow/></IconButton>
            <IconButton color="inherit" onClick={() => notify(StepperPolicy.Faster)}><FastForward/></IconButton>
        </div>
    );
}
