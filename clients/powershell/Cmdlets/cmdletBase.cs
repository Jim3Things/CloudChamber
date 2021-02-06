using System.Management.Automation;
using System.Net.Http;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     Common attributes and support methods for all CloudChamber cmdlets.
    /// </summary>
    public class CmdletBase : PSCmdlet
    {
        protected CmdletBase(string prefix)
        {
            Prefix = prefix;
        }

        protected string Prefix { get; }

        /// <summary>
        ///     ThrowOnHttpFailure throws a terminating error if the http request did not
        ///     succeed.  The error is structured to include information from the http
        ///     response message.
        /// </summary>
        /// <param name="resp">HTTP response to evaluate.</param>
        /// <param name="errorName">Descriptive error ID string.</param>
        /// <param name="target">optional target object associated with the error.</param>
        protected void ThrowOnHttpFailure(HttpResponseMessage resp, string errorName, object target)
        {
            if (!resp.IsSuccessStatusCode)
                ThrowTerminatingError(resp.ToErrorRecord(errorName, target));
        }
    }

    /// <summary>
    ///     Extend the CmdletBase with the common Session parameter to handle the
    ///     normal logged-in operations.
    /// </summary>
    public class LoggedInCmdlet : CmdletBase
    {
        protected LoggedInCmdlet(string prefix) : base(prefix) { }

        /// <summary>
        ///     Session is the logged-in session to use for the operation.
        /// </summary>
        [Parameter(Mandatory = true)]
        public Session Session { get; set; }
    }
}
