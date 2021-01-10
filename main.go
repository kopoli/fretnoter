package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kopoli/appkit"
)

var (
	version     = "Undefined"
	timestamp   = "Undefined"
	buildGOOS   = "Undefined"
	buildGOARCH = "Undefined"
	progVersion = "" + version
)

func fault(err error, message string, arg ...string) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s%s: %s\n", message, strings.Join(arg, " "), err)
		os.Exit(1)
	}
}

func main() {
	opts := appkit.NewOptions()

	opts.Set("program-name", os.Args[0])
	opts.Set("program-version", progVersion)
	opts.Set("program-timestamp", timestamp)
	opts.Set("program-buildgoos", buildGOOS)
	opts.Set("program-buildgoarch", buildGOARCH)

	base := appkit.NewCommand(nil, "", "Display fretboard notes")
	optVersion := base.Flags.Bool("version", false, "Display version")

	_ = appkit.NewCommand(base, "gui", "Start gui (default)")
	_ = appkit.NewCommand(base, "query q", "Query music database")

	err := base.Parse(os.Args[1:], opts)
	if err == flag.ErrHelp {
		os.Exit(0)
	}
	fault(err, "Parsing command line failed")

	if *optVersion {
		fmt.Println(appkit.VersionString(opts))
		os.Exit(0)
	}

	cmd := opts.Get("cmdline-command", "")
	args := appkit.SplitArguments(opts.Get("cmdline-args", ""))

	switch cmd {
	case "gui", "":
		err = GUIMain(progVersion)
		fault(err, "Running GUI failed")
	case "query":
		for i := range args {
			ch, err := DetectChord(args[i])
			fault(err, "Detecting chord failed")

			if len(ch) > 0 {
				fmt.Printf("%s: %s\n", args[i], strings.Join(ch, ", "))
			}
		}
	}

	os.Exit(0)
}
