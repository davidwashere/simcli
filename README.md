# SimCLI
Utility for simulating CLI response for testing, mocking, and more

## Usage


### Config
`SimCLI` requires a config file defining the commands to respond to and the responses to respond with

By default `simcli.yaml` is expected in the current working directory. To change this set a new path via the `SIMCLI_CONFIG` environment variable

```yaml
responses:
  - name: helloResponse
    input: data/hello.txt
  - name: progressResponse
    input: data/progress.txt
    delay: 300
  - name: errorResponse
    input: data/error.txt
    rc: 1
commands:
  - args: hello
    responseName: helloResponse
  - args: progress
    responseName: progressResponse
defaultResponse: errorResponse

```

### Execute
To run:
`simcli <args>`

Example:

```sh
$ simcli hello
hello
this
is
sample
output

```


## TODO
- [x] Add ability to sleep/delay output
- [x] Specify config via env Var
- [ ] Add validation step if any response is used where file is missing
- [ ] Add ability to pick specific lines 
- [ ] Add ability to skip lines that start with stuff
- [ ] Add tasks other than reading files, like making http requests
- [ ] Add task for accepting data via stdin, define end condition
- [ ] Learning mode - pass args to command and record stdout, stdin, return codes, etc. and create an appropriate config
- [ ] Create GUI for creating tasks / plans