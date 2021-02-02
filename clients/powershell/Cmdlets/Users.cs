using System.Management.Automation;
using System.Net.Http;
using CloudChamber.Cmdlets.Protos;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets
{

    /// <summary>
    ///     NamedUserCmdlets further specializes UserCmdlets by including the most
    ///     common parameters.
    /// </summary>
    public class NamedUserCmdlets : CmdletBase
    {
        public NamedUserCmdlets() : base("/api/users") { }

        /// <summary>
        ///     Name is the name of the user to operate on.
        /// </summary>
        [Parameter(Position = 0, Mandatory = true)]
        public string Name { get; set; }

        /// <summary>
        ///     Session is the logged-in session to use for the operation.
        /// </summary>
        [Parameter(Mandatory = true)]
        public Session Session { get; set; }
    }

    /// <summary>
    ///     GetUsersCmdlet gets the list of users known to the cluster.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Users)]
    public class GetUsersCmdlet : CmdletBase
    {
        public GetUsersCmdlet()  : base("/api/users/") { }

        /// <summary>
        ///     Session contains the currently logged-in http client to use when
        ///     contacting the cluster.
        /// </summary>
        [Parameter(Mandatory = true)]
        public Session Session { get; set; }

        /// <summary>
        ///     ProcessRecord retrieves the user list and writes it to the output
        ///     stream.
        /// </summary>
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync(Prefix).Result;

            ThrowOnHttpFailure(resp, "GetUsersList", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var list = JsonConvert.DeserializeObject<UserList>(msg);

            WriteObject(list.Users, true);
        }
    }

    /// <summary>
    ///     GetUserCmdlet retrieves the attributes associated with the specified user.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.User)]
    public class GetUserCmdlet : NamedUserCmdlets
    {
        /// <summary>
        ///     ProcessRecord retrieves the user details and writes them to the
        ///     output stream.
        /// </summary>
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync($"{Prefix}/{Name}").Result;

            ThrowOnHttpFailure(resp, "GetUserDetails", Name);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var details = JsonConvert.DeserializeObject<PublicUserDetails>(msg);

            WriteObject(new UserDetails
            {
                Name = Name,
                Enabled = details.Enabled,
                ManageAccounts = details.ManageAccounts,
                Protected = details.Protected,
                ETag = resp.Headers.GetHeader("ETag", "-1")
            });
        }
    }

    /// <summary>
    ///     NewUserCmdlet creates a new user with the specified name and supplied
    ///     attributes.
    /// </summary>
    [Cmdlet(VerbsCommon.New, Names.User)]
    public class NewUserCmdlet : NamedUserCmdlets
    {
        /// <summary>
        ///     Password contains the initial password for this user.
        /// </summary>
        [Parameter(Position = 1, Mandatory = true)]
        public string Password { get; set; }

        /// <summary>
        ///     Admin is true if this user can manager other users' accounts.
        /// </summary>
        [Parameter]
        public SwitchParameter Admin { get; set; }

        /// <summary>
        ///     Enabled is true if this user is allowed to login and use the
        ///     cluster.
        /// </summary>
        [Parameter]
        public SwitchParameter Enabled { get; set; }

        /// <summary>
        ///     ProcessRecord creates the new user and writes the details for it
        ///     into the output stream.
        /// </summary>
        protected override void ProcessRecord()
        {
            var json = JsonConvert.SerializeObject(new NewUserDetails
            {
                Enabled = Enabled,
                ManageAccounts = Admin,
                Password = Password
            });

            var resp = Session.Client.PostAsync(
                $"{Prefix}/{Name}",
                new StringContent(json)).Result;

            ThrowOnHttpFailure(resp, "NewUserError", Name);

            WriteObject(new UserDetails
            {
                Name = Name,
                Enabled = Enabled,
                ManageAccounts = Admin,
                Protected = false,
                ETag = resp.Headers.GetHeader("ETag", "-1")
            });
        }
    }

    /// <summary>
    ///     RemoveUserCmdlet removes an existing use from the target cluster.
    /// </summary>
    [Cmdlet(VerbsCommon.Remove, Names.User)]
    public class RemoveUserCmdlet : NamedUserCmdlets
    {
        /// <summary>
        ///     ProcessRecord performs the deletion.
        /// </summary>
        protected override void ProcessRecord()
        {
            var resp = Session.Client.DeleteAsync($"{Prefix}/{Name}").Result;

            ThrowOnHttpFailure(resp, "DeleteUserError", Name);

            WriteObject($"User {Name} deleted.");
        }
    }
}
