# Flow

Flow process
 - steps -> task
   - sub steps -> task
----
flow, step, task

flow:
 - execution plan
 - processes steps 
 - handles services injection

step:
 - executes step of the plan
 - handles multiple tasks same in parallel

task:
 - executes unit of work
 - single action units with results
----

machine -> processes task -> returns output of the step