using System;
using System.Collections.Generic;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets.Protos
{
    #region data structures

    /// <summary>
    ///     UserList contains the list of users known on the cluster, along with
    ///     their URIs, and summary information needed to help with a list display.
    /// </summary>
    public class UserList
    {
        [JsonProperty("users")] public List<UserListEntry> Users { get; set; }
    }

    /// <summary>
    ///     Each UserListEntry provides a summary for a known user on the cluster.
    ///     That summary information includes the name, the access URI, and whether
    ///     or not the user is protected from deletion.
    /// </summary>
    public class UserListEntry
    {
        /// <summary>
        ///     Name contains the string used to identify the user when logging in.
        /// </summary>
        [JsonProperty("name")]
        public string Name { get; set; }

        /// <summary>
        ///     Uri contains the address to use to get or modify attributes for
        ///     user.
        /// </summary>
        [JsonProperty("uri")]
        public Uri Uri { get; set; }

        /// <summary>
        ///     Protected is true if this user cannot be deleted.  It is included
        ///     in the summary information in order to aid any visual list
        ///     construction - i.e. to avoid adding icons to delete the user if it
        ///     cannot be deleted.
        /// </summary>
        [JsonProperty("protected")]
        public bool Protected { get; set; }
    }

    /// <summary>
    ///     PublicUserDetails holds the attributes for a user, except for the
    ///     password.
    /// </summary>
    public class PublicUserDetails
    {
        /// <summary>
        ///     ManageAccounts is true if this user can perform user management
        ///     operations.
        /// </summary>
        [JsonProperty("canManageAccounts")]
        public bool ManageAccounts { get; set; }

        /// <summary>
        ///     Enabled is true if this user can log into a cluster for any reason.
        /// </summary>
        [JsonProperty("enabled")]
        public bool Enabled { get; set; }

        /// <summary>
        ///     Protected is true if this user cannot be removed.
        /// </summary>
        [JsonProperty("neverDelete")]
        public bool Protected { get; set; }
    }

    /// <summary>
    ///     NewUserDetails contains the attributes to pass on a user creation
    ///     message.
    /// </summary>
    public class NewUserDetails
    {
        /// <summary>
        ///     ManageAccounts is true if this user can perform user management
        ///     operations.
        /// </summary>
        [JsonProperty("canManageAccounts")]
        public bool ManageAccounts { get; set; }

        /// <summary>
        ///     Enabled is true if this user can log into a cluster for any reason.
        /// </summary>
        [JsonProperty("enabled")]
        public bool Enabled { get; set; }

        /// <summary>
        ///     Password contains the initial password for this user.
        /// </summary>
        [JsonProperty("password")]
        public string Password { get; set; }
    }

    /// <summary>
    ///     UserDetails annotates a user's public details with their name and the
    ///     revision number that was current when the details were read. This is
    ///     expected to be used by Powershell scripts when performing operations
    ///     on a given user.
    /// </summary>
    public class UserDetails : PublicUserDetails
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

    #endregion data structures
}
