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
	Format        FormatMap
	ProcessOption func(string, any) (string, error)
	OptionStart   rune
}

func (f *PrettyFormatter) ApplyFormat(value string, context any) ([]string, error) {
	isFmt := false
	option := ""
	var current rune = 0
	pieces := []string{}
	last := 0
	for idx, char := range value {
		if !isFmt {
			if char == f.OptionStart {
				isFmt = true
			}
			continue
		}

		if current == 0 && char == f.OptionStart {
			option = string(f.OptionStart)
		} else {
			if current == 0 {
				next, ok := f.Format[char]
				if !ok {
					return nil, fmt.Errorf("unrecognized format option '%c'", char)
				}

				current = char
				option = option + string(char)

				if len(next) > 0 {
					continue
				}
			} else {
				_, ok := f.Format[current][char]
				if !ok {
					return nil, fmt.Errorf("unrecognized param '%c' for format option '%c'", char, current)
				}

				option = option + string(char)
			}
		}

		if idx > 1 {
			pieces = append(pieces, value[last:idx-len(option)])
		}

		if rune(option[0]) == f.OptionStart {
			pieces = append(pieces, string(f.OptionStart))
		} else {
			res, err := f.ProcessOption(option, context)
			if err != nil {
				return nil, err
			}
			pieces = append(pieces, res)
		}

		last = idx + 1
		isFmt = false
		current = 0
		option = ""
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
	if f.OptionStart == rune(0) {
		return ""
	}

	newVal := map[string]map[string]int{}
	for k, v := range f.Format {
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
	Format: FormatMap{
		'H': nil, 'h': nil, 's': nil,
		'c': map[rune]int{'s': 0}},
	ProcessOption: func(chars string, context any) (string, error) {
		c, ok := context.(*object.Commit)
		if !ok {
			return "", fmt.Errorf("no commit")
		}

		switch chars {
		case "h":
			return c.Hash.String()[:7], nil
		case "H":
			return c.Hash.String(), nil
		case "s":
			return strings.TrimSpace(strings.SplitN(c.Message, "\n", 1)[0]), nil
		case "cs":
			return c.Author.When.Format("2006-01-02"), nil
		default:
			return "", fmt.Errorf("error")
		}
	},
	OptionStart: '%',
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
		formatted, err := logPrettyFormatter.ApplyFormat(options.Pretty[7:], c)
		if err != nil {
			return nil, err
		}
		lines = append(lines, strings.Join(formatted, ""))
	} else {
		lines = append(lines, fmt.Sprintf("commit %s", commitHash))
		lines = append(lines, fmt.Sprintf("\nAuthor: %s <%s>", c.Author.Name, c.Author.Email))
		lines = append(lines, fmt.Sprintf("\nDate:   %s", commitDate))
		lines = append(lines, "\n")
		cntLines := 0
		for line := range strings.Lines(c.Message) {
			lines = append(lines, fmt.Sprintf("\n    %s", strings.TrimSpace(line)))
			cntLines += 1
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
			fmt.Print(line)
		}

		return nil
	})

	if err != nil && err.Error() != "stop iteration" {
		return err
	}
	fmt.Println()

	return nil
}
