package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: aux4-json <command> [options]\n")
		fmt.Fprintf(os.Stderr, "Commands: get, pretty, inline, index, group, collect, merge, page, count, exec, select\n")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "get":
		runGet(args)
	case "pretty":
		runPretty(args)
	case "inline":
		runInline(args)
	case "index":
		runIndex(args)
	case "group":
		runGroup(args)
	case "collect":
		runCollect(args)
	case "merge":
		runMerge(args)
	case "page":
		runPage(args)
	case "count":
		runCount(args)
	case "exec":
		runExec(args)
	case "select":
		runSelect(args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}
