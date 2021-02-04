using System.Collections.Generic;
using System.Management.Automation;
using CloudChamber.Cmdlets.Protos;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets
{
    public class SimulationCmdlets : CmdletBase
    {
        protected SimulationCmdlets() : base("/api/simulation")
        {
        }

        /// <summary>
        ///     Session is the logged-in session to use for the operation.
        /// </summary>
        [Parameter(Mandatory = true)]
        public Session Session { get; set; }
    }

    [Cmdlet(VerbsCommon.Get, Names.Simulation)]
    [OutputType(typeof(SimulationStatus))]
    public class GetSimulationCmdlet : SimulationCmdlets
    {
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync(Prefix).Result;
            ThrowOnHttpFailure(resp, "GetSimulationStatus", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var status = JsonConvert.DeserializeObject<SimulationStatus>(msg);

            WriteObject(status);
        }
    }

    [Cmdlet(VerbsCommon.Get, Names.Sessions)]
    [OutputType(typeof(List<SessionEntry>))]
    public class SessionsCmdlet : SimulationCmdlets
    {
        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}/sessions";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetSimulationSessionList", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonConvert.DeserializeObject<SessionList>(msg);

            WriteObject(details.Sessions, true);
        }
    }

    [Cmdlet(VerbsCommon.Get, Names.Session)]
    [OutputType(typeof(ClusterSession))]
    public class ClusterSessionCmdlet : SimulationCmdlets
    {
        [Parameter(Mandatory = true)] public long ActiveSessionId { get; set; }

        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}/sessions/{ActiveSessionId}";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetSimulationSessionDetails", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonConvert.DeserializeObject<ClusterSession>(msg);

            WriteObject(details);
        }
    }
}
