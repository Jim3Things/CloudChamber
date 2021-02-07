using System;
using System.Collections.Generic;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;

namespace CloudChamber.Cmdlets.Protos
{
    public class SimulationStatus
    {
        [JsonProperty("frontEndStartedAt")]
        [JsonConverter(typeof(IsoDateTimeConverter))]
        public DateTime FrontEndStartedAt { get; set; }

        [JsonProperty("inactivityTimeout")]
        [JsonConverter(typeof(DurationConverter))]
        public TimeSpan InactivityTimeout { get; set; }
    }

    public class SessionList
    {
        [JsonProperty("sessions")] public List<SessionEntry> Sessions { get; set; }
    }

    public class SessionEntry
    {
        [JsonProperty("id")] public long Id { get; set; }

        [JsonProperty("uri")] public Uri Uri { get; set; }
    }

    public class ClusterSession
    {
        [JsonProperty("userName")] public string Name { get; set; }

        [JsonProperty("timeout")]
        [JsonConverter(typeof(IsoDateTimeConverter))]
        public DateTime Expiry { get; set; }
    }
}
