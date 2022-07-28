# SimCLI
Utility for simulating CLI responses for testing, mocking, and more

## Usage

### Config
`SimCLI` requires a config file defining the tasks to execute when specific args are suppied to `simcli`

By default `simcli.yaml` is expected in the current working directory. To change this set a new path via the `SIMCLI_CONFIG` environment variable

Refer to [simcli.yaml](simcli.yaml) for an example config

**Config Spec**
Key | Desc
--- | ---
`tasks` | defines the tasks that can be used in commands
`tasks.type` | task type defines behavior and required fields, see Task Types section below for details
`tasks.name` | the unique identifier to reference this task in commands
`tasks.input` | the file containing the data that will be used as input to this task
`tasks.delay` | the delay in milliseconds between each line printed of input (defaults to 0)
`tasks.repeat` | the number of times to repeat the task, or `forever` to repeat forever
`commands` | defines the possible commands to respond to
`commands.args` | the args exactly as they appear when passed to `simcli` to trigger this commands tasks
`commands.tasks` | the tasks to execute for this command
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
- [ ] Validate config file (ie: no tasks missing required fields, no commands referring to unknown tasks, etc.)
- [ ] Add task for accepting data via stdin, define end condition
- [ ] Learning mode - pass args to command and record stdout, stdin, return codes, etc. and create an appropriate config
- [ ] Create GUI for creating tasks / plans
- [ ] Add ability to create a .exe using a particular name
  - [ ] Add ENVVAR / SUBCMD for creating an .exe that has a particular name that matches real CLI
- [ ] Allow CLI flags with Trigger like simcli SIMCLI_FLAGS [flags]
- [ ] Add condition to repeat a task x # of times or 'forever' ie: `repeat: 3` or `repeat: forever`