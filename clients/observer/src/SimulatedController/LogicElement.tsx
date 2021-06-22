// This module provides the common look and feel for logic subsystems in the
// controller.

import { Avatar, Chip } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import React from "react";
import { LogicIcon } from '../common/Icons';
import { impactsSlice, useAppDispatch } from '../store/Store';
import { Impact, Impacts, impactToColor } from './Constants';

interface styleProps {
    impact: Impact
}

const useStyles = makeStyles((theme) => ({
    card: (props: styleProps) => ({
        display: 'flex',
        justifyContent: 'center',
        flexWrap: 'wrap',
        pb: theme.spacing(2),
        backgroundColor: impactToColor(props.impact, theme),
        '& > *': {
            margin: theme.spacing(0.5),
        },
    }),
}))

// Display a logic element in the simulated controller.  Much of the internals
// of this function are temporary, and serve only to provide a trial framing
// for the controller structure.
export function LogicElement(props: {
    title: string,
    impact: Impact,
    selectedImpacts: Impacts
}) {
    const dispatch = useAppDispatch()

    const clickEvent = (ev: React.MouseEvent<HTMLDivElement, MouseEvent>): void => {
        if (props.impact === Impact.ImpactNone) {
            dispatch(impactsSlice.actions.update(props.selectedImpacts))
        } else {
            dispatch(impactsSlice.actions.clear())
        }
    }

    const classes = useStyles({ impact: props.impact })

    return <Chip
        className={classes.card}
        variant="outlined"
        color="primary"
        label={props.title}
        onClick={clickEvent}
        avatar = {
        <Avatar>
            <LogicIcon />
        </Avatar>
        }
    />
}
