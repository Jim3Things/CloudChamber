using System;
using Newtonsoft.Json;

namespace CloudChamber.Cmdlets.Protos
{
    /// <summary>
    ///     DurationConverter converts between the JSON representation of a protobuf
    ///     Duration type and a C# TimeSpan.
    /// </summary>
    public class DurationConverter : JsonConverter<TimeSpan>
    {
        public override bool CanRead => true;
        public override bool CanWrite => true;

        /// <summary>
        ///     Format a TimeSpan value into the duration JSON string
        /// </summary>
        /// <param name="writer">output JSON writer</param>
        /// <param name="value">TimeSpan to record</param>
        /// <param name="serializer">unused</param>
        public override void WriteJson(JsonWriter writer, TimeSpan value, JsonSerializer serializer)
        {
            var sec = value.TotalMilliseconds / 1000;
            var s = $"{sec}s";
            writer.WriteValue(s);
        }

        /// <summary>
        ///     Convert a JSON string of the form {ss}.{ff}s into a TimeSpan.
        /// </summary>
        /// <param name="reader">JSON data reader</param>
        /// <param name="objectType">unused</param>
        /// <param name="existingValue">unused</param>
        /// <param name="hasExistingValue">unused</param>
        /// <param name="serializer">unused</param>
        /// <returns>Corresponding TimeSpan value</returns>
        public override TimeSpan ReadJson(
            JsonReader reader,
            Type objectType,
            TimeSpan existingValue,
            bool hasExistingValue,
            JsonSerializer serializer)
        {
            var s = (string) reader.Value;
            if (string.IsNullOrEmpty(s)) return TimeSpan.Zero;

            if (!float.TryParse(s.Trim().TrimEnd('s'), out var f)) return TimeSpan.Zero;

            var ms = (int) (f * 1_000);
            return new TimeSpan(0, 0, 0, 0, ms);
        }
    }
}
