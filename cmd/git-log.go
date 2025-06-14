package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/spf13/cobra"
)

var gitLogPretty string
var gitLogDate string
var gitLogColor string
var gitLogAbbrevCommit bool
var gitLogDecorate bool
var gitLogNoShowSignature bool

var gitLogCmd = &cobra.Command{
	Use: "log [<revision-range>]",
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("args", "a", args)

		revA := ""
		revB := "HEAD"

		var revRange string
		if len(args) > 0 {
			revRange = args[0]
			parts := strings.Split(revRange, "..")
			if len(parts) != 2 {
				return fmt.Errorf("Invalid revision range format. Expected 'commitA..commitB', got '%s'", revRange)
			}
			revA = parts[0]
			revB = parts[1]
		}

		abs, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		r, err := git.PlainOpen(abs)
		if err != nil {
			return err
		}

		var commitA *object.Commit
		var hashA *plumbing.Hash
		if revA != "" {
			hashA, err = r.ResolveRevision(plumbing.Revision(revA))
			if err != nil {
				log.Fatalf("Error resolving revision %s: %v", revA, err)
			}

			commitA, err = r.CommitObject(*hashA)
			if err != nil {
				log.Fatalf("Could not get commit object for %s: %v", hashA, err)
			}
		}

		hashB, err := r.ResolveRevision(plumbing.Revision(revB))
		if err != nil {
			log.Fatalf("Error resolving revision %s: %v", revB, err)
		}

		cIter, err := r.Log(&git.LogOptions{From: *hashB})
		if err != nil {
			log.Fatalf("Error getting commit log: %v\n", err)
		}

		printedFirst := false
		err = cIter.ForEach(func(c *object.Commit) error {
			if commitA != nil {
				if c.Hash == commitA.Hash {
					return fmt.Errorf("stop iteration")
				}
				isAncestor, err := c.IsAncestor(commitA)
				if err != nil {
					return err
				}
				if isAncestor {
					return fmt.Errorf("stop iteration")
				}
			}

			if printedFirst {
				fmt.Println()
			}
			printedFirst = true

			fmt.Println("commit", c.Hash)
			fmt.Println("Author:", c.Author.Name, "<"+c.Author.Email+">")
			fmt.Println("Date:  ", c.Author.When)
			fmt.Println()
			for line := range strings.Lines(c.Message) {
				fmt.Print("    ", line)
			}
			return nil
		})

		if err != nil && err.Error() != "stop iteration" {
			return err
		}

		return nil
	},
}

func init() {
	gitCmd.AddCommand(gitLogCmd)

	gitLogCmd.Flags().StringVar(
		&gitLogPretty,
		"pretty",
		"",
		"")

	gitLogCmd.Flags().StringVar(
		&gitLogDate,
		"date",
		"",
		"")

	gitLogCmd.Flags().StringVar(
		&gitLogColor,
		"color",
		"",
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogAbbrevCommit,
		"abbrev-commit",
		false,
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogDecorate,
		"decorate",
		false,
		"")

	gitLogCmd.Flags().BoolVar(
		&gitLogNoShowSignature,
		"no-show-signature",
		false,
		"")
}
