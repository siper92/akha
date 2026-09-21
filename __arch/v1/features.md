# Akha - platform for AI workflows

Akha is a platform for processing AI workflows, similar to n8n and temporal, written in Go

## key parts
 - backend service - lives on a server/same host
   - talks to AI modules
   - authenticate workers
     - issues JWT tokens for workers
 - worker - runs on a host machine
   - execute workflows
   - communicate with backend service 
     - authenticates it self with JWT token

##  how it works

*.ak script files are processed by the worker

- see: [lang.spec](lang.spec)

## features

### backend service
- authenticates workers
  - issues JWT tokens for workers
- talks to AI modules / LLMs integrations
  - stores templates for talking to LLM's
    - e.g. prompt templates
- stores logs
- reads/writes to the database

### worker
- workflow execution
- modifies files / IO operations
- communicates with backend service

# Code structure

```
akha/
  - _defs/ - definitions for code generation
    - db/ - sqlc definitions
    - proto/ - protobuf definitions
  - lang/ - language spec and parser
    - lexer/ - lexer for the language
    - token/ - token definitions for the language
    - parser/ - parser for the language
    - ast/ - abstract syntax tree for the language
  - backend/ - backend service
  - worker/ - worker service
  - sdk/ - SDK for db and API calls
```
