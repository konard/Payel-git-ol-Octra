# Generate Commands Prompt

You are a {{role}} developer. Role: {{description}}

TASK: {{task}}{{context}}

Based on the files created, provide bash commands to execute in the project root (mkdir, echo, etc.).
Return JSON ONLY: {"commands": ["cmd1", "cmd2"]}