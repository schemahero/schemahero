# Proposal Author

You are tasked with writing a proposal for new or edited functionality in this code. This command allows you to understand the goal and help produce a detailed proposal.

## Initial Response

When invoked WITH parameters:
```
I'll help you write a proposal for [summary]. Let's first check if this idea requires a proposal.
```

When invoked WITHOUT parameters:
```
I'll help you think through a new proposal.

Please describe what your goals are:
- What is the desired change?
- Do you have any initial thoughts on how you'd like to implement?
```

## Follow up response



## Subagent use

You SHOULD use the following subagents (and any subagents they recommend) to help the user with their request:

- proposal-needed
- proposal-writer

## Shortcut (tickets and stories)

If the user's request references a ticket or shortcut story, use the shortcut agent to find the story. 
Never update the shortcut story with anything. At this time, you should treat Shortcut as a readonly API.


## Follow up instructions from the user

At any time the user may reject your recommendation. They may accept the research and reject the proposal or simply reject both. When this happens, regardless of the step you are at, if the user provides additional context, you should ALWAYS restart the entire process. Read the current research and proposal documents if they were created, use them + the code base + the user's additional context and recreate these docs from scratch.
