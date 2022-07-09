package main

import (
	"fmt"
	"os"
	"strings"
	"vsartor.com/gochs/blogger"
	"vsartor.com/gochs/log"
)

func main() {
	args, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("failed to parseArgs: %s\n", err.Error())
		os.Exit(1)
	}

	log.Info("<b>srcDir<r> set to <b>%s<r>", args.srcDir)
	log.Info("<b>dstDir<r> set to <b>%s<r>", args.dstDir)
	if args.prod {
		log.Info("<b>production flag<r> is set")
	}

	if args.debug {
		log.AllowDbg()
	}

	err = blogger.CompileSource(args.srcDir, args.dstDir, args.prod)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// parseArgs is a custom function responsible for parsing CLI arguments.
// It receives arguments as a slice instead of directly accessing os.Args
// to facilitate testing in the future. It returns a structure containing
// relevant pieces of information.
// The rationale for not using a library like "flag" is that the CLI for
// this application is far too straight-forward, so I preferred having
// explicit code for parsing in this function.
func parseArgs(args []string) (cliArgs, error) {
	dirCnt := 0 // counts how many dirs we've already parsed
	var dirs [2]string
	prod := false
	debug := false

	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			// branch for parsing flags

			switch arg {
			case "-p", "-prod":
				prod = true
			case "-d", "-debug":
				debug = true
			default:
				return cliArgs{}, fmt.Errorf("unknown flag: %s", arg)
			}
		} else {
			// branch for parsing args

			if dirCnt >= 2 {
				return cliArgs{}, fmt.Errorf("only expected 2 directories, found 3rd: %s", arg)
			}

			dirs[dirCnt] = arg
			dirCnt++
		}
	}

	if dirCnt < 2 {
		return cliArgs{}, fmt.Errorf("expected 2 directories, found: %d", dirCnt)
	}

	return cliArgs{
		srcDir: dirs[0],
		dstDir: dirs[1],
		prod:   prod,
		debug:  debug,
	}, nil
}

type cliArgs struct {
	srcDir, dstDir string // source and destination directories
	prod, debug    bool   // flags to indicate production build & debug mode
}
