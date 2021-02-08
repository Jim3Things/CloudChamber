using System.Management.Automation;
using System.Net.Http;
using CloudChamber.Protos.Admin;
using Google.Protobuf;
using Google.Protobuf.Collections;

namespace CloudChamber.Protos.Admin
{
    /// <summary>
    ///     Extend UserPublic with the identifying name, and the associated
    ///     revision tag.
    /// </summary>
    public partial class UserPublic
    {
        /// <summary>
        ///     Name contains the string used to identify the user when logging in.
        /// </summary>
        public string Name { get; set; }

        /// <summary>
        ///     ETag contains the revision that was current when the data was read.
        ///     It should be supplied on any attempt to modify the user data, in
        ///     order to avoid updates based on stale data.
        /// </summary>
        public string ETag { get; set; }
    }
}

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     NamedUserCmdlets further specializes UserCmdlets by including the most
    ///     common parameters.
    /// </summary>
    public class NamedUserCmdlets : LoggedInCmdlet
    {
        public NamedUserCmdlets() : base("/api/users") { }

        /// <summary>
        ///     Name is the name of the user to operate on.
        /// </summary>
        [Parameter(Position = 0, Mandatory = true)]
        public string Name { get; set; }
    }

    /// <summary>
    ///     GetUsersCmdlet gets the list of users known to the cluster.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.Users)]
    [OutputType(typeof(RepeatedField<UserList.Types.Entry>))]
    public class GetUsersCmdlet : LoggedInCmdlet
    {
        public GetUsersCmdlet() : base("/api/users/") { }

        /// <summary>
        ///     ProcessRecord retrieves the user list and writes it to the output
        ///     stream.
        /// </summary>
        protected override void ProcessRecord()
        {
            var resp = Session.Client.GetAsync(Prefix).Result;

            ThrowOnHttpFailure(resp, "GetUsersList", null);

            var msg = resp.Content.ReadAsStringAsync().Result;
            var list = JsonParser.Default.Parse<UserList>(msg);

            WriteObject(list.Users, true);
        }
    }

    /// <summary>
    ///     GetUserCmdlet retrieves the attributes associated with the specified user.
    /// </summary>
    [Cmdlet(VerbsCommon.Get, Names.User)]
    [OutputType(typeof(UserPublic))]
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
            var details = JsonParser.Default.Parse<UserPublic>(msg);

            details.Name = Name;
            details.ETag = resp.Headers.GetHeader("ETag", "-1");

            WriteObject(details);
        }
    }

    /// <summary>
    ///     NewUserCmdlet creates a new user with the specified name and supplied
    ///     attributes.
    /// </summary>
    [Cmdlet(VerbsCommon.New, Names.User)]
    [OutputType(typeof(UserPublic))]
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
            var json = JsonFormatter.Default.Format(new UserDefinition
            {
                Enabled = Enabled,
                CanManageAccounts = Admin,
                Password = Password
            });

            var resp = Session.Client.PostAsync(
                $"{Prefix}/{Name}",
                new StringContent(json)).Result;

            ThrowOnHttpFailure(resp, "NewUserError", Name);

            WriteObject(new UserPublic
            {
                Name = Name,
                Enabled = Enabled,
                CanManageAccounts = Admin,
                NeverDelete = false,
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
