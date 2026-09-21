# task

help me design a parser for a custom language
the idea is to design a flow language to accomplish similar usage as n8n
I'm thinking of solidity style machine that reads a *.ak (or other still to be chosen) file and creates/execites an execution flow that uses a Go VM to execute it

break it in to steps to design the VM (virtual machine) that will execute the flow
focus more on the design of the whole system not the implementation of the VM itself

1. design a lexer/parser for the language
 - the language must allow easy addition of modules
 - will support conditional execution, loops, and function calls - but in a later step as this is a design of the base structure
2. design a flow execution engine that will execute the parsed flow

write a task.res.ai.md file that describes the task in detail
 - focus on clarifying questions and requirements
 - take in account a more rich feature set than the simple example for now we are building the base structure
   - example of a language spec can be found in ./__arch/v1/lang.spec

## key features
- easy to add new modules (functions) to the language
- tracable execution flow with logging and error handling
- stable and secure execution environment (sandboxed) 
  - keeps logs and handles retries and errors

## idea

reads a *.ak file and creates a execution flow that uses a Go VM to execute it
example:
```
Akha.Log("Hello World")
```
logs in the current dir akha.log file

the spec for the language can be found - ./__arch/v1/lang.spec

# do
 - suggest other alternative technologies which you think have similar idea
   - preferably open source and free to use
 - add a specific section on top with design approach and architecture for these type of systems

# don't 
 - read/write files not listed here
 - exist the current dir