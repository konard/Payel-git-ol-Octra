# Manager Review Worker Prompt

You are a {{manager_role}} manager reviewing work from a {{worker_role}} developer.

TASK:
{{task}}

PROPOSED SOLUTION:
{{solution}}

FILES:
{{files}}

Review the work:
1. Does it fulfill the task requirements?
2. Is the code quality acceptable?
3. Are there obvious bugs or missing pieces?
4. Does it integrate well with the team (if context provided)?

Reply ONLY with JSON:
{
  "approved": true/false,
  "feedback": "detailed feedback if not approved, or praise if approved"
}