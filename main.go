package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strings"
	"path/filepath"
)

func main() {
	app := cli.NewApp()

	app.Name = "rn"
	app.Usage = "cut substr from file name recursively"

	app.Flags = []cli.Flag{
		cli.BoolTFlag{
			Name:  "verbose, ver",
			Usage: "print affected filenames, true by default",
		},
		cli.BoolFlag{
			Name:  "dry-run, d",
			Usage: "display affected filenames without making any changes",
		},
	}

	app.Action = cli.ActionFunc(DefaultAction)

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "App exited with error: %v\n", err)
		os.Exit(1)
	}
}

func DefaultAction(ctx *cli.Context) error {
	if ctx.NArg() < 1 {
		return fmt.Errorf("Usage: rn [text to cut] [text to replace]")
	}
	toRemove := ctx.Args()

	replacerPairs := make([]string, 0, len(toRemove)*2)
	for _, str := range toRemove {
		replacerPairs = append(replacerPairs, str, "")
	}
	replacer := strings.NewReplacer(replacerPairs...)

	wd, _ := os.Getwd()

	files, err := ioutil.ReadDir(wd)
	if err != nil {
		return err
	}

	verbose := ctx.GlobalBoolT("verbose")
	dryRun := ctx.GlobalBoolT("dry-run")

	var do func(files []os.FileInfo, wd string)
	do = func(files []os.FileInfo, wd string) {
		for _, f := range files {
			fPath := filepath.Join(wd, f.Name())

			if f.IsDir() {
				subDirFiles, err := ioutil.ReadDir(fPath)
				if err != nil && verbose {
					fmt.Fprintf(os.Stderr, "Failed to read dir %s: %v\n", fPath, err)
					continue
				}
				do(subDirFiles, fPath)
			}

			newFileName := replacer.Replace(f.Name())
			if f.Name() == newFileName {
				continue
			}
			newFileName = strings.TrimSpace(newFileName)
			newFPath := filepath.Join(wd, newFileName)

			if verbose || dryRun {
				fmt.Println(f.Name(), "->", newFileName)
				if dryRun {
					continue
				}
			}

			if err = os.Rename(fPath, newFPath); err != nil && verbose {
				fmt.Fprintf(os.Stderr, "Failed to rename %s: %v\n", f.Name(), err)
				continue
			}
		}
	}
	do(files, wd)

	return nil
}
