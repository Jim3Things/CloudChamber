import React, {FunctionComponent} from 'react';
import {Badge, Toolbar} from "@material-ui/core";
import {Pause, PlayArrow} from "@material-ui/icons";
import {makeStyles} from "@material-ui/core/styles";

import {StepperMode, Timestamp} from "./proxies/StepperProxy";

const useStyles = makeStyles((theme) => ({
    root: {
        flexGrow: 1,
    },
    iconTag: {
        fontSize: "small"
    }
}));

// This method constructs a status bar containing currently running summary information

// TODO: This currently only has information on the stepper status, others need to be
//       added as they make sense.

export const StatusBar: FunctionComponent<{ cur: Timestamp }> = ({cur}) => {
    const classes = useStyles();

    // Helpers to control visibility of the different type of execution icons
    const hideBadge = (cur: Timestamp) => (cur.mode !== StepperMode.Running) || (cur.rate <= 1)
    const hidePlay = (cur: Timestamp) => (cur.mode !== StepperMode.Running) || (cur.rate !== 1)
    const hidePause = (cur: Timestamp) => (cur.mode !== StepperMode.Paused)

    // Construct the badge text as the rate - e.g. "1x", "2x", etc
    const badgeString = (cur: Timestamp) => "" + cur.rate + "x"

    // Helper to construct the faster icon
    const badgeIcon = (cur: Timestamp) => {
        return (
            <div hidden={hideBadge(cur)}>
                <Badge anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'right',
                }} badgeContent={badgeString(cur)}>
                    <PlayArrow className={classes.iconTag}/>
                </Badge>
            </div>
        )
    }

    // Helper to construct the simple run-at-1x icon
    const playIcon = (cur: Timestamp) =>
        <div hidden={hidePlay(cur)}>
            <PlayArrow className={classes.iconTag}/>
        </div>

    // Helper to construct the 'currently paused' icon, used for both the
    // pause and single-step stepper actions
    const pauseIcon = (cur: Timestamp) =>
        <div hidden={hidePause(cur)}>
            <Pause className={classes.iconTag}/>
        </div>

    return (
        <div className={classes.root}>
        <Toolbar variant="dense">
            <div className={classes.root}/>
            <div>
                {badgeIcon(cur)}
                {playIcon(cur)}
                {pauseIcon(cur)}
            </div>
            &nbsp;&nbsp;
            {cur.now}
        </Toolbar>
        </div>
    );
}
