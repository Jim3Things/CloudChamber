import { asNumber } from "../utils"

/* eslint-disable */
export const protobufPackage = "common";

/**
 * Define the structure for a simulated time
 *
 * Simulated time is defined as an incrementing tick count, that begins at zero.
 */
export interface Timestamp {
  ticks: number;
}

export const Timestamp = {
  fromJSON(object: any): Timestamp {
    return {
      ticks: asNumber(object.ticks),
    }
  },
};
