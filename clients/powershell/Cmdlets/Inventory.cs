using System.Management.Automation;
using CloudChamber.Protos.Inventory;
using Google.Protobuf;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     Get the summary list of racks, and the memoized information about
    ///     the maximum sizes used by the racks.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Racks)]
    [OutputType(typeof(External.Types.ZoneSummary))]
    public class GetRacksCmdlet : LoggedInCmdlet
    {
        public GetRacksCmdlet() : base("/api/racks") { }

        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync(Prefix).Result;
            ThrowOnHttpFailure(resp, "GetRacksList", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var summary = JsonParser.Default.Parse<External.Types.ZoneSummary>(msg);

            WriteObject(summary);
        }
    }

    /// <summary>
    ///     Get the detail information for the specified rack.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Rack)]
    [OutputType(typeof(External.Types.Rack))]
    public class GetRackCmdlet : LoggedInCmdlet
    {
        public GetRackCmdlet() : base("/api/racks/") { }

        /// <summary>
        ///     Name of the rack to retrieve.
        /// </summary>
        [Parameter(Mandatory = true)]
        public string Name { get; set; }

        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}{Name}";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetRackDetails", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonParser.Default.Parse<External.Types.Rack>(msg);

            WriteObject(details);
        }
    }

    /// <summary>
    ///     Get the detail information for the specified blade.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Blade)]
    [OutputType(typeof(BladeCapacity))]
    public class GetBladeCmdlet : LoggedInCmdlet
    {
        public GetBladeCmdlet() : base("/api/racks/") { }

        /// <summary>
        ///     Name of the rack to retrieve.
        /// </summary>
        [Parameter(Mandatory = true)]
        public string Name { get; set; }

        /// <summary>
        ///     Identify the blade within the specified rack to retrieve.
        /// </summary>
        [Parameter(Mandatory = true)]
        public long Id { get; set; }

        protected override void ProcessRecord()
        {
            var uri = $"{Prefix}{Name}/Blades/{Id}";
            var resp = Session.Client.GetAsync(uri).Result;
            ThrowOnHttpFailure(resp, "GetBladeDetails", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonParser.Default.Parse<BladeCapacity>(msg);

            WriteObject(details);
        }
    }
}
