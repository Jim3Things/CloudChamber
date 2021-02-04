using System.Collections.Generic;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;

namespace CloudChamber.Cmdlets.Protos
{
    public class LogEventModule
    {
        public enum ImpactEnum
        {
            Invalid,
            Read,
            Create,
            Modify,
            Delete,
            Execute,
        }

        [JsonProperty("impact")]
        [JsonConverter(typeof(StringEnumConverter))]
        public ImpactEnum Impact { get; set; }

        [JsonProperty("name")] public string Name { get; set; }
    }

    public class LogEvent
    {
        public enum ActionEnum
        {
            Trace,
            UpdateSpanName,
            UpdateReason,
            SpanStart,
            AddLink
        }

        public enum SeverityEnum
        {
            Debug,
            Reason,
            Info,
            Warning,
            Error,
            Fatal,
        }

        [JsonProperty("tick")] public long Tick { get; set; }

        [JsonProperty("severity")]
        [JsonConverter(typeof(StringEnumConverter))]
        public SeverityEnum Severity { get; set; }

        [JsonProperty("name")] public string Name { get; set; }

        [JsonProperty("text")] public string Text { get; set; }

        [JsonProperty("stackTrace")] public string StackTrace { get; set; }

        [JsonProperty("impacted")] public List<LogEventModule> Impacted { get; set; }

        [JsonProperty("eventAction")]
        [JsonConverter(typeof(StringEnumConverter))]
        public ActionEnum Action { get; set; }

        [JsonProperty("spanId")] public string ChildSpanID { get; set; }

        [JsonProperty("linkId")] public string LinkID { get; set; }
    }

    public class LogEntry
    {
        [JsonProperty("name")] public string Name { get; set; }

        [JsonProperty("spanID")] public string SpanID { get; set; }

        [JsonProperty("parentID")] public string ParentID { get; set; }

        [JsonProperty("traceID")] public string TraceID { get; set; }

        [JsonProperty("status")] public string Status { get; set; }

        [JsonProperty("stackTrace")] public string StackTrace { get; set; }

        [JsonProperty("infrastructure")] public bool Infrastructure { get; set; }

        [JsonProperty("reason")] public string Reason { get; set; }

        [JsonProperty("startingLink")] public string StartingLink { get; set; }

        [JsonProperty("linkSpanID")] public string LinkSpanID { get; set; }

        [JsonProperty("linkTraceID")] public string LinkTraceId { get; set; }

        [JsonProperty("event")] public List<LogEvent> Event { get; set; }
    }

    public class TraceEntry
    {
        [JsonProperty("id")] public long Id { get; set; }

        [JsonProperty("entry")] public LogEntry Entry { get; set; }
    }

    public class Traces
    {
        [JsonProperty("lastId")] public long LastId { get; set; }

        [JsonProperty("entries")] public List<TraceEntry> Entries { get; set; }
    }

    public class TracePolicy
    {
        [JsonProperty("maxEntriesHeld")] public long MaxEntriesHeld { get; set; }

        [JsonProperty("firstId")] public long FirstID { get; set; }
    }
}
