using System;
using System.Management.Automation;
using System.Net.Http;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     Session holds the state for an active connection to a cluster.
    /// </summary>
    public class Session : IDisposable
    {
        /// <summary>
        ///     This constructor initializes the session and establishes the client
        ///     context for the target cluster.
        /// </summary>
        /// <param name="uri">address of the target cluster.</param>
        /// <param name="name">logged in name.</param>
        public Session(Uri uri, string name)
        {
            Name = name;
            Client = new HttpClient
            {
                BaseAddress = uri
            };
        }

        /// <summary>
        ///     Client is the active HTTP client to use for any REST calls.
        /// </summary>
        public HttpClient Client { get; private set; }

        /// <summary>
        ///     Name is the username for the currently logged in user.
        /// </summary>
        public string Name { get; }

        /// <summary>
        ///     Performs the tasks associated with freeing, releasing, or resetting
        ///     the http client resources.
        /// </summary>
        public void Dispose()
        {
            GC.SuppressFinalize(this);

            if (Client != null)
            {
                Client.Dispose();
                Client = null;
            }
        }
    }

    /// <summary>
    ///     Connect to a specified account on the target cluster.
    /// </summary>
    [Cmdlet(VerbsCommunications.Connect, Names.Account)]
    [OutputType(typeof(Session))]
    public class LoginCmdlet : PSCmdlet
    {
        /// <summary>
        ///     ClusterUri is the address of the target cluster.
        /// </summary>
        [Parameter(Position = 0, Mandatory = true)]
        public Uri ClusterUri { get; set; }

        /// <summary>
        ///     Name is the username of the account to log in.
        /// </summary>
        [Parameter(Position = 1, Mandatory = true)]
        public string Name { get; set; }

        /// <summary>
        ///     Password is the password string to use when logging in.
        /// </summary>
        [Parameter(Position = 2, Mandatory = true)]
        public string Password { get; set; }

        protected override void ProcessRecord()
        {
            var session = new Session(ClusterUri, Name);
            var path = $"/api/users/{Name}?op=login";

            var resp = session.Client.PutAsync(path, new StringContent(Password)).Result;

            resp.EnsureSuccessStatusCode();

            var msg = resp.Content.ReadAsStringAsync().Result;

            WriteObject(session);
        }
    }

    /// <summary>
    ///     Disconnect an active session.
    /// </summary>
    [Cmdlet(VerbsCommunications.Disconnect, Names.Account)]
    [OutputType(typeof(Session))]
    public class LogoutCmdlet : PSCmdlet
    {
        /// <summary>
        ///     Session is the active session to disconnect.
        /// </summary>
        [Parameter(Mandatory = true)]
        public Session Session { get; set; }

        protected override void ProcessRecord()
        {
            var path = $"/api/users/{Session.Name}?op=logout";

            var resp = Session.Client.PutAsync(path, new StringContent(string.Empty)).Result;

            resp.EnsureSuccessStatusCode();

            var msg = resp.Content.ReadAsStringAsync().Result;

            WriteObject(Session);
        }
    }
}
