# flow

## idea

A flow is a state machine that moves one piece of typed data through a chain of steps.

- flow -> step -> task
  - flow: owns the graph and the active step, runs a step, reads its transition, moves to the next step
  - step: one state of the machine, reads and updates the flow data through the service
  - task: plain work a step calls (an ai ask, a file read), gets values in, returns values out

## how a flow is constructed

- a graph definition lists steps and their transitions
  - `step: { transition: next }` or `transition: { next, retry }`
  - the first step in the graph is the start step
  - `finish` as a target ends the flow
- each step name in the graph is bound to one step exec
- every step of a flow shares the same data type, picked by the flow (generics)
  - one flow implementation runs many flows, each with its own data type

## how data moves

- the caller creates a data service with the initial data and passes it to the flow
- the flow passes the same service to every step, the flow itself holds no data
- a step reads with `Data()` and writes with `Update` (or `SetData` to replace)
  - `Update` is the default, it is an atomic read-modify-write
  - a value from `Data()` is a snapshot, never mutate it in place
- a task never sees the service, it never mutates flow or step data
- input comes from the initial data, not from the context

## transitions

- a step returns a transition and an error
  - `done` and named transitions move along the graph
  - `error` with an error is recoverable, the flow follows the `error` edge
  - a self `error` edge is a retry, limited by its `retry` count
  - any other transition with an error stops the flow
  - `unknown` or an empty transition stops the flow
  - `finish` ends the flow

## rules - interfaces and types

- define the interface before the type, the type implements it, never the other way round
- every implementation is asserted with `var _ I = (*Impl)(nil)`
- interface definitions are upgraded, not rewritten
  - adding methods is fine, removing or changing signatures is not
  - private methods are desired, they keep the machine internals in the package
- flows, steps and services are generic over the data type, no `any` data bags
- steps depend on the service interface, never on the concrete service
- external dependencies (ai, fs, db) are reached through small interfaces
