// This module contains the proxy handler for calling the REST User management
// service in the Cloud Chamber backend.

// Define the user details as supplied by the REST service.

// TODO: Investigate having these produced directly from the originating
//       protobuf file.

import {ETagHeader, failIfError, getETag} from "./Session";
import {Rights, UserList, UserUpdate, UserPassword} from "../../../../pkg/protos/admin/users";

export interface PublicUserDetails {
    enabled: boolean
    rights: Rights
    neverDelete: boolean
    eTag: number
}

export class UserDetails implements PublicUserDetails{
    password: string;
    rights: Rights
    enabled: boolean;
    neverDelete: boolean
    eTag: number

    constructor() {
        this.password = "";
        this.rights =
        {
            canInjectFaults: false,
            canManageAccounts: false,
            canModifyInventory: false,
            canModifyWorkloads: false,
            canPerformRepairs: false,
            canStepTime: false
        }
        this.enabled = false;
        this.neverDelete = false;
        this.eTag = -1
    }
}

const nullDetails = new UserDetails()

// Utility class that provides a proxy to the Cloud Chamber User management
// service.

// TODO: This proxy current fakes out the actual REST call.  It imposes an
//       artificial delay to simulate the remote call and works against a
//       local temporary store.  This will be replaced as we integrate with
//       the actual Cloud Chamber service.

export class UsersProxy {
    // List all known user names
    public list(): Promise<UserList> {
        const path = "/api/users/"
        const request = new Request(path, {method: "GET"})

        return fetch(request)
            .then((resp: Response) => {
                failIfError(request, resp)

                return resp.json() as Promise<UserList>
            })
    }

    // Add a new user.
    public add(name: string, body: UserDetails): Promise<string> {
        const path = "/api/users/" + name
        const details = {
            password: body.password,
            rights: body.rights,
            enabled: body.enabled
        }

        const value = JSON.stringify(details)

        const request = new Request(path, {method: "POST", body: value })

        return fetch(request)
            .then ((resp) => {
                failIfError(request, resp)
                return "user " + name + " added"
            })
    }

    // Get the details for a user.
    public get(name: string): Promise<UserDetails> {
        const path = "/api/users/" + name
        const request = new Request(path, {method: "GET"})

        return this.getETagAndDetails(request)
    }

    // Update the details for a user.
    public set(name: string, body: UserDetails): Promise<UserDetails> {
        const path = "/api/users/" + name
        const details : UserUpdate = {
            rights: body.rights,
            enabled: body.enabled
        }

        const value = JSON.stringify(details)

        const request = new Request(
            path,
            {
                method: "PUT",
                body: value,
                headers: ETagHeader(body.eTag)
            })

        return this.getETagAndDetails(request)
    }

    // Remove a user.
    public remove(name: string): Promise<string> {
        const path = "/api/users/" + name
        const request = new Request(path, {method: "DELETE"})

        return fetch(request)
            .then((resp) => {
                failIfError(request, resp)

                return "User " + name + " deleted"
            })
    }

    // Set the password for a user
    public setPassword(name: string, body: UserDetails, oldPassword: string, newPassword: string): Promise<number> {
        const path = "/api/users/" + name + "?password"
        const msg : UserPassword = {
            oldPassword: oldPassword,
            newPassword: newPassword,
            force: false
        }

        const value = JSON.stringify(msg)

        const request = new Request(
            path,
            {
                method: "PUT",
                body: value,
                headers: ETagHeader(body.eTag)
            })

        return fetch(request)
            .then((resp) => {
                failIfError(request, resp)

                return getETag(resp)
            })
    }

    private getETagAndDetails(request: Request) : Promise<UserDetails> {
        let eTag: number

        return fetch(request)
            .then((resp: Response) => {
                failIfError(request, resp)

                eTag = getETag(resp)

                return resp.json() as Promise<PublicUserDetails>
            })
            .then((value) => {
                return {...nullDetails, ...value, eTag}
            });
    }
}
