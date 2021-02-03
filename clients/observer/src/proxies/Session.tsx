// Manage the REST session attributes

interface JsonUserDetails {
    enabled: boolean
    accountManager: boolean
    neverDelete: boolean
}

export interface SessionUser {
    name: string
    enabled: boolean
    accountManager: boolean
    neverDelete: boolean
}

// CCError is a specialization of the Error interface that carries the
// original response object along with it.  This allows the catch handler to
// asynchronously retrieve the extended message details that are in the body.
export class CCError extends Error {
    public resp: Response

    constructor(response: Response, msg: string) {
        super(msg);
        this.resp = response
    }

    public toString(): string {
        return super.toString()
    }
}

export class Session {
    // login, and get the user details for the logged in user.  Attach as
    // session details here.
    public logon(username: string, password: string): Promise<SessionUser> {
        const path = "/api/users/" + username + "?op=login"
        const request = new Request(path, {method: "PUT", body: password})

        return fetch(request)
            .then((resp) => {
                failIfError(request, resp)

                const detailsPath = "/api/users/" + username
                const requestDetails = new Request(detailsPath, {method: "GET"})

                return getJson<JsonUserDetails>(requestDetails)
            })
            .then((details) => {
                return {...details, name: username }
            })
    }

    // Log out of the current session
    public logout(username: string) : Promise<string> {
        const path = "/api/users/" + username + "?op=logout"
        const request = new Request(path, {method: "PUT"})

        return fetch(request)
            .then((resp) => {
                if (!resp.ok) {
                    // Something went wrong.  So we need to force that the
                    // session is gone and continue as if the logout was
                    // successful.
                    deleteCookie("CC-Session")
                }

                return "logged out"
            })
    }
}

// Throw a consistent error if the response indicates a failure to process
export function failIfError(request: Request, resp: Response) {
    if (!resp.ok) {
        throw new CCError(
            resp,
            "Error in response, path='" + request.url + "' status: (" + resp.status + ") " + resp.statusText)
    }
}

// Return the best error details: either the extended CloudChamber text in the
// body, or the normal string, if this is a generic error.
export function getErrorDetails(msg: any, save: (val: string) => void): void {
    if (msg.hasOwnProperty("resp")) {
        const err: CCError = {...msg}
        err.resp.text().then((details) => {
            save(details )
        })
    }

    save (msg.toString())
}

// Convert the response body into a JSON-parsed type
export function getJson<T>(request: Request, signal?: AbortSignal | undefined): Promise<T> {
    return fetch(request, { signal: signal })
        .then((resp: Response) => {
            failIfError(request, resp)
            return resp.json() as Promise<T>
        })
}

// +++ ETag support functions

// Get the value of an ETag, as a number
export function getETag(resp: Response): number {
    const tag = resp.headers.get("ETag")
    if (tag === null) {
        return -1
    }

    return parseInt(tag, 10)
}

// Set the ETag into a header as a match condition
export function ETagHeader(tag: number) : HeadersInit {
    let requestHeaders: HeadersInit = new Headers()
    requestHeaders.set('If-Match', tag.toString(10))

    return requestHeaders
}

// --- ETag support functions

// Delete the named cookie by setting it to no value, and to have already
// expired
export function deleteCookie(name: string) {
    const oneHourInMs = 60 * 60 * 1000

    document.cookie = name +
        "=; expires=" + (new Date(Date.now() - oneHourInMs)).toUTCString() +
        "; path=/";
}