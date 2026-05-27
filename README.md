# rcc (retrospective code coverage + lines of code)

rcc can be used to create a graph/plot of test coverage and LoC data based on
a project's git commit history. Graph data is gathered by walking the
commit history, cloning each commit into a temporary directory and running
LoC and/or unit and/or integration tests on the cloned worktree.

It supports (and tries to auto-detect) few built-in programming languages
and their commonly used coverage tools. For other languages or different needs,
relevant commands to obtain coverage can be provided by the user. Same applies
to regex used to extract coverage value from test output - regexes documented
by [gitlab](https://docs.gitlab.com/ci/testing/code_coverage/coverage_reporting/#coverage-regex-patterns)
should generally work.

Currently [built-in support](languages.go):
- Go (`go test` for unit tests, `go test -tags=integration` for integration tests)
- Python (`pytest` / [pytest-cov](https://pypi.org/project/pytest-cov/))
- Java (Gradle / jacoco)

Dependencies for integration tests should possibly be started before running rcc.
Alternatively, built-in defaults can be overridden by specifying custom commands
for test execution.

By default, five parallel workers will be spawned to run analysis in parallel.
In theory, this should work well for unit tests. For integration tests, it's
more likely that rcc should be limited to a single worker (see usage below).

rcc supports multiple output formats for gathered data:
- HTML (embeds a bundled [chartjs](https://www.chartjs.org/), or links to it).
- PNG requires [gnuplot](http://www.gnuplot.info/) to be installed.
- JSON (using the chartjs dataset structure); raw data, no plot.
Output format is determined by the filename (extension) given for output, which
defaults to `rcc-output.html`.

## installation

No binary release available (yet?). Go required.

```bash
go install github.com/schnoddelbotz/rcc@main
```

## usage

First argument to `rcc` is a path to the project (under git version control) to be analyzed.
Without that argument, the current working directory will be analyzed.

Assuming a Go project in current working directory, to run LoC and unit test coverage, run
```bash
rcc
```

This will produce the default graph format (`rcc-output.html`).
Should the project language not be specified and auto-detection fail, `rcc` will only analyse LoC.

Overview of currently supported flags to influence `rcc` behaviour:

```bash
Flags:
  -I, --cover-integration              Run integration tests (for given --language)
  -X, --cover-integration-cmd string   Custom shell command for running integration tests
  -R, --cover-regex string             Custom regex to extract coverage value from test command output
  -C, --cover-unit-cmd string          Custom shell command for running unit tests
  -d, --debug                          Enable debug output
  -h, --help                           help for retrospective-code-coverage
  -J, --html-no-embed-chartjs          Do not embed ChartJS into generated .html, but link it
  -j, --html-no-embed-json             Do not embed JSON data into generated .html, but link it
  -i, --include-languages strings      Explicitly list languages for LoC
  -l, --language string                Enables details and coverage for given language
  -D, --no-cover-duration              Do not include duration for coverage runs in graph
  -U, --no-cover-unit                  Do not run unit tests (for given --language)
  -O, --open                           Open graph upon completion
  -o, --output string                  Plot/Graph html/json/png output filename (default "rcc-output.html")
  -s, --skip-autodetect                Disable language auto detection
  -t, --tmp string                     Temp directory path to use for history clones (default "/tmp")
  -w, --workers int                    Number of workers (default 5)
```

A more custom example - analyses the project at given path, opens the graph in PNG format upon
completion, also runs integration tests and only analyses (and graphs) LoC for given languages:
```bash
rcc -OIi Go,JavaScript,HTML -o my.png /path/to/my/project
```

## libraries used

- [go-git](https://github.com/go-git/go-git) to analyse git history
- [gocloc](https://github.com/hhatto/gocloc) to get LoC data
- [cobra](https://github.com/spf13/cobra) CLI command line interface

## status / todo

Fun project, WIP. Open tasks:

- [ ] add option to disable LoC
- [ ] add option to limit time git history --range (time/sha)
- [ ] add cli flag: output png dimensions
- [ ] drop cobra?
- [ ] add option to create/use ramdisk for tmp git clones
- [ ] improve test coverage m( ... and add graph as example here
- [ ] add JSON --append mode, to extend an exist json ouput file (/w 1 or more commits, based on range)

## license

MIT
