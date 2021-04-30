import React, {ReactNode, useState} from "react"
import {makeStyles, Theme} from "@material-ui/core/styles"
import {IconButton, Popover, Typography} from "@material-ui/core"
import {Variant} from "@material-ui/core/styles/createTypography"

import {MoreOrLess} from "./If"

const useStyles = makeStyles((theme: Theme) => ({
    paper: {
        padding: theme.spacing(1),
    }
}))

// ExpandingLabel is an element that emulates a drop-down display.  It provides
// a text label, and displays the children as a popover based on the expand
// button that this element manages.
export function ExpandingLabel(props: {
    label: string,
    variant?: Variant,
    children?: ReactNode
}) {
    const classes = useStyles()

    const [anchorEl, setAnchorEl] = useState<HTMLButtonElement | null>(null)

    const onClick = (e: React.MouseEvent<HTMLButtonElement | undefined>) => {
        setAnchorEl(open ? null : e.currentTarget)
    }

    const onClose = () => setAnchorEl(null)

    const open = Boolean(anchorEl)

    return <React.Fragment>
        <Typography variant={props.variant}>
            {props.label}
        </Typography>
        <IconButton
            size="small"
            color="inherit"
            onClick={onClick}
            aria-owns={open ? 'detail-popover' : undefined}
            aria-haspopup={true}
        >
            <MoreOrLess cond={!open}/>
        </IconButton>
        <Popover
            id="detail-popover"
            classes={{paper: classes.paper}}
            anchorEl={anchorEl}
            anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'left'
            }}
            transformOrigin={{
                vertical: 'top',
                horizontal: 'left'
            }}
            onClose={onClose}
            open={open}
            disableRestoreFocus
        >
            {props.children}
        </Popover>
    </React.Fragment>
}
