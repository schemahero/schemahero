# Proposal Implementation

You are tasked with implementing a detailed and approved technical proposal in this code. This command allows you to understand the proposal and proceed with the implementation.

## Initial Response

When invoked WITH parameters and when the parameter is a filename in the proposals directory:
```
I'll get started implementing [filename]. Let's first check if there an any questions before I start.
```

When invoked WITHOUT parameters:
```
Tell me the filename of the proposal you'd like implemented
```

When invoked WITH a parameter but the parameter doesn't match a proposal filename in the `proposals` directory:
```
I can't find that file. Tell me the filename of the proposal you'd like me to implement.
```

## Research and Implementation Plan

Along with the implementation plan, there likely is a file that has `_research` appended to the filename. This is where all thoughts and research for various options have been documented. While you should primarily base your implementation on the provided proposal/implementation doc, the _research is available if you need to scan and understand some of the background.

## Separate PRs

If the implementation plan contains a section that shows separate PRs being made, limit your work to the next PR only. When completed, update the proposal to indicate the PR has been implemented so that next run, you will know to start on the next phase.

## Subagents

When writing code, use the following subagents, in addition to normal agents:

- go-developer: this subagent is used to follow patterns we want for Go code.