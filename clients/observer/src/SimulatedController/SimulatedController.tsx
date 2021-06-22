import { Grid } from '@material-ui/core';
import { createStyles, makeStyles } from '@material-ui/core/styles';
import { Container } from '../common/Cells';
import { impactsSelector, useAppSelector } from '../store/Store';
import { Impact, NoImpacts } from './Constants';
import { LogicElement } from './LogicElement';
import { StorageElement } from './StorageElement';

const useStyles = makeStyles((theme) =>
    createStyles({
        layer: {
            paddingBottom: theme.spacing(2)
        }
    }),
)

function Layer(props: any) {
    const classes = useStyles()

    return <Grid container XS={12} className={classes.layer} {...props} />
}

// Display the simulated controller's components and stores in a logical
// structure.  Emphasis is on showing the layered relationship between the
// stores and logic components.
//
// Note that logic elements are active, in that they can be clicked to show
// what storage elements they read, and which ones they modify.
export function SimulatedController() {
    const impacts = useAppSelector(impactsSelector)

    return (
        <Container xs={12}>
            <Layer justify="space-around">
                <StorageElement
                    impact={impacts.controllerImpacts.WorkloadGoal}
                    title="Workload Goal State" />

                <StorageElement
                    impact={impacts.controllerImpacts.InventoryGoal}
                    title="Inventory Goal State" />
            </Layer>

            <Layer xs={6} justify="space-evenly">
                <LogicElement
                    impact={impacts.controllerImpacts.RollingUpgrade}
                    selectedImpacts={{
                        ...NoImpacts,
                        WorkloadGoal: Impact.ImpactRead,
                        WorkloadTargetPartial: Impact.ImpactWrite,
                        RollingUpgrade: Impact.ImpactSelected,
                    }}
                    title="Rolling Update" />

                <LogicElement
                    impact={impacts.controllerImpacts.PhasedUpdate}
                    selectedImpacts={{
                        ...NoImpacts,
                        WorkloadGoal: Impact.ImpactRead,
                        WorkloadTargetPartial: Impact.ImpactWrite,
                        PhasedUpdate: Impact.ImpactSelected,
                    }}
                    title="Phased Update" />

                <LogicElement
                    impact={impacts.controllerImpacts.AutoScaler}
                    selectedImpacts={{
                        ...NoImpacts,
                        WorkloadGoal: Impact.ImpactRead,
                        WorkloadTargetPartial: Impact.ImpactWrite,
                        AutoScaler: Impact.ImpactSelected,
                    }}
                    title="Auto-Scaler" />
            </Layer>

            <Layer xs={6} justify="space-evenly">
                <LogicElement
                    impact={impacts.controllerImpacts.AddInventory}
                    selectedImpacts={{
                        ...NoImpacts,
                        InventoryGoal: Impact.ImpactRead,
                        InventoryTargetPartial: Impact.ImpactWrite,
                        AddInventory: Impact.ImpactSelected,
                    }}
                    title="Add Inventory" />

                <LogicElement
                    impact={impacts.controllerImpacts.RetireInventory}
                    selectedImpacts={{
                        ...NoImpacts,
                        InventoryGoal: Impact.ImpactRead,
                        InventoryTargetPartial: Impact.ImpactWrite,
                        RetireInventory: Impact.ImpactSelected,
                    }}
                    title="Retire Inventory" />

                <LogicElement
                    impact={impacts.controllerImpacts.BurnIn}
                    selectedImpacts={{
                        ...NoImpacts,
                        InventoryGoal: Impact.ImpactRead,
                        InventoryTargetPartial: Impact.ImpactWrite,
                        BurnIn: Impact.ImpactSelected,
                    }}
                    title="Burn-In" />
            </Layer>

            <Layer justify="space-around">
                <StorageElement
                    impact={impacts.controllerImpacts.WorkloadTargetPartial}
                    title="Workload Target State (partial)" />

                <StorageElement
                    impact={impacts.controllerImpacts.InventoryTargetPartial}
                    title="Inventory Target State (partial)" />
            </Layer>

            <Layer justify="center">
                <LogicElement
                    impact={impacts.controllerImpacts.Scheduler}
                    selectedImpacts={{
                        ...NoImpacts,
                        WorkloadTargetPartial: Impact.ImpactRead,
                        WorkloadTargetFull: Impact.ImpactWrite,
                        InventoryTargetPartial: Impact.ImpactRead,
                        InventoryTargetFull: Impact.ImpactWrite,
                        Scheduler: Impact.ImpactSelected,
                    }}
                    title="Scheduler" />
            </Layer>

            <Layer justify="space-around">
                <StorageElement
                    impact={impacts.controllerImpacts.WorkloadTargetFull}
                    title="Workload Target State" />

                <StorageElement
                    impact={impacts.controllerImpacts.InventoryTargetFull}
                    title="Inventory Target State" />
            </Layer>

            <Layer justify="space-around">
                <LogicElement
                    impact={impacts.controllerImpacts.WorkloadRepairManager}
                    selectedImpacts={{
                        ...NoImpacts,
                        WorkloadTargetFull: Impact.ImpactRead,
                        Observed: Impact.ImpactRead,
                        WorkloadActions: Impact.ImpactWrite,
                        WorkloadRepairManager: Impact.ImpactSelected,
                    }}
                    title="Workload Repair Manager" />

                <LogicElement
                    impact={impacts.controllerImpacts.InventoryRepairManager}
                    selectedImpacts={{
                        ...NoImpacts,
                        InventoryTargetFull: Impact.ImpactRead,
                        Observed: Impact.ImpactRead,
                        InventoryActions: Impact.ImpactWrite,
                        InventoryRepairManager: Impact.ImpactSelected,
                    }}
                    title="Inventory Repair Manager" />
            </Layer>

            <Layer justify="space-around">
                <StorageElement
                    impact={impacts.controllerImpacts.Observed}
                    title="Observed State" />

                <StorageElement
                    impact={impacts.controllerImpacts.WorkloadActions}
                    title="Workload Actions" />

                <StorageElement
                    impact={impacts.controllerImpacts.InventoryActions}
                    title="Inventory Actions" />
            </Layer>

            <Layer justify="space-around">
                <LogicElement
                    impact={impacts.controllerImpacts.Monitor}
                    selectedImpacts={{
                        ...NoImpacts,
                        Observed: Impact.ImpactWrite,
                        Monitor: Impact.ImpactSelected
                    }}
                    title="Monitor" />

                <LogicElement
                    impact={impacts.controllerImpacts.Effector}
                    selectedImpacts={{
                        ...NoImpacts,
                        InventoryActions: Impact.ImpactRead,
                        WorkloadActions: Impact.ImpactRead,
                        Effector: Impact.ImpactSelected,
                    }}
                    title="Effector" />
            </Layer>
        </Container>
    )
}
