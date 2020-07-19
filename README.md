# Sanctions Search

Thank you for taking the time to complete our code challenge!

For this problem, you will build an application for searching publicly available sanction data. You may take as long as you like with your solution. The data provided is just one of many different sources of sanctions data. So, while we have not asked you to build other integrations, please take into account the idea that your solution will be thought about in the context of a more general set of datasets.

## Important

A couple notes to hopefully help make this a low-stress experience:

First, please do not hesistate to ask us any questions you may have. You will not be penalized for asking questions; your questions help us improve the clarity of the instructions so it benefits you and future candidates.

Second, do not be shy about using your preferred language, libraries, databases, or frameworks. We do not expect you to learn our stack for this challenge; your time is better spent using the tools you know that are suitable for the problem at hand.

Third, if you need to learn anything while working on the challenge, that's okay! We know you have many skills and experiences, so it's not a strike against you if you need to do some reading.

Finally, we value your time so we don't expect you to spend more time than necessary on polish. Just focus on the fundamentals! Feel free to include a README.md in your solution covering polish you would make if you have ideas for improvements.

----

# Requirements

Look for **must** to indicate a requirement.

In short, your solution:

0. may use any language, libraries, database, or frameworks
1. must implement the required [API](#api)
2. must be dockerized.
   - Please update [Dockerfile](./Dockerfile) and [docker-compose.yml](./docker-compose.yml) as needed.
   - Please build any dependencies within one or more Docker images
3. must include one or more [unit or integration test](#testing)
4. must pass the [smoke-tests](#smoke-tests)

----

## API

Your solution will consist of an API that provides search functionality against the EU Sanctions list.

#### Bootstrapping

You can consume sanctions from the EU in [CSV](https://sigmaratings.s3.us-east-2.amazonaws.com/eu_sanctions.csv) or [XML](https://sigmaratings.s3.us-east-2.amazonaws.com/eu_sanctions.xml) format.

Your solution **must** be able to fetch and load the data on startup. In other words, when we receive your solution, we will run it with `make run` and it should fetch the data file and bootstrap the database automatically.

#### `GET /search`
Your server should have a `/search` endpoint that takes a `name` query parameter with a person's name. It **must** respond with an array of matches with the following shape:
```json
{
  "logicalId": 98765,
  "matchingAlias": "Kim Jung Un",
  "otherAliases": ["Rocket Man"],
  "relevance": 0.92,
}
```

* `logicalId` **must** be the "entity logical id" of the matching alias and **must** be unique per object in any response
* `relevance` **must** be a float in the range 0 to 1 (inclusive) indicating how close the result is to the users search
  - a `relevance` value of 1 indicates an exact string match between the search and either the name or one of the aliases.
* `matchingAlias` **must** be the "whole name" for a given "entity logical id" with the strongest `relevance`
* `otherAliases` **must** be the other aliases for the same logical id

#### `GET /status`
In order to communicate to the smoke-test when the server is ready, your solution **must** include a `/status` endpoint that returns an error code until the bootstrapping is complete and the server is ready to serve requests.

----

## Testing

We'd like to know how you think about testing. There are many valid ways to approach testing, so the only requirements are:
1. you **must** include one or more unit or integration tests
2. `make test` **must** run your test(s) _inside a docker container_

In order to avoid testing taking an unreasonable amount of time, it is okay to write one meaningful test case and stub some additional test cases to indicate what you feel is important to test. We do not expect you to explain how you would achieve 100% test coverage; we just want to know where you would focus your efforts if you had the time.

----

## Smoke Tests

We provide a smoke test that can be run against your api with `make smoke-test` for some minimal verification. We try to keep the tests up-to-date, but the sanctions list changes over time, so let us know if the cases don't seem to match what you see in the actual sanctions list.

Your solution **must** include the smoke tests and they must continue to run with `make smoke-test`. If you feel the need to alter the smoke tests other than adding additional cases, please check with us first because we use them for reviewing submissions.

The API is already stubbed with fixtures that pass all the tests to serve as an naive sample implementation. If `make smoke-test` does not work for you and you believe you have `docker` and `docker-compose` installed and up-to-date, please let us know.

To ensure the `make smoke-test` works properly, make sure you update the value of `API` for the `test` service in [docker-compose.yml](./docker-compose.yml).

We may run additional test cases against your solution.
