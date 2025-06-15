package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type GitLogOptions struct {
	Pretty          string
	Date            string
	Color           string
	AbbrevCommit    bool
	Decorate        bool
	NoShowSignature bool
}

type FormatMap map[rune]map[rune]int

type PrettyFormatter struct {
	format FormatMap
	proc   func(string) (string, error)
	chr    rune
}

func (f *PrettyFormatter) Format(value string) ([]string, error) {
	isFmt := false
	command := ""
	var current rune = 0
	pieces := []string{}
	last := 0
	for idx, char := range value {
		if !isFmt {
			if char == f.chr {
				isFmt = true
			}
			continue
		}

		if current == 0 && char == f.chr {
			command = string(f.chr)
		} else {
			if current == 0 {
				next, ok := f.format[char]
				if !ok {
					return nil, fmt.Errorf("unrecognized option '%c'", char)
				}

				current = char
				command = command + string(char)

				if len(next) > 0 {
					continue
				}
			} else {
				_, ok := f.format[current][char]
				if !ok {
					return nil, fmt.Errorf("unrecognized param '%c' for option '%c'", char, current)
				}

				command = command + string(char)
			}
		}

		if idx > 1 {
			pieces = append(pieces, value[last:idx-len(command)])
		}

		if rune(command[0]) == f.chr {
			pieces = append(pieces, string(f.chr))
		} else {
			res, err := f.proc(command)
			if err != nil {
				return nil, err
			}
			pieces = append(pieces, res)
		}

		last = idx + 1
		isFmt = false
		current = 0
		command = ""
	}

	if isFmt {
		return nil, fmt.Errorf("unterminated format string")
	}

	if last < len(value) {
		pieces = append(pieces, value[last:])
	}

	return pieces, nil
}

func (f *PrettyFormatter) SprintFormat() string {
	if f.chr == rune(0) {
		return ""
	}

	newVal := map[string]map[string]int{}
	for k, v := range f.format {
		if v == nil {
			newVal[string(k)] = nil
		} else {
			newVal[string(k)] = map[string]int{}

			for k2, v2 := range v {
				newVal[string(k)][string(k2)] = v2
			}
		}
	}

	return fmt.Sprintf("%+v", newVal)
}

var logPrettyFormatter = PrettyFormatter{
	format: nil,
	proc:   nil,
	chr:    '%',
}

func FormatCommit(c *object.Commit, options *GitLogOptions) ([]string, error) {
	if c == nil || options == nil {
		return nil, fmt.Errorf("invalid nil parameters")
	}

	lines := []string{}

	commitHash := c.Hash.String()
	if options.AbbrevCommit {
		commitHash = commitHash[:7]
	}

	commitDate := c.Author.When.Format(time.UnixDate)
	if options.Date == "short" {
		commitDate = c.Author.When.Format("2006-01-02")
	}

	if strings.Index(options.Pretty, "format:") == 0 {
		formatted, err := logPrettyFormatter.Format(options.Pretty[:7])
		if err != nil {
			return nil, err
		}
		lines = append(lines, strings.Join(formatted, ""))
	} else {
		lines = append(lines, fmt.Sprintf("commit %s", commitHash))
		lines = append(lines, fmt.Sprintf("Author: %s <%s>", c.Author.Name, c.Author.Email))
		lines = append(lines, fmt.Sprintf("Date:   %s", commitDate))
		lines = append(lines, "")
		for line := range strings.Lines(c.Message) {
			lines = append(lines, fmt.Sprintf("    %s", strings.TrimSpace(line)))
		}
	}

	return lines, nil
}

func PrintLog(repoDir string, revRange string, options GitLogOptions) error {
	revA := ""
	revB := "HEAD"

	if revRange != "" {
		parts := strings.Split(revRange, "..")
		if len(parts) != 2 {
			return fmt.Errorf("Invalid revision range format. Expected 'commitA..commitB', got '%s'", revRange)
		}
		revA = parts[0]
		revB = parts[1]
	}

	abs, err := filepath.Abs(repoDir)
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
			return fmt.Errorf("Error resolving revision %s: %v", revA, err)
		}

		commitA, err = r.CommitObject(*hashA)
		if err != nil {
			return fmt.Errorf("Could not get commit object for %s: %v", hashA, err)
		}
	}

	hashB, err := r.ResolveRevision(plumbing.Revision(revB))
	if err != nil {
		return fmt.Errorf("Error resolving revision %s: %v", revB, err)
	}

	cIter, err := r.Log(&git.LogOptions{From: *hashB})
	if err != nil {
		return fmt.Errorf("Error getting commit log: %v\n", err)
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

		lines, err := FormatCommit(c, &options)
		if err != nil {
			return err
		}

		for _, line := range lines {
			fmt.Println(line)
		}

		return nil
	})

	if err != nil && err.Error() != "stop iteration" {
		return err
	}

	return nil
}
