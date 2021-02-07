using System;
using Newtonsoft.Json;
using Newtonsoft.Json.Converters;

namespace CloudChamber.Cmdlets.Protos
{
#region data structures

    /// <summary>
    ///     Timestamp contains the simulated time at the target cluster.
    /// </summary>
    public class Timestamp
    {
	    /// <summary>
	    ///     Ticks is the simulated time in ticks.
	    /// </summary>
	    [JsonProperty("ticks")]
        public long Ticks { get; set; }
    }

    /// <summary>
    ///     StepperPolicy contains the attributes returned from the cluster for the
    ///     current simulated time policy.
    /// </summary>
    public class StepperPolicy
    {
	    /// <summary>
	    ///     The set of recognized simulated time policies.
	    /// </summary>
	    public enum PolicyEnum
        {
	        /// <summary>
	        ///     Invalid denotes that no valid policy was provided.
	        /// </summary>
	        Invalid,

	        /// <summary>
	        ///     NoWait denotes that the simulation immediately steps time forward
	        ///     to the next waiter.
	        /// </summary>
	        NoWait,

	        /// <summary>
	        ///     Manual denotes that simulated time only moves forward due to an
	        ///     external advance operation.
	        /// </summary>
	        Manual,

	        /// <summary>
	        ///     Measured denotes that simulated time moves forward at a known
	        ///     rate in real time.
	        /// </summary>
	        Measured
        }

	    /// <summary>
	    ///     Policy identifies the current simulated time policy in effect.
	    /// </summary>
	    [JsonProperty("policy")]
        [JsonConverter(typeof(StringEnumConverter))]
        public PolicyEnum Policy { get; set; }

	    /// <summary>
	    ///     Delay indicates the delay between simulated time increments.  This
	    ///     is zero unless the policy is Measured.
	    /// </summary>
	    [JsonProperty("measuredDelay")]
        [JsonConverter(typeof(DurationConverter))]
        public TimeSpan Delay { get; set; }

	    /// <summary>
	    ///     ETag contains the ETag header value returned with the Http response
	    ///     containing this policy instance.
	    /// </summary>
	    [JsonIgnore]
        public string ETag { get; set; }
    }

    #endregion data structures
}
