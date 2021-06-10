# Intro
This is a list of some of the best practices using this tool. Please contribute if you have any additions or changes to current or non existing best practices

* [`Different envs`](#different-envs).

# Different envs
We primarily use 2 methods:

* A single file, where each environment is a parameter, see prompts and options, where you can populate one with an environment variable env:STAGE and STAGE=dev teller ...
* Keep a config file per environment, similar to what you would do with .env file (.env.example, .env.production, etc.) -- but with teller none of configuration files contain any sensitive information (as opposed to .env) so you're safe.

The best practice really depends on the size of your team and how you prefer to work. We imagine if the team is small, and the use cases are not many, a single file would be great. If the team is large, or maybe you're enabling other teams -- keeping a file per environment would be better, and this way you can "distribute" your teller files per use case in a central way.