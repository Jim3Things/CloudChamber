# Cloud Chamber

Cloud Chamber is an exploration tool that provides a visible simulation of a cloud infrastructure.  This repository
contains the backend service portion of the tool.  The `https://github.com/Jim3Things/cloud_chamber_react_ts` repository
provides the UI.

Cloud Chamber provides a deep simulation experience with:

- Visible representation of the control plane services and the simulated environment.
- Direct control over the passage of simulated time.
- Full lifecycle support for workload operations, from creation through removal.
- Full support for inventory changes, from repair simulation through planned replacement
- Controlled injection of hard and intermittent faults into any component.
- Tracing streams that show how each event impacts the simulation.
- Descriptions in the traces for the reasons behind control decisions.

Cloud Chamber supports exploration of design alternatives through a modular structure that allows replacement of
control plane components.  Scheduling and repair management components are specifically designed to be easily
replaced.

The initial version of Cloud Chamber focuses on a single cluster view and on the compute portion of the control plane.
Later versions will broaden this, both to other parts of the control plane, such as network management, and to
larger portions of a cloud, such as multi-cluster regions.

Cloud Chamber is currently under development, and in its early milestones.


## Design Principles

Link TBD

## Dependency Choices

Link TBD

## Contributing

Link TBD
