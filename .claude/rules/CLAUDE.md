# General rules:
- Write small files, preferably no more than 200 lines.
- Write code that is as clear as possible.
- Speak Russian.
- Commit every significant change, but avoid committing errors whenever possible.

# Refactoring rule:
- The user says they need to split file ```<N>``` because the file is probably large.
  The file splitting algorithm is as follows:
  ```text
  Every file contains some functions in any case. You take the file you need to split, read it, and then

  delete one function, a router, a class—it doesn't matter—and move it to another file. Then you come back and repeat the process until you've split the file completely, then check for errors if there are any.
  ```

# Deploy rule:
- When a user sends something like this: ```[root@vds1669329 CrewAi]#``` , it means that 
  the user is on the server and needs to be given the necessary commands if something is broken or 
  something needs to be checked. If it's code-related, you need to fix it locally and then push it.
  Exceptions (env files)

# Reserch rule:
- When a user sends you a link to a library or resource, before installing it, you should:
  ```bash
  fireclaw
  ```
  save it to ```.claude\docs``` in md format and examine the resulting file.
  After that, install the library and use it based on the information you've previously examined.

# Context rule:
- In each new session, create a folder in the ```.claude\context``` folder with the name of the task you'll be working on.
  Then, create an .md file called ```START.md``` and describe the task itself.
  After that, create ```PLAN.md```, where you describe the plan you'll follow.
  Then, create ```FINEL.md```, where you describe the steps you've completed and the files you've changed, their lines, and possibly any details.
  The user can refine any details. You can update ```PLAN.md``` by adding new steps and finalizing ```FINEL.md```.
- After completing the task (by creating a ```commit```), you create a new folder and complete the previous step.

# Refrash rule:
- Read this file periodically while updating your context to remember the rules listed above.
  I suggest doing this after each context update.