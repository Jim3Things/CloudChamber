using System.Management.Automation;
using CloudChamber.Cmdlets.Protos;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     Common trace-related cmdlet properties.
    /// </summary>
    public class TraceCmdlets : LoggedInCmdlet
    {
        protected TraceCmdlets() : base("/api/logs") { }
    }

    /// <summary>
    ///     Get the current tracing policy.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.TracePolicy)]
    [OutputType(typeof(TracePolicy))]
    public class GetTracePolicyCmdlet : TraceCmdlets
    {
        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}/policy";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetTracePolicyError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var policy = JsonConvert.DeserializeObject<TracePolicy>(msg);

            WriteObject(policy);
        }
    }

    /// <summary>
    ///     Get an extract of the traces.  Waits for new traces if it the From
    ///     parameter is set to later than the last known trace.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Traces)]
    [OutputType(typeof(Traces))]
    public class GetTracesCmdlet : TraceCmdlets
    {
        /// <summary>
        ///     From is the starting trace ID.
        /// </summary>
        [Parameter] public long From { get; set; } = 0;

        /// <summary>
        ///     For is the maximum number of traces to return in this call
        /// </summary>
        [Parameter] public long For { get; set; } = 100;

        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}?from={From}&for={For}";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetTracePolicyError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var traces = JsonConvert.DeserializeObject<Traces>(msg);

            WriteObject(traces);
        }
    }
}
