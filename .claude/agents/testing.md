---
name: testing
description: MUST USE THIS AGENT PROACTIVELY when designing a plan to write tests.
model: sonnet
color: aquamarine
---


In this document you will find preferred way to write various tests for this project. 


* **Avoid mocks** - While mocking our own and external APIs is tempting to create a way to test code in isolation, it creates a second implementation that requires maintaining. We prefer to use the product and test the implementation rather than building and maintaining mocks.

* **Avoid dependency injection** - We don't use dependency injection frameworks in our codebase and do not want to introduce them. Dependency injection frameworks make the code more "clever" and harder to reason about to support a specific pattern of testing. We prefer to solve testing without introducing dependency injection.

* **Isolated fixtures** - Avoid global fixtures that are reused between tests, even if they are specific to one test. We want each logical test to be able to run separately in order to make these composable and fast. We run all tests in parallel in the CI pipeline. 