import { asArray, asBool, asString } from "../utils"

/* eslint-disable */
export const protobufPackage = "admin";

/** Rights contains the set of known action rights that a user holds. */
export interface Rights {
  /**
   * can_manage_accounts indicates that this user can create, modify, or
   * delete user accounts - both others, and its own.
   */
  canManageAccounts: boolean;
  /**
   * can_step_time indicates that this user can change the simulated time and
   * simulated time policy.
   */
  canStepTime: boolean;
  /**
   * can_modify_workloads indicates that this user can create, update, or
   * delete simulated workloads.
   */
  canModifyWorkloads: boolean;
  /**
   * can_modify_inventory indicates that this user can add, modify, or remove
   * the definition views of simulated inventory.
   */
  canModifyInventory: boolean;
  /**
   * can_inject_faults indicates that this user can inject forced faults
   * into the running simulation.
   */
  canInjectFaults: boolean;
  /**
   * can_perform_repairs indicates that this used can manually repair a fault
   * in the simulated inventory.
   */
  canPerformRepairs: boolean;
}

/**
 * Limited exposure of user attributes for use when returning information
 * to a remote client.
 */
export interface UserPublic {
  enabled: boolean;
  /** The rights this user has */
  rights: Rights | undefined;
  /** True, if this user entry can never be deleted. */
  neverDelete: boolean;
}

/**
 * Public definition for a user, sent by a remote client to the Cloud Chamber
 * controller when the user is created.
 */
export interface UserDefinition {
  password: string;
  enabled: boolean;
  /** The rights this user has */
  rights: Rights | undefined;
}

/**
 * Public definition for the fields for a user that can be mutated in an
 * update operation
 */
export interface UserUpdate {
  enabled: boolean;
  /** The rights this user has */
  rights: Rights | undefined;
}

/** Public definition for a request to set a new password for a user */
export interface UserPassword {
  /** The existing password, used to check that this is a legit request */
  oldPassword: string;
  /** The new password */
  newPassword: string;
  /** An administrative flag to force the new password */
  force: boolean;
}

/** User list response. */
export interface UserList {
  /** List of known users */
  users: UserList_Entry[];
}

/** A single user entry */
export interface UserList_Entry {
  /** Name of the user */
  name: string;
  /** Uri to use to access the user */
  uri: string;
  /** True, if the user is protected against changes */
  protected: boolean;
}

export const Rights = {
  fromJSON(object: any): Rights {
    if (object === undefined || object === null) {
      return {
        canManageAccounts: false,
        canStepTime: false,
        canModifyWorkloads: false,
        canModifyInventory: false,
        canInjectFaults: false,
        canPerformRepairs: false,
      }
    }

    return {
        canManageAccounts: asBool(object.canManageAccounts),
        canStepTime: asBool(object.canStepTime),
        canModifyWorkloads: asBool(object.canModifyWorkloads),
        canModifyInventory: asBool(object.canModifyInventory),
        canInjectFaults: asBool(object.canInjectFaults),
        canPerformRepairs: asBool(object.canPerformRepairs),
    }
  },
};

export const UserPublic = {
  fromJSON(object: any): UserPublic {
    return {
      enabled: asBool(object.enabled),
      rights: Rights.fromJSON(object.rights),
      neverDelete: asBool(object.neverDelete),
    }
  },
};

export const UserDefinition = {
  fromJSON(object: any): UserDefinition {
    return {
      password: asString(object.password),
      enabled: asBool(object.enabled),
      rights: Rights.fromJSON(object.rights),
    }
  },
};

export const UserUpdate = {
  fromJSON(object: any): UserUpdate {
    return {
      enabled: asBool(object.enabled),
      rights: Rights.fromJSON(object.rights),
    }
  },
};

export const UserPassword = {
  fromJSON(object: any): UserPassword {
    return {
      oldPassword: asString(object.oldPassword),
      newPassword: asString(object.newPassword),
      force: asBool(object.force),
    }
  },
};

export const UserList = {
  fromJSON(object: any): UserList {
    return {
      users: asArray<UserList_Entry>(UserList_Entry.fromJSON, object.users),
    }
  },
};

export const UserList_Entry = {
  fromJSON(object: any): UserList_Entry {
    return {
      name: asString(object.name),
      uri: asString(object.uri),
      protected: asBool(object.protected),
    }
  },
};
