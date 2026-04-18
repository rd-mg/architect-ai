# Spec: Hook Surface

## Requirement: Global Hook Registration
The system MUST allow registration of global pre-task and post-task hooks.

#### Scenario: Registration
- GIVEN a hook function
- WHEN hooks.RegisterPreTask(fn) is called
- THEN the function is added to the global registry.

## Requirement: Safe Execution
Hooks MUST NOT crash the main application flow.

#### Scenario: Hook Panic
- GIVEN a hook that panics
- WHEN the hook is fired
- THEN the panic is caught, logged, and the main process continues.
