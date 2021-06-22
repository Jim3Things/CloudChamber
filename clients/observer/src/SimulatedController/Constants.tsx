// This module contains temporary definitions for the impact claims on the
// controller components.  Once the simulated controller components have been
// implemented and operating these definitions will be superseded or fully
// replaced.

import { Theme } from "@material-ui/core/styles"

export enum Impact {
    ImpactNone = 0,
    ImpactRead = 1,
    ImpactWrite = 2,
    ImpactSelected = 3
}

// Impacts defines the impact statement for each controller element.
export interface Impacts {
    WorkloadGoal: Impact,
    InventoryGoal: Impact,

    RollingUpgrade: Impact,
    PhasedUpdate: Impact,
    AutoScaler: Impact,

    AddInventory: Impact,
    RetireInventory: Impact,
    BurnIn: Impact,

    WorkloadTargetPartial: Impact,
    InventoryTargetPartial: Impact,

    Scheduler: Impact,

    WorkloadTargetFull: Impact,
    InventoryTargetFull: Impact,

    WorkloadRepairManager: Impact,
    InventoryRepairManager: Impact,

    Observed: Impact,
    WorkloadActions: Impact,
    InventoryActions: Impact,

    Monitor: Impact,
    Effector: Impact
}

// NoImpacts acts as the 'quick reset' for the impact statements.  It is used
// to set the initial state, and to set the default values on all impact claims.
export const NoImpacts : Impacts = {
    WorkloadGoal: Impact.ImpactNone,
    InventoryGoal: Impact.ImpactNone,

    RollingUpgrade: Impact.ImpactNone,
    PhasedUpdate: Impact.ImpactNone,
    AutoScaler: Impact.ImpactNone,

    AddInventory: Impact.ImpactNone,
    RetireInventory: Impact.ImpactNone,
    BurnIn: Impact.ImpactNone,

    WorkloadTargetPartial: Impact.ImpactNone,
    InventoryTargetPartial: Impact.ImpactNone,

    Scheduler: Impact.ImpactNone,

    WorkloadTargetFull: Impact.ImpactNone,
    InventoryTargetFull: Impact.ImpactNone,

    WorkloadRepairManager: Impact.ImpactNone,
    InventoryRepairManager: Impact.ImpactNone,

    Observed: Impact.ImpactNone,
    WorkloadActions: Impact.ImpactNone,
    InventoryActions: Impact.ImpactNone,

    Monitor: Impact.ImpactNone,
    Effector: Impact.ImpactNone
}

// impactToColor converts an impact claim into a background color.  This is used
// temporarily to highlight expected data flows in the controller.
export function impactToColor(impact: Impact, theme: Theme): string {
    switch (impact) {
        case Impact.ImpactSelected:
            return theme.palette.background.default

        case Impact.ImpactNone:
            return theme.palette.action.disabledBackground

        case Impact.ImpactRead:
            return "yellow"

        case Impact.ImpactWrite:
            return "red"
    }
}
