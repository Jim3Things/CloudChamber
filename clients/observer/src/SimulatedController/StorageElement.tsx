// This module provides the common look and feel for stores in the controller.

import { Avatar, Chip } from "@material-ui/core";
import { makeStyles } from '@material-ui/core/styles';
import { StorageIcon } from "../common/Icons";
import { Impact, impactToColor } from "./Constants";

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

// Display a storage element in the simulated controller.  Much of the internals
// of this function are temporary, and serve only to provide a trial framing
// for the controller structure.
export function StorageElement(props: {
    title: string,
    impact: Impact
}) {
    const classes = useStyles({impact: props.impact})

    return <Chip
        className={classes.card}
        variant="outlined"
        color="default"
        avatar={
                <Avatar>
                    <StorageIcon />
                </Avatar>
            }
            label={props.title}
        />
}
