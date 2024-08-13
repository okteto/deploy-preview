package main

import (
	"errors"
	"fmt"
	"github.com/hoshsadiq/godotenv"
	flag "github.com/spf13/pflag"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type sliceString []string

func (i *sliceString) String() string {
	if i == nil {
		return ""
	}
	return strings.Join(*i, ";\n") + ";\n"
}

func (i *sliceString) Set(value string) error {
	vars, err := godotenv.Unmarshal(value)
	if err != nil {
		return err
	}

	for k, v := range vars {
		*i = append(*i, fmt.Sprintf("%s=%s", k, v))
	}
	return nil
}

func (i *sliceString) Type() string {
	return "var"
}

type DeployOptions struct {
	branch    string
	file      string
	logLevel  string
	name      string
	scope     string
	timeout   time.Duration
	variables sliceString
	wait      bool
	comment   string

	ci ciInfo
}

func exitIfErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

/*
todo: the follow code is left:

	if [ ! -z "$OKTETO_CA_CERT" ]; then
	   echo "Custom certificate is provided"
	   echo "$OKTETO_CA_CERT" > /usr/local/share/ca-certificates/okteto_ca_cert.crt
	   update-ca-certificates
	fi
*/
func main() {
	ci, err := getCIInfo()
	exitIfErr(err)

	opts := DeployOptions{
		ci: ci,
	}
	flagSet := flag.NewFlagSet("deploy-preview", flag.ContinueOnError)
	flagSet.StringVar(&opts.branch, "branch", ci.DefaultBranch(), "the branch to deploy (defaults to the current branch)")
	flagSet.StringVar(&opts.scope, "scope", "global", "the scope of preview environment to create. Accepted values are ['personal', 'global']")
	flagSet.StringVar(&opts.logLevel, "log-level", getLogLevel("warn"), "amount of information output (debug, info, warn, error)")
	flagSet.DurationVar(&opts.timeout, "timeout", 5*time.Minute, "the length of time to wait for completion, zero means never. Any other values should contain a corresponding time unit e.g. 1s, 2m, 3h ")
	flagSet.Var(&opts.variables, "var", "set a preview environment variable, this will be parsed as an env file, but can be set more than once")
	flagSet.BoolVar(&opts.wait, "wait", false, "wait until the preview environment deployment finishes (defaults to false)")
	flagSet.StringVar(&opts.file, "file", "", "relative path within the repository to the okteto manifest (default to okteto.yaml or .okteto/okteto.yaml)")
	flagSet.StringVar(&opts.comment, "comment", "", "Specify custom comment. Prefix with @ to read from a file")
	err = flagSet.Parse(os.Args)

	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// pflag already takes care of printing the usage.
			os.Exit(0)
		}

		exitIfErr(err)
	}

	err = validateInput(flagSet, &opts)
	exitIfErr(err)

	err = deployPreview(opts)
	if err != nil {
		log.Printf("deploy failed due to: %s", err)
	}
	var success = err == nil

	message, err := generateMessage(opts.name, success, opts.comment)
	exitIfErr(err)

	err = notify(ci, message)
	exitIfErr(err)

	if !success {
		os.Exit(1)
	}
}

func notify(ci ciInfo, message string) error {
	if ci != nil {
		return ci.Notify(message)
	}

	log.Printf("Not notifying anything, CI not supported")
	return nil
}

func getCIInfo() (ciInfo, error) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		return newGitHub()
	}

	return nil, errors.New("unsupported CI environment")
}

func getLogLevel(def string) string {
	// https://docs.github.com/en/actions/monitoring-and-troubleshooting-workflows/enabling-debug-logging
	// https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
	if os.Getenv("RUNNER_DEBUG") == "1" {
		return "debug"
	}

	return def
}

func validateInput(flagSet *flag.FlagSet, opts *DeployOptions) error {
	if flagSet.NArg() != 2 {
		return errors.New("preview environment name is required")
	}

	opts.branch = opts.ci.DefaultBranch()
	if opts.branch == "" {
		// this essentially means that retrieveDefaultBranch was unable to find a value
		return errors.New("failed to detect branch")
	}

	opts.name = flagSet.Arg(1)
	return nil
}

func deployPreview(opts DeployOptions) error {
	args := []string{"preview", "deploy", opts.name}
	args = append(args, fmt.Sprintf("--scope=%s", opts.scope))
	args = append(args, fmt.Sprintf("--branch=%s", opts.branch))
	args = append(args, fmt.Sprintf("--repository=%s", opts.ci.RepositoryURL()))
	args = append(args, fmt.Sprintf("--sourceUrl=%s", opts.ci.SourceURL()))

	if opts.timeout > 0 {
		args = append(args, fmt.Sprintf("--timeout=%s", opts.timeout.String()))
	}

	if opts.file != "" {
		args = append(args, fmt.Sprintf("--file=%s", opts.file))
	}

	if logLevel := getLogLevel(opts.logLevel); logLevel != "" {
		args = append(args, fmt.Sprintf("--log-level=%s", logLevel))
	}

	for _, variable := range opts.variables {
		args = append(args, fmt.Sprintf("--var=%s", variable))
	}

	args = append(args, "--wait")

	cmd := exec.Command("okteto", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "OKTETO_DISABLE_SPINNER=1")

	log.Printf("running: okteto %s", strings.Join(args, " "))

	return cmd.Run()
}
