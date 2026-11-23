# plan - AIStepFlowTasker

## goal

- a flow (state machine) runs steps
- a step runs tasks with services given by the flow
- a task calls the AI, the AI uses file tools until the task is done

## layout

```
go.mod                  module aiflow (⚠️ module name unknown, rename if v4 joins the root module)
cmd/main.go             wiring + the 3 task records
config/config.go        hardcoded values behind Get funcs
debug/debug.go          global debug service, stdout + .cache/debug/debug-<timestamp>.txt
vfs/vfs.go              FS interface
vfs/memory.go           in memory FS (tests)
vfs/disk.go             disk FS locked to one dir (os.Root)
ai/ai.go                Client, Tool, Schema, Request, Response
ai/tool.go              generic NewTool[In]
ai/anthropic.go         anthropic SDK client + tool loop
tools/files.go          createFile, readFile, deleteFile over vfs.FS
flow/flow.go            Step, Transition, Runner, Flow engine
flow/services.go        ServiceProvider (step level)
flow/services_task.go   TaskServiceProvider (task level)
flow/repository.go      TaskRecord, Status, TaskRepository, MemoryRepository
flow/task.go            Task interface, AITask
flow/steps.go           State, PickStep, ExecuteStep, RecordStep, NewTasker
```

## dependencies

- config <- debug <- ai <- tools, flow <- cmd
- vfs <- tools
- only `ai/anthropic.go` imports the anthropic SDK

## api

### config

- config.GetCacheDir() string
- config.GetWorkDir() string
- config.GetModel() string
- config.GetMaxTokens() int64
- config.GetMaxToolRounds() int

### debug

- debug.Logger
  - Logger.Log(format string, args ...any)
- debug.Service() Logger
- fileLogger - prints with fmt, appends the same line to the debug file

### vfs

- vfs.FS
  - FS.WriteFile(ctx, name, content string) error
  - FS.ReadFile(ctx, name string) (string, error)
  - FS.DeleteFile(ctx, name string) error
- vfs.Memory - NewMemory()
- vfs.Disk - NewDisk(dir)
- vfs.ErrNotFound

### ai

- ai.Client
  - Client.Run(ctx, Request) (Response, error)
- ai.Tool
  - Tool.Name() string
  - Tool.Description() string
  - Tool.Schema() Schema
  - Tool.Call(ctx, input json.RawMessage) (string, error)
- ai.Schema{Properties, Required}
- ai.Request{System, Prompt, Tools}
- ai.Response{Text, ToolCalls}
- ai.NewTool[In](name, description, schema, fn func(ctx, In) (string, error)) *FuncTool[In]
- ai.Anthropic - NewAnthropic()
- conversation{api, req, tools, messages}.exec(ctx)
- errors: ErrRefused, ErrTruncated, ErrMaxRounds, ErrBadInput

```
conversation.exec:
  for round < maxRounds:
    msg = Messages.New(model, system, messages, tools)
    refusal -> ErrRefused, max_tokens -> ErrTruncated
    messages += msg
    for block in msg: text -> keep, tool_use -> tool.Call -> tool_result
    stop != tool_use -> return Response
    messages += user(all tool_results)
  ErrMaxRounds
```

- one prompt can trigger many tool calls over many rounds
- a failing tool is sent back as an error result, the AI decides what to do

### tools

- tools.Files(fs vfs.FS) []ai.Tool
  - createFile{fileName, content}
  - readFile{fileName}
  - deleteFile{fileName}

### flow

- flow.Step[SP ServiceProvider, T any]
  - Step.Name() string
  - Step.Exec(ctx, sp SP, data T) (Transition, error)
- flow.Transition - Next(step), End()
- flow.Runner[T]
  - Runner.Run(ctx, data T) error
- flow.Flow[SP, T] - New(sp, steps...), first step is the start
- flow.ServiceProvider
  - Repository() TaskRepository
  - Tasks() TaskServiceProvider
- flow.TaskServiceProvider
  - AI() ai.Client
  - Tools() []ai.Tool
- flow.Task[SP, In, Out]
  - Task.Exec(ctx, sp SP, in In) (Out, error)
- flow.AITask - TaskInput{Title, Description} -> TaskOutput{Text, ToolCalls}
- flow.TaskRecord{ID, Title, Description, Status, Result}
- flow.TaskRepository
  - NextPending(ctx) (TaskRecord, bool, error)
  - Save(ctx, TaskRecord) error
  - List(ctx) ([]TaskRecord, error)
- flow.MemoryRepository - NewMemoryRepository(records...)
- flow.State{Current, Output, Processed}
- flow.NewTasker(sp) (*Flow[ServiceProvider, *State], error)
- errors: ErrNoSteps, ErrDuplicateStep, ErrUnknownStep, ErrTaskNotFound

## state machine

```
pick    -> execute   a pending task record exists
pick    -> end       no pending task records
execute -> record    always, failure is kept in the state
record  -> pick      task done
record  -> end       task failed
```

## data rules

- flow owns the state (`*State`) for the whole run
- steps read and change the state, reach services only through the SP
- tasks get input from the step and return output to the step, no state access
- tasks reach AI and tools only through the task SP
- debug and config are global funcs

## not done here

- tests (task forbids)
- go commands (`go mod tidy` is left to the user)
- refusal fallback models, the run stops with ErrRefused
