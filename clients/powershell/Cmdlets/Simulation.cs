using System.Collections.Generic;
using System.Management.Automation;
using CloudChamber.Cmdlets.Protos;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     SimulationCmdlets provides the common specialization used by all
    ///     simulation object cmdlets.
    /// </summary>
    public class SimulationCmdlets : LoggedInCmdlet
    {
        protected SimulationCmdlets() : base("/api/simulation") { }
    }

    /// <summary>
    ///     Get the current status of the simulation.
    /// </summary>
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

    /// <summary>
    ///     Get the summary list of active logged-in sessions.
    /// </summary>
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

    /// <summary>
    ///     Get the details for a given logged-in session.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Session)]
    [OutputType(typeof(ClusterSession))]
    public class ClusterSessionCmdlet : SimulationCmdlets
    {
        /// <summary>
        ///     Identifier for the session, such as supplied in the summary
        ///     list.
        /// </summary>
        [Parameter(Mandatory = true)] public long Id { get; set; }

        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}/sessions/{Id}";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetSimulationSessionDetails", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonConvert.DeserializeObject<ClusterSession>(msg);

            WriteObject(details);
        }
    }
}
