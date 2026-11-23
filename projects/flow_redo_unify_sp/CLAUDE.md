# flow redesign

in this flow redesign I'm trying to create a flexible flow focused on interface first design

idea: easy creation of flows with steps 
 - stepping on good interface and generic design
 - easy mapping of branching flows 
   - step lead to different steps based on the outcome

# flow control
flow controls the data and env access while steps are iterations in the program that needs to be executed
 - task are unit of works witch are pure 
   - they only get input return output nut dont change env 
   - task might talk to API's

# definitions
flow: the main state machine (exec environment)
step: a single step/iteration in the flow 
  - uses service providers modifies the flow state
task: unit of work ran in a step
  - gets input, returns output, 
  - uses service providers 
    - the SP can be initialized in the step not the same type as the flow SP