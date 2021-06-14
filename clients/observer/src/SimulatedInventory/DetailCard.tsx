// DetailCard builds the common components for an inventory element details popup.
// This ensures a commonality of visual structure, while allowing the caller to
// specify child elements to handle the pre-element unique items.

import React from "react"
import {
    Card, CardContent, CardHeader,
    createStyles,
    Popover,
    Typography
} from "@material-ui/core"
import {Condition, conditionToString} from "../pkg/protos/inventory/common"
import {makeStyles} from "@material-ui/core/styles"
import {ModalProps} from "@material-ui/core/Modal"

const useStyles = makeStyles((theme) =>
    createStyles({
        popover: {
            pointerEvents: 'none'
        },
        paper: {
            padding: theme.spacing(1),
        },
        root: {
            maxWidth: 600,
        },
    }),
)

export function DetailCard(props: {
    id: string,
    open: boolean,
    elementName: string
    enabled: boolean,
    condition: Condition,
    anchorEl?: Element | null | undefined,
    onClose?: ModalProps['onClose'],
    children? : React.ReactNode
}) {
    const classes = useStyles()

    return <Popover
        id={props.id}
        className={classes.popover}
        classes={{
            paper: classes.paper,
        }}
        open={props.open}
        anchorEl={props.anchorEl}
        anchorOrigin={{
            vertical: 'top',
            horizontal: 'right',
        }}
        transformOrigin={{
            vertical: 'top',
            horizontal: 'left',
        }}
        onClose={props.onClose}
        disableRestoreFocus
    >
        <Card className={classes.root}>
            <CardHeader
                title={"Details for " + props.elementName}>
            </CardHeader>
            <CardContent>
                <Typography paragraph>
                    {props.elementName} is {props.enabled ? "enabled" : "disabled"} for use.
                    It is {conditionToString(props.condition)}.
                </Typography>
                <Typography paragraph>

                </Typography>

                {props.children}

            </CardContent>
        </Card>
    </Popover>

}
