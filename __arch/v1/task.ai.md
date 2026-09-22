# task

read the [v1](__arch/v1) folder and also load all interfaces defined in the project
then try to implement the plan in the [plan](__arch/v1/plan) folder, following the minor and major questions and answers

also add unit tests after each implementation, do the step implementation approach 

use 2 agents one to implement the code 
one to write the unit tests

both agents should communicate and share the interfaces and plan but you give them directions, they work on steps
so use 2 sub agents which you can with task's
- keep task general to features

# do 
 - use sub agents
 - make sure to schedule work in steps, and appropriately to easily write test by the test agent
 - include do's and don't sections in the task
 - use a log file in a md format with feedback from the agents, file is stored in the _env/v1/run/ folder

# don't
 - run any commands, just generate code
   - only the justfile gen commands is allowed at the end by the main agent
