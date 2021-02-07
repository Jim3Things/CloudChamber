using System.Linq;
using System.Management.Automation;
using System.Net.Http;
using System.Net.Http.Headers;

namespace CloudChamber.Cmdlets
{
    /// <summary>
    ///     This class contains extension methods for the HttpResponseMessage.
    /// </summary>
    public static class ResponseHelpers
    {
        /// <summary>
        ///     ToErrorRecord converts an http failure status into an ErrorRecord.
        ///     The resulting record includes the error text, if any, from the
        ///     responding service.
        /// </summary>
        /// <param name="resp">HTTP response message under evaluation</param>
        /// <param name="errorName">Category name for the error.</param>
        /// <param name="target">optional reference to target of this request</param>
        /// <returns></returns>
        public static ErrorRecord ToErrorRecord(this HttpResponseMessage resp, string errorName,
            object target)
        {
            var msg = resp.Content.ReadAsStringAsync().Result;

            return new ErrorRecord(
                new HttpRequestException(msg),
                errorName,
                ErrorCategory.InvalidOperation,
                target);
        }
    }

    /// <summary>
    ///     This class contains extension methods for the HttpResponseHeaders
    /// </summary>
    public static class HeaderHelpers
    {
        /// <summary>
        ///     GetHeader retrieves the first value for the specified header, or a
        ///     caller-specified default.
        /// </summary>
        /// <param name="headers">The collection of http headers to search.</param>
        /// <param name="name">The name of the header field requested.</param>
        /// <param name="missing">The value to return if the specified header is not found.</param>
        /// <returns></returns>
        public static string GetHeader(this HttpResponseHeaders headers, string name,
            string missing)
        {
            return headers.TryGetValues(name, out var values)
                ? values.First()
                : missing;
        }
    }
}
