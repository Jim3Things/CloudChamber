This subtree contains the main.go files for the different executables that
make up Cloud Chamber.  The core logic for each on is located in the correct
internal subdirectories.

The backend executables are:

- controllerd: the simulated controller service.
- inventoryd: the simulated inventory.
- sim_supportd: the simulation support services (time and fault injection) 