# SimCLI
Utility to mock CLI's, simulate responses, add delays/repeats, and more. Assists testing and cross-platform development that involves executing external programs

## Usage

### Building / Running

```sh
# clone this respond
git clone git@github.com:davidwashere/simcli.git

# build / install globally (assumes go bin dir on path)
go install ./cmd/simcli

# then run
simcli hello
```

### Config
`SimCLI` requires a config file defining the tasks to execute when specific args are supplied to `simcli`

Refer to [simcli.yaml](simcli.yaml) for an example config

By default `simcli.yaml` is expected in the current working directory, to change this path set it in the `SIMCLI_CONFIG` environment variable

A `task` defines what to do, a `command` chains tasks together and executes them when `args` are matched

**Config Spec**
Key | Desc
--- | ---
`tasks` | defines the list of tasks that can be used in commands
`tasks.type` | task type defines behavior and required fields, see Task Types section below for details
`tasks.name` | the unique identifier to reference this task in commands
`tasks.input` | the file containing the data that will be used as input to this task
`tasks.initdelay` | the delay, in milliseconds, to wait before starting the task (defaults to 0)
`tasks.delay` | the delay, in milliseconds, between each line being printed for `sysout` and `syserr` tasks (defaults to 0)<br>_Note: delay is an estimate (not high precision), for delays < 16ms output is batched to simulate expected throughput_
`tasks.repeat` | the number of times to repeat the task, or `forever` to repeat forever
`commands` | defines the possible commands to respond to
`commands.args` | the args exactly as they appear when passed to `simcli` to trigger this commands' tasks
`commands.tasks` | the list of tasks to execute for this command
`commands.rc` | the return / exit code to use after all tasks are complete (defaults to 0)
`defaultCommand` | the default command to execute if no commands are matched, see `commands` above for spec

#### Task Types

Type | Desc
--- | ---
`sysout` | will print the contents of `input` file to `sysout`
`syserr` | will print the contents of `input` file ot `syserr`
`file` | will copy the contents of `input` file to `outPath`
`hang` | will cause program to hang forever


### Execute
To run:
`simcli <args>`

Example:

```sh
$ simcli hello
this
is
from
hello.txt
```


## TODO
- Validate config file (ie: no tasks missing required fields, no commands referring to unknown tasks, etc.)
- Add task for accepting data via stdin, define end condition
- Learning mode - pass args to command and record stdout, stdin, return codes, etc. and create an appropriate config
- Add ability to create a .exe using a particular name
  - Add ENVVAR / SUBCMD for creating an .exe that has a particular name that matches real CLI
  - Option to embed tasks / commands into .exe?
- Allow CLI flags with Trigger like simcli SIMCLI_FLAGS [flags]
- Randomness and/or A/B testing - certain percentage of requests or certain amount of requests before then failure or vice versa
- Restructure code for extensibility
- add debug flag and debug logging
- ability to add delays by lines, something like: 1-30:100, 31-:2000, etc.
  - a single file can be used as part of multiple tasks - if tasks could specify line numbers this could be achieved with existing api
- adjust arg parsing to not require flags in specific order and allow wildcards
- add task 'touch' w/ permissions
- add to file task permissions
- add 'cmd' task to execute an external command (no longer mocking at this point?)
- add cli args to simcli for learning, configuring, etc. - args defined in `simcli.yaml` take precedence
- add 'learning mode' - to capture the output of a command
- add batch param to sysout/err - cannot override batch if less than 16ms
- add sysboth task for stdout and err
