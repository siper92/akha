# flow generator

idea: easy creation of flows with steps witch leverage AI to generate results

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