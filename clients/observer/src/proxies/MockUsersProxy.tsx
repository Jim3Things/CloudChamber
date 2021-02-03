// This module contains the proxy handler for calling the REST User management
// service in the Cloud Chamber backend.

// Define the user details as supplied by the REST service.

// TODO: Investigate having these produced directly from the originating
//       protobuf file.

export class UserDetails {
    password: string;
    canManage: boolean;
    enabled: boolean;

    constructor() {
        this.password = "";
        this.canManage = false;
        this.enabled = false;
    }
}

// Utility class that provides a proxy to the Cloud Chamber User management
// service.

// TODO: This proxy current fakes out the actual REST call.  It imposes an
//       artificial delay to simulate the remote call and works against a
//       local temporary store.  This will be replaced as we integrate with
//       the actual Cloud Chamber service.

export class MockUsersProxy {
    // This is the temporary debug store.
    users: Record<string, UserDetails> = {
        "Admin": {
            password: "SomePassword",
            canManage: true,
            enabled: true
        },
        "Alice": {
            password: "SecondPassword",
            canManage: false,
            enabled: true
        }
    };

    // List all known user names
    public list(): Promise<string[]> {
        return new Promise<string[]>((resolve, reject) => {
            let list: string[] = [];
            for (let key in this.users) {
                list = [...list, key];
            }
            setTimeout(() => resolve(list), 100);
        });
    }

    // Add a new user.
    public add(name: string, body: UserDetails): Promise<string> {
        return new Promise<any>((resolve, reject) => {
            setTimeout(() => {
                if (this.users[name] === undefined) {
                    this.users[name] = body;
                    resolve("user " + name + " added");
                } else {
                    reject("user " + name + " already exists");
                }
            }, 100);
        });
    }

    // Get the details for a user.
    public get(name: string): Promise<UserDetails> {
        return new Promise<UserDetails>((resolve, reject) => {
            setTimeout(() => {
                if (this.users[name] !== undefined) {
                    resolve(this.users[name])
                } else {
                    reject(null);
                }
            }, 100);
        });
    }

    // Update the details for a user.
    public set(name: string, body: UserDetails): Promise<any> {
        return new Promise<any>((resolve, reject) => {
            let updated: Record<string, UserDetails> = {};
            for (let key in this.users) {
                if (name === key) {
                    updated[key] = body;
                } else {
                    updated[key] = this.users[key];
                }
            }

            setTimeout(() => {
                this.users = updated;
                resolve();
            });
        });
    }

    // Remove a user.
    public remove(name: string): Promise<string> {
        return new Promise<string>((resolve, reject) => {
            let updated: Record<string, UserDetails> = {};
            let found = false;

            // Never delete the system account
            if (name !== "Admin") {
                for (let key in this.users) {
                    if (name === key) {
                        found = true;
                    } else {
                        updated[key] = this.users[key];
                    }
                }
            }

            setTimeout(() => {
                if (found) {
                    this.users = updated;
                    resolve();
                } else {
                    // This is just a placeholder error...
                    reject("user " + name + " could be not be deleted.")
                }
            });
        });
    }
}