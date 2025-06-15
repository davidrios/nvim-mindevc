package git

import (
	"fmt"
	"strings"
	"testing"
)

var processor1 = PrettyFormatter{
	format: FormatMap{'c': nil},
	proc: func(chars string) (string, error) {
		return "world", nil
	},
	chr: '%',
}

var processor2 = PrettyFormatter{
	format: FormatMap{'a': nil, 'b': nil},
	proc: func(chars string) (string, error) {
		switch chars {
		case "a":
			return "hello", nil
		case "b":
			return "olá", nil
		default:
			return fmt.Sprintf("<invalid:%s>", chars), nil
		}
	},
	chr: '%',
}

var processor3 = PrettyFormatter{
	format: FormatMap{'x': nil, 'b': map[rune]int{'r': 0, 'x': 0}},
	proc: func(chars string) (string, error) {
		switch chars {
		case "x":
			return "planet X", nil
		case "br":
			return "Brasil", nil
		case "bx":
			return "not BR", nil
		default:
			return fmt.Sprintf("<invalid:%s>", chars), nil
		}
	},
	chr: '%',
}

func TestPrettyFormat(t *testing.T) {
	testTable := []struct {
		formatter *PrettyFormatter
		input     string
		success   bool
		output    string
	}{
		{
			formatter: &processor1,
			input:     "hello %c",
			success:   true,
			output:    "hello world",
		},
		{
			formatter: &processor1,
			input:     "hello %c.",
			success:   true,
			output:    "hello world.",
		},
		{
			formatter: &processor1,
			input:     "hell%%co",
			success:   true,
			output:    "hell%co",
		},
		{
			formatter: &processor1,
			input:     "hell%%",
			success:   true,
			output:    "hell%",
		},
		{
			formatter: &processor1,
			input:     "%chello",
			success:   true,
			output:    "worldhello",
		},
		{
			formatter: &processor1,
			input:     "%%chello",
			success:   true,
			output:    "%chello",
		},
		{
			formatter: &processor1,
			input:     "%bhello",
			success:   false,
			output:    "unrecognized option 'b'",
		},
		{
			formatter: &processor1,
			input:     "%%bhello",
			success:   true,
			output:    "%bhello",
		},
		{
			formatter: &processor2,
			input:     "",
			success:   true,
			output:    "",
		},
		{
			formatter: &processor2,
			input:     "a",
			success:   true,
			output:    "a",
		},
		{
			formatter: &processor2,
			input:     "%a world, %b mundo",
			success:   true,
			output:    "hello world, olá mundo",
		},
		{
			formatter: &processor2,
			input:     "x%b mundo. world, %a",
			success:   true,
			output:    "xolá mundo. world, hello",
		},
		{
			formatter: &processor2,
			input:     "x%%b mundo. world, %a",
			success:   true,
			output:    "x%b mundo. world, hello",
		},
		{
			formatter: &processor2,
			input:     "x%%g mundo. world, %d",
			success:   false,
			output:    "unrecognized option 'd'",
		},
		{
			formatter: &processor3,
			input:     "%x is nice, but %br is better. If %bx, I don't know",
			success:   true,
			output:    "planet X is nice, but Brasil is better. If not BR, I don't know",
		},
		{
			formatter: &processor3,
			input:     "%x is nice, but %bz is better. If %bx, I don't know",
			success:   false,
			output:    "unrecognized param 'z' for option 'b'",
		},
		{
			formatter: &processor3,
			input:     "%x is nice, but %b% is better",
			success:   false,
			output:    "unrecognized param '%' for option 'b'",
		},
	}

	for _, tv := range testTable {
		t.Run(fmt.Sprintf("%+v:%s", tv.formatter.SprintFormat(), tv.input), func(t *testing.T) {
			res, err := tv.formatter.Format(tv.input)
			if tv.success {
				if err != nil {
					t.Fatalf("unexpected error '%s'", err)
				}
				output := strings.Join(res, "")
				if output != tv.output {
					t.Fatalf("unexpected output '%s', expecting '%s'", output, tv.output)
				}
			} else {
				if err == nil {
					t.Fatalf("expecting error")
				}
				if err.Error() != tv.output {
					t.Fatalf("unexpected error '%s'", err.Error())
				}
			}
		})
	}
}
