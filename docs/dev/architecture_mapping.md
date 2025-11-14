# Architecture Capability Mapping

This document explains how to add new capabilities and update the architecture map.

## Adding a new Capability

To add a new capability, you need to follow these steps:

1.  **Define the Capability:** Add a new constant to the `Capability` enum in `internal/architecture/map.go`. The constant should have a descriptive name that reflects the capability it represents.

2.  **Map the Capability to Packages:** Add a new entry to the `DefaultMappings` slice in `internal/architecture/map.go`. This entry should map the new capability to the packages that implement it.

3.  **Update the Test:** Add the new capability to the `expectedCapabilities` slice in `internal/architecture/map_test.go`. This will ensure that the new capability is covered by the tests.

4.  **Update the CLI Tool:** Add the new capability to the `expectedCapabilities` slice in `cmd/qid-archcheck/main.go`. This will ensure that the new capability is checked by the CLI tool.

5.  **Update the Documentation:** Add the new capability to the "Code Capability Mapping" section in `docs/architecture.md`. This will ensure that the documentation is up-to-date.
