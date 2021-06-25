// This module contains support functions for converting an item from a parsed
// json string into a final field value.
//
// Note that these functions both ensure the correct type for the result, as
// well as filling in any missing data with its default value.

// asItem is the workhorse function.  It takes an input value and either
// returns the default value, or invokes the supplied conversion function.
export function asItem<T>(f: (item: any) => T, item: any, def: T): T {
    if (item !== undefined && item !== null) {
        return f(item)
    }

    return def
}

// asBool specializes asItem for the boolean type.
export function asBool(item: any): boolean {
    return asItem<boolean>(Boolean, item, false)
}

// asNumber specializes asItem for the number type.
export function asNumber(item: any): number {
    return asItem<number>(Number, item, 0)
}

// asString specializes asItem for the string type.
export function asString(item: any): string {
    return asItem<string>(String, item, "")
}

// asArray processes an entire array
export function asArray<T>(f: (item: any) => T, item: any): T[] {
	const res: T[] = []

    if (item !== undefined && item !== null) {
    	for (const e of item) {
    		res.push(f(e))
    	}
    }

	return res
}

// asMap creates and fills in a Map that has the specified key and value types.
// It uses the supplied function for any type conversions from the input value
// to the key and map entry value.
export function asMap<K, V>(entries: any, action: (key: any, value: any) => [K, V]): Map<K, V> {
  const res = new Map<K, V>()

  if (entries !== undefined && entries != null) {
    Object.entries(entries).forEach(([key, value]) => {
        const [k, v] = action(key, value)
        res.set(k, v)
    })
  }

  return res
}

// +++ well known type handling

export interface Duration {
	seconds: number
	nanos: number
}

// Get the nanosecond component from the duration string
function parseNano(val: string) : number {
    let nanoIndex = val.indexOf("n")
    if (nanoIndex > -1) {
        return +val.substr(0, nanoIndex)
    }

    return 0
}

export function durationFromJson(duration: string | undefined) : Duration {
	let val : Duration = {seconds: 0, nanos: 0}

   	if (duration !== undefined && duration !== null) {
		let indexS = duration.indexOf("s")
		if (indexS > -1) {
		    const segment1 = duration.substr(0, indexS)
		    val.seconds = +segment1

		    val.nanos = parseNano(duration.substr(indexS + 1))
		} else {
		    val.nanos = parseNano(duration)
		}
   	}

    return val
}
