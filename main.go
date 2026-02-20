package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dlclark/regexp2"
	"github.com/itchyny/gojq"
)

var version string

const (
	colorReset   = "\033[0m"
	colorBlue    = "\033[1;34m"
	colorYellow  = "\033[1;33m"
	colorRed     = "\033[1;31m"
	colorMagenta = "\033[1;35m"
)

func levelColor(level string) string {
	switch strings.ToUpper(level) {
	case "WARN", "WARNING":
		return colorYellow
	case "ERROR":
		return colorRed
	case "FATAL":
		return colorMagenta
	default:
		return ""
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type options struct {
	format     string
	level      string
	grep       string
	ignoreCase bool
}

func run() error {
	opts := options{}
	flag.StringVar(&opts.format, "format", "", "jq expression holding with the output format")
	flag.StringVar(&opts.level, "level", "", "jq expression to find the log level")
	flag.StringVar(&opts.grep, "grep", "", "PCRE pattern to highlight matches in blue")
	flag.BoolVar(&opts.ignoreCase, "i", false, "case-insensitive matching for --grep")
	flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if f := flag.Lookup("version"); f != nil && f.Value.String() == "true" {
		fmt.Println(version)
		return nil
	}

	var reader io.Reader = os.Stdin

	if args := flag.Args(); len(args) > 0 {
		file, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		reader = file
	}

	return processLines(reader, opts)
}

func highlightMatches(output string, re *regexp2.Regexp, baseColor string) string {
	var result strings.Builder
	pos := 0
	m, _ := re.FindStringMatch(output)
	for m != nil {
		result.WriteString(output[pos:m.Index])
		result.WriteString(colorBlue)
		result.WriteString(m.String())
		result.WriteString(colorReset)
		if baseColor != "" {
			result.WriteString(baseColor)
		}
		pos = m.Index + m.Length
		m, _ = re.FindNextMatch(m)
	}
	result.WriteString(output[pos:])
	return result.String()
}

func processLines(reader io.Reader, opts options) error {
	scanner := bufio.NewScanner(reader)

	if opts.format == "" {
		opts.format = ".message"
	}

	if opts.level == "" {
		opts.level = "(.level_name // .level)"
	}

	formatQuery, err := gojq.Parse(opts.format)
	if err != nil {
		return fmt.Errorf("failed to parse format query: %w", err)
	}

	var levelQuery *gojq.Query
	if opts.level != "" {
		levelQuery, err = gojq.Parse(opts.level)
		if err != nil {
			return fmt.Errorf("failed to parse level query: %w", err)
		}
	}

	var grepRe *regexp2.Regexp
	if opts.grep != "" {
		reFlags := regexp2.None
		if opts.ignoreCase {
			reFlags |= regexp2.IgnoreCase
		}
		grepRe, err = regexp2.Compile(opts.grep, reFlags)
		if err != nil {
			return fmt.Errorf("failed to compile grep pattern: %w", err)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		var data any
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			fmt.Println(line)
			continue
		}

		color := ""
		if levelQuery != nil {
			iter := levelQuery.Run(data)
			if result, ok := iter.Next(); ok {
				if str, ok := result.(string); ok {
					color = levelColor(str)
				}
			}
		}

		iter := formatQuery.Run(data)
		result, ok := iter.Next()

		if !ok {
			fmt.Println(line)
			continue
		}

		if _, isErr := result.(error); isErr {
			fmt.Println(line)
			continue
		}

		var output string
		if str, ok := result.(string); ok {
			output = str
		} else if result != nil {
			b, err := json.Marshal(result)
			if err != nil {
				fmt.Println(line)
				continue
			}
			output = string(b)
		} else {
			output = line
		}

		if grepRe != nil {
			output = highlightMatches(output, grepRe, color)
		}

		if color != "" {
			fmt.Printf("%s%s%s\n", color, output, colorReset)
		} else {
			fmt.Println(output)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
