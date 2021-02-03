import React from 'react';
import {
    List, ListItem, ListItemIcon, ListItemText,
    Paper} from "@material-ui/core";
import {makeStyles} from "@material-ui/core/styles";
import {
    Menu,
    Warning,
    Error,
    ErrorOutline,
    HelpOutline,
    BugReport,
    Info,
    ExpandMore,
    ExpandLess
} from '@material-ui/icons';

import {
    LogEntry,
    LogEvent,
    LogEventType,
    LogSeverity
} from "../proxies/LogProxy";
import {Organizer} from "./Organizer";
import {RenderIf} from "../common/If";
import {SettingsState} from "../Settings";

interface styleProps {
    indent: number
    infra: boolean
    height: number
}

const useStyles = makeStyles((theme) => ({
    root: (props: styleProps) => ({
        maxHeight: props.height,
        minHeight: props.height,
        overflow: "auto",
        fontSize: "small",
        pt: 0,
        pb: 0
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
        fontSize: "small",
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.success.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    }),
    warning: (props: styleProps) => ({
        fontSize: "small",
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.warning.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    }),
    error: (props: styleProps) => ({
        fontSize: "small",
        color: (props.infra
            ? theme.palette.grey.A400
            : theme.palette.error.main),
        backgroundColor: (props.infra
            ? theme.palette.action.hover
            : theme.palette.background.paper)
    })
}));

export interface ExpansionHandler {
    (id: string): void
}

const ExpandIcon = (props: { expanded: boolean }) =>
    props.expanded ? <ExpandLess/> : <ExpandMore/>

// FilteredCount produces the child element count for a span after applying the
// active display filters.
function FilteredCount(
    event: LogEvent[],
    settings: SettingsState,
    organizer: Organizer
): number {
    if (settings.logSettings.showDebug && settings.logSettings.showInfra) {
        return event.length
    }

    let count = event.length;

    for (const item of event) {
        switch (item.eventAction) {
            case LogEventType.SpanStart:
                if (!settings.logSettings.showInfra) {
                    const span = organizer.get(item.spanId)

                    if ((span !== undefined) && span.infrastructure) {
                        count -= 1
                    }
                }
                break

            case LogEventType.AddLink:
                if (!settings.logSettings.showInfra) {
                    const span = organizer.get(item.linkId)

                    if ((span !== undefined) && span.infrastructure) {
                        count -= 1
                    }
                }
                break

            case LogEventType.Trace:
                if (!settings.logSettings.showDebug && (item.severity === LogSeverity.Debug)) {
                    count -= 1
                }
                break
        }
    }

    return count
}

// TraceSpanElement provides the list entry for a trace span element
function TraceSpanElement(props: {
    text: string,
    reason: string | null,
    expanded: boolean,
    expandable: boolean,
    onTrackChange: ExpansionHandler,
    indent: number,
    infra: boolean,
    id: string
}) {
    const classes = useStyles({infra: props.infra, indent: props.indent, height: 0})

    return <ListItem dense button onClick={() => props.onTrackChange(props.id)} className={classes.nested}>
        <ListItemIcon>
            <Menu/>
        </ListItemIcon>
        <ListItemText className={classes.labelText} primary={props.text} secondary={props.reason}/>
        <RenderIf cond={props.expandable}>
            <ExpandIcon expanded={props.expanded}/>
        </RenderIf>
    </ListItem>
}

// TraceSpanSubtree draws the entry for the span, and all trace events within that
// span.  It recursively executes to draw all sub-spans, as they are encountered.
function TraceSpanSubtree(props: {
    organizer: Organizer,
    settings: SettingsState,
    onTrackChange: ExpansionHandler,
    indent: number,
    id: string
}) {
    const entry = props.organizer.get(props.id)

    if (entry === undefined) {
        return <TraceSpanElement
            text="Missing"
            reason={null}
            expanded={false}
            expandable={false}
            onTrackChange={props.onTrackChange}
            indent={props.indent}
            infra={false}
            id={props.id}/>
    }

    if (entry.infrastructure && !props.settings.logSettings.showInfra) {
        return <React.Fragment/>
    }

    const canExpand = FilteredCount(entry.event, props.settings, props.organizer) > 0
    const isExpanded = canExpand && props.organizer.isExpanded(props.id)

    return <React.Fragment>
        <TraceSpanElement
            text={entry.name}
            reason={entry.reason}
            expanded={isExpanded}
            expandable={canExpand}
            onTrackChange={props.onTrackChange}
            indent={props.indent}
            infra={entry.infrastructure}
            id={props.id}
        />

        <RenderIf cond={isExpanded}>
            {entry.event.map((ev) => {
                return <TraceEvent
                    organizer={props.organizer}
                    settings={props.settings}
                    onTrackChange={props.onTrackChange}
                    indent={props.indent + 4}
                    event={ev}
                    entry={entry}
                    infra={entry.infrastructure}
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
    onTrackChange: ExpansionHandler,
    indent: number,
    entry: LogEntry,
    event: LogEvent,
    infra: boolean
}) {
    const classes = useStyles({indent: props.indent, infra: props.infra, height: 0})

    // Construct the icon to use to denote the severity code in a log event.
    const SevIcon = (props: { sev: number }) => {
        switch (props.sev) {
            case LogSeverity.Debug:
                return <BugReport className={classes.success}/>

            case LogSeverity.Info:
                return <Info className={classes.success}/>

            case LogSeverity.Warning:
                return <Warning className={classes.warning}/>

            case LogSeverity.Error:
                return <ErrorOutline className={classes.error}/>

            case LogSeverity.Fatal:
                return <Error className={classes.error}/>

            default:
                return <HelpOutline className={classes.error}/>
        }
    }

    switch (props.event.eventAction) {
        case LogEventType.SpanStart: {
            // If this event is for a span start, then recursively process that span,
            // unless that child span has already been aged out.
            const span = props.organizer.get(props.event.spanId)
            if (span === undefined) {
                return <TraceSpanElement
                    expanded={false}
                    expandable={false}
                    onTrackChange={props.onTrackChange}
                    text={"Missing:" + props.event.spanId}
                    reason="Could not find this log entry"
                    indent={props.indent}
                    infra={props.infra}
                    id={props.event.spanId}
                />
            }

            if (!props.settings.logSettings.showInfra && span.infrastructure) {
                return <React.Fragment/>
            }

            return <TraceSpanSubtree
                organizer={props.organizer}
                settings={props.settings}
                onTrackChange={props.onTrackChange}
                indent={props.indent}
                id={props.event.spanId}/>
        }

        case LogEventType.AddLink: {
            // If this event is for an add link, then see if that link has been
            // registered.  If it has, then recursively process the registering
            // span, if it can be found.  If it can't, then just ignore it.
            const span = props.organizer.getViaLink(props.entry.spanID, props.entry.traceID, props.event.linkId)

            if (span !== undefined) {
                return <TraceSpanSubtree
                    organizer={props.organizer}
                    settings={props.settings}
                    onTrackChange={props.onTrackChange}
                    indent={props.indent}
                    id={span}/>
            }

            return <React.Fragment/>
        }

        default:
            // Format a normal trace event.
            if (!props.settings.logSettings.showDebug && props.event.severity === LogSeverity.Debug) {
                return <React.Fragment/>
            }

            return <ListItem dense className={classes.nested}>
                <ListItemIcon>
                    <SevIcon sev={props.event.severity}/>
                </ListItemIcon>
                <ListItemText className={classes.labelText} primary={props.event.text} />
            </ListItem>
    }
}

export function LogDisplay(props: {
    height: number,
    organizer: Organizer,
    settings: SettingsState,
    onTrackChange: ExpansionHandler
}) {
    const classes = useStyles({indent: 0, infra: false, height: props.height})

    return (
        <Paper variant="outlined" className={classes.root}>
            <List dense disablePadding>
                {props.organizer.roots.map((key) => {
                    return <TraceSpanSubtree
                        settings={props.settings}
                        organizer={props.organizer}
                        onTrackChange={props.onTrackChange}
                        indent={0}
                        id={key}/>
                })}
            </List>
        </Paper>
    );
}
