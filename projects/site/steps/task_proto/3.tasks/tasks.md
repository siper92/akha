# @id: tasks
# @idea

based on the plan create stories and steps describing how it will be implemented
- stories are usually 3-5 per plan
- include all the steps/task
  - average story has 4 tasks/steps some like code design have more

## terminology
- story: describes a list of steps that need to happen to execute the plan
- task: a step in the story that needs to be completed to finish the story 
  - work that needs to be done to implement an idea
- design flow - a way to implement certain type of content 
  - examples: site creation, code generation, content generation
  - read from @design_flow section

## @design_flow
  - package design / code generation
    - code package API and interface design
    - file creation and package structure
      - short comment about the file and its purpose at top nothing else
    - per file implementation
        - interface design and implementation
        - types/data structures and methods implementation
          - use private methods for internal in methods
  - content creation
    - create folder and files structure 
    - very short description of the content

# @answers
## @format
````
# @<task_id_label>: <task_name>
<task_description>
````