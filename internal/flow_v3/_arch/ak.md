# definitions

## flow 
def: executing multiple actions to produce final result
provides data sources to actions  

## step
def: a single unit of work in a flow
talks to flow to get data 

## task
def: a single unit of work in a step
pure function, no side effects, works on input and outputs results
can be tested in isolation

# Flow example

[site.flow.md](site.flow.md)

clarify -> done -> plan
clarify -> error -> clarify_retry (max retries 2)
plan -> done -> task
plan -> error -> plan_retry (max retries 2)
task -> done -> implement
implement -> task_done -> implement
implement_next -> error -> implement_retry (max retries 2)
implement -> done -> finish