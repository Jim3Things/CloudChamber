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