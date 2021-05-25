import React from 'react'
import {List, ListItem, ListItemIcon, ListItemText, Paper} from "@material-ui/core"
import {makeStyles} from "@material-ui/core/styles"
import {BugReport, Error, ErrorOutline, HelpOutline, Info, Warning} from '@material-ui/icons'

import {Organizer} from "./Organizer"
import {MoreOrLess, RenderIf} from "../common/If"
import {SettingsState} from "../Settings"
import {Action, Event, Severity} from "../pkg/protos/log/entry"
import {GetAfterResponse_traceEntry} from "../pkg/protos/services/requests"
import {logSelector, logSlice, settingsSelector, useAppDispatch, useAppSelector} from "../store/Store"

interface styleProps {
    indent: number
    infra: boolean
    height: number
    size: string
}

const useStyles = makeStyles((theme) => ({
    root: (props: styleProps) => ({
        maxHeight: props.height,
        minHeight: props.height,
        overflow: "auto",
        fontSize: "small",
        pt: 0,
        pb: 0,
    }),
    nested: (props: styleProps) => ({
        fontSize: "small",
        paddingLeft: theme.spacing(props.indent),
        pt: 0,
        pb: 0
    }),
    labelText: (props: styleProps) => ({
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper),
        whiteSpace: "pre-wrap",
    }),
    success: (props: styleProps) => ({
        fontSize: props.size,
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.success.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    }),
    warning: (props: styleProps) => ({
        fontSize: props.size,
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.warning.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    }),
    error: (props: styleProps) => ({
        fontSize: props.size,
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.error.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    })
}))

export interface ExpansionHandler {
    (id: string): void
}

// Construct the icon to use to denote the severity code in a log event.
function SevIcon(props: {
    indent: number,
    severity: Severity,
    iconSize: string,
    infra: boolean }) {
    const classes = useStyles({indent: props.indent, infra: props.infra, height: 0, size: props.iconSize})

    switch (props.severity) {
        case Severity.Debug:
            return <BugReport className={classes.success}/>

        case Severity.Info:
            return <Info className={classes.success}/>

        case Severity.Warning:
            return <Warning className={classes.warning}/>

        case Severity.Error:
            return <ErrorOutline className={classes.error}/>

        case Severity.Fatal:
            return <Error className={classes.error}/>

        default:
            return <HelpOutline className={classes.error}/>
    }
}


// FilteredCount produces the child element count for a span after applying the
// active display filters.
function FilteredCount(
    event: Event[],
    settings: SettingsState,
    organizer: Organizer
): number {
    if (settings.logSettings.showDebug && settings.logSettings.showInfra) {
        return event.length
    }

    let count = event.length

    for (const item of event) {
        switch (item.eventAction) {
            case Action.SpanStart:
                if (!settings.logSettings.showInfra) {
                    const span = organizer.get(item.spanId)

                    if ((span !== undefined) && span.entry.infrastructure) {
                        count -= 1
                    }
                }
                break

            case Action.AddLink:
                if (!settings.logSettings.showInfra) {
                    const span = organizer.get(item.linkId)

                    if ((span !== undefined) && span.entry.infrastructure) {
                        count -= 1
                    }
                }
                break

            case Action.Trace:
                if (!settings.logSettings.showDebug && (item.severity === Severity.Debug)) {
                    count -= 1
                }
                break
        }
    }

    return count
}

// TraceSpanElement provides the list entry for a trace span element
function TraceSpanElement(props: {
    severity: Severity,
    text: string,
    reason: string | null,
    expanded: boolean,
    expandable: boolean,
    indent: number,
    infra: boolean,
    id: string
}) {
    const classes = useStyles({infra: props.infra, indent: props.indent, height: 0, size: "small"})
    const dispatch = useAppDispatch()

    return <ListItem dense button onClick={() => dispatch(logSlice.actions.flip(props.id))} className={classes.nested}>
        <ListItemIcon>
            <SevIcon
                indent={props.indent}
                infra={props.infra}
                severity={props.severity}
                iconSize="medium"
            />
        </ListItemIcon>
        <ListItemText className={classes.labelText} primary={props.text} secondary={props.reason}/>
        <RenderIf cond={props.expandable}>
            <MoreOrLess cond={props.expanded}/>
        </RenderIf>
    </ListItem>
}

// TraceSpanSubtree draws the entry for the span, and all trace events within that
// span.  It recursively executes to draw all sub-spans, as they are encountered.
function TraceSpanSubtree(props: {
    organizer: Organizer,
    settings: SettingsState,
    indent: number,
    id: string
}) {
    const entry = props.organizer.get(props.id)

    if (entry === undefined) {
        return <TraceSpanElement
            severity={Severity.Debug}
            text="Missing"
            reason={null}
            expanded={false}
            expandable={false}
            indent={props.indent}
            infra={false}
            id={props.id}/>
    }

    if (entry.entry.infrastructure && !props.settings.logSettings.showInfra) {
        return <React.Fragment/>
    }

    const canExpand = FilteredCount(entry.entry.event, props.settings, props.organizer) > 0
    const isExpanded = canExpand && props.organizer.isExpanded(props.id)

    return <React.Fragment>
        <TraceSpanElement
            severity={entry.maxSeverity}
            text={entry.entry.name}
            reason={entry.entry.reason}
            expanded={isExpanded}
            expandable={canExpand}
            indent={props.indent}
            infra={entry.entry.infrastructure}
            id={props.id}
        />

        <RenderIf cond={isExpanded}>
            {entry.entry.event.map((ev) => {
                return <TraceEvent
                    organizer={props.organizer}
                    settings={props.settings}
                    indent={props.indent + 4}
                    event={ev}
                    entry={entry}
                    infra={entry.entry.infrastructure}
                />
            })}
        </RenderIf>
    </React.Fragment>
}

// TraceEvent draws the list entry for an individual trace event, or to start
// the processing for a child span.
function TraceEvent(props: {
    organizer: Organizer,
    settings: SettingsState,
    indent: number,
    entry: GetAfterResponse_traceEntry,
    event: Event,
    infra: boolean
}) {
    const classes = useStyles({indent: props.indent, infra: props.infra, height: 0, size: "small"})

    switch (props.event.eventAction) {
        case Action.SpanStart: {
            // If this event is for a span start, then recursively process that span,
            // unless that child span has already been aged out.
            const span = props.organizer.get(props.event.spanId)
            if (span === undefined) {
                return <TraceSpanElement
                    severity={Severity.Debug}
                    expanded={false}
                    expandable={false}
                    text={"Missing:" + props.event.spanId}
                    reason="Could not find this log entry"
                    indent={props.indent}
                    infra={props.infra}
                    id={props.event.spanId}
                />
            }

            if (!props.settings.logSettings.showInfra && span.entry.infrastructure) {
                return <React.Fragment/>
            }

            return <TraceSpanSubtree
                organizer={props.organizer}
                settings={props.settings}
                indent={props.indent}
                id={props.event.spanId}/>
        }

        case Action.AddLink: {
            // If this event is for an add link, then see if that link has been
            // registered.  If it has, then recursively process the registering
            // span, if it can be found.  If it can't, then just ignore it.
            const span = props.organizer.getViaLink(props.entry.entry.spanID, props.entry.entry.traceID, props.event.linkId)

            if (span !== undefined) {
                return <TraceSpanSubtree
                    organizer={props.organizer}
                    settings={props.settings}
                    indent={props.indent}
                    id={span}/>
            }

            return <React.Fragment/>
        }

        default:
            // Format a normal trace event.
            if (!props.settings.logSettings.showDebug && props.event.severity === Severity.Debug) {
                return <React.Fragment/>
            }

            return <ListItem dense className={classes.nested}>
                <ListItemIcon>
                    <SevIcon
                        indent={props.indent}
                        infra={props.infra}
                        severity={props.event.severity}
                        iconSize="small"
                    />
                </ListItemIcon>
                <ListItemText className={classes.labelText} primary={props.event.text}/>
            </ListItem>
    }
}

export function LogDisplay(props: {
    matchId: string}) {
    const elem = document.getElementById(props.matchId)
    const height = elem !== null ? (elem.offsetHeight - 5) : 500

    const classes = useStyles({indent: 0, infra: false, height: height, size: "small"})

    const settings = useAppSelector(settingsSelector)
    const logData = useAppSelector(logSelector)

    return (
        <Paper variant="outlined" className={classes.root}>
            <List dense disablePadding>
                {logData.organizer.roots.map((key) => {
                    return <TraceSpanSubtree
                        settings={settings}
                        organizer={logData.organizer}
                        indent={1}
                        id={key}/>
                })}
            </List>
        </Paper>
    )
}
