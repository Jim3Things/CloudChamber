using System;
using System.Management.Automation;
using System.Net.Http;
using CloudChamber.Protos.Common;
using CloudChamber.Protos.Services;
using Google.Protobuf;

namespace CloudChamber.Protos.Services
{
    /// <summary>
    ///     Extend to include the revision tag, and a usable form of the
    ///     measured policy delay.
    /// </summary>
    public partial class StatusResponse
    {
        /// <summary>
        ///     ETag contains the revision that was current when the data was read.
        ///     It should be supplied on any attempt to modify the user data, in
        ///     order to avoid updates based on stale data.
        /// </summary>
        public string ETag { get; set; }

        /// <summary>
        ///     Return the measured policy real time delay between incrementing
        ///     the simulated time in a format that is normally usable by .Net
        ///     applications.
        /// </summary>
        public TimeSpan Delay => MeasuredDelay.ToTimeSpan();
    }
}

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     StepperCmdlets is the base class providing common prefix and command
    ///     parameter values.
    /// </summary>
    public class StepperCmdlets : LoggedInCmdlet
    {
        protected StepperCmdlets() : base("/api/stepper") { }
    }

    /// <summary>
    ///     ETagStepperCmdlets extends the StepperCmdlets base class with the
    ///     common ETag check and If-Match handling.
    /// </summary>
    public class ETagStepperCmdlets : StepperCmdlets
    {
        /// <summary>
        ///     This is the etag value that denotes a forced match.
        /// </summary>
        protected const string ForceMatchETag = "\"-1\"";

        /// <summary>
        ///     The stepper policy version that should be active in order to
        ///     apply this change.
        /// </summary>
        [Parameter(Mandatory = true, ParameterSetName = "VersionCheck")]
        public string Revision { get; set; }

        /// <summary>
        ///     Force the change through regardless of current policy version.
        /// </summary>
        [Parameter(Mandatory = true, ParameterSetName = "Force")]
        public SwitchParameter Force { get; set; }

        /// <summary>
        ///     Determine the match tag to provide
        /// </summary>
        /// <returns>The ETag match string to send to the cluster</returns>
        /// <exception cref="InvalidOperationException">Force was set to false.</exception>
        protected string CalculateETag()
        {
            switch (ParameterSetName)
            {
                case "Force":
                    if (!Force) throw new InvalidOperationException("-Force must be true");

                    return ForceMatchETag;

                case "VersionCheck":
                    return Revision;
            }

            throw new InvalidOperationException(
                $"Invalid mix of arguments: {ParameterSetName}, {Force}, {Revision}");
        }
    }

    /// <summary>
    ///     NowCmdlet returns the current simulated time.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Time)]
    [OutputType(typeof(long))]
    public class NowCmdlet : StepperCmdlets
    {
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync($"{Prefix}/now").Result;
            ThrowOnHttpFailure(resp, "GetNowError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var res = JsonParser.Default.Parse<Timestamp>(msg);

            WriteObject(res.Ticks);
        }
    }

    /// <summary>
    ///     StepperStatusCmdlet returns the currently active StepperPolicy values.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.TimePolicy)]
    [OutputType(typeof(StepperPolicy))]
    public class StepperStatusCmdlet : StepperCmdlets
    {
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync($"{Prefix}").Result;
            ThrowOnHttpFailure(resp, "GetStepperPolicyError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;

            var policy = JsonParser.Default.Parse<StatusResponse>(msg);
            policy.ETag = resp.Headers.GetHeader("ETag", "\"-1\"");

            WriteObject(policy);
        }
    }

    /// <summary>
    ///     AdvanceTimeCmdlet moves the simulated time forward by the specified
    ///     number of simulated time ticks.
    /// </summary>
    [Cmdlet(VerbsCommon.Step, Names.Time)]
    [OutputType(typeof(long))]
    public class AdvanceTimeCmdlet : StepperCmdlets
    {
        /// <summary>
        ///     The number of ticks to advance simulated time by.
        /// </summary>
        [Parameter]
        public long Ticks { get; set; } = 1;

        protected override void ProcessRecord()
        {
            var resp = Session.Client.PutAsync(
                $"{Prefix}?advance={Ticks}",
                new StringContent(string.Empty)).Result;

            ThrowOnHttpFailure(resp, "GetNowError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var now = JsonParser.Default.Parse<Timestamp>(msg);

            WriteObject(now.Ticks);
        }
    }

    /// <summary>
    ///     ResumeTimeCmdlet sets the active policy to Measured.
    /// </summary>
    [Cmdlet(VerbsLifecycle.Resume, Names.Time)]
    [OutputType(typeof(string))]
    public class ResumeTimeCmdlet : ETagStepperCmdlets
    {
        /// <summary>
        ///     The number of ticks to advance simulated time every second.
        /// </summary>
        [Parameter(Mandatory = true)]
        public long Rate { get; set; } = 1;

        protected override void ProcessRecord()
        {
            var tag = CalculateETag();

            var request = new HttpRequestMessage
            {
                RequestUri = new Uri($"{Session.Client.BaseAddress}{Prefix}?mode=automatic:{Rate}"),
                Method = HttpMethod.Put,
            };

            request.Headers.Add("If-Match", tag);

            var resp = Session.Client.SendAsync(request).Result;

            ThrowOnHttpFailure(resp, "ResumeTimeError", null);

            WriteObject(resp.Headers.GetHeader("ETag", "-1"));
        }
    }

    /// <summary>
    ///     SuspendCmdlet set the active policy to Manual.
    /// </summary>
    [Cmdlet(VerbsLifecycle.Suspend, Names.Time)]
    [OutputType(typeof(string))]
    public class SuspendTimeCmdlet : ETagStepperCmdlets
    {
        protected override void ProcessRecord()
        {
            var tag = CalculateETag();

            var request = new HttpRequestMessage
            {
                RequestUri = new Uri($"{Session.Client.BaseAddress}{Prefix}?mode=manual"),
                Method = HttpMethod.Put,
            };

            request.Headers.Add("If-Match", tag);

            var resp = Session.Client.SendAsync(request).Result;

            ThrowOnHttpFailure(resp, "SuspendTimeError", null);

            WriteObject(resp.Headers.GetHeader("ETag", "-1"));
        }
    }

    /// <summary>
    ///     WaitTimeCmdlet waits until the cluster's simulated time advances to the
    ///     supplied deadline.
    /// </summary>
    [Cmdlet(VerbsLifecycle.Wait, Names.Time)]
    [OutputType(typeof(StatusResponse))]
    public class WaitTimeCmdlet : StepperCmdlets
    {
        /// <summary>
        ///     The waiting deadline in simulated time ticks.
        /// </summary>
        [Parameter(Mandatory = true)]
        public long Until { get; set; }

        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync($"{Prefix}/now?after={Until}").Result;

            ThrowOnHttpFailure(resp, "WaitUntilTimeError", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var res = JsonParser.Default.Parse<StatusResponse>(msg);

            WriteObject(res);
        }
    }
}
