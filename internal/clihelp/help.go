package clihelp

import (
	_ "embed"
	"fmt"
	"io"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

//go:embed help.yaml
var embeddedHelp []byte

type Option struct {
	Flag        string `yaml:"flag"`
	Description string `yaml:"description"`
}

type Command struct {
	Path     string   `yaml:"path"`
	Usage    string   `yaml:"usage"`
	Summary  string   `yaml:"summary"`
	Aliases  []string `yaml:"aliases"`
	Examples []string `yaml:"examples"`
}

type Catalog struct {
	Program       string    `yaml:"program"`
	Intro         string    `yaml:"intro"`
	Usage         []string  `yaml:"usage"`
	GlobalOptions []Option  `yaml:"global_options"`
	Commands      []Command `yaml:"commands"`

	commandsByPath  map[string]Command
	aliasesToPath   map[string]string
	commandPathList []string
	globalBoolFlags map[string]struct{}
	globalValFlags  map[string]struct{}
	helpFlags       map[string]struct{}
}

func Load() (*Catalog, error) {
	return parseCatalogYAML(embeddedHelp)
}

func parseCatalogYAML(data []byte) (*Catalog, error) {
	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("invalid help yaml: %w", err)
	}
	if c.Program == "" {
		return nil, fmt.Errorf("invalid help yaml: missing program")
	}
	if len(c.Usage) == 0 {
		return nil, fmt.Errorf("invalid help yaml: missing usage")
	}
	if len(c.Commands) == 0 {
		return nil, fmt.Errorf("invalid help yaml: missing commands")
	}

	c.commandsByPath = make(map[string]Command, len(c.Commands))
	c.aliasesToPath = make(map[string]string)
	c.commandPathList = make([]string, 0, len(c.Commands))
	c.globalBoolFlags = map[string]struct{}{}
	c.globalValFlags = map[string]struct{}{}
	c.helpFlags = map[string]struct{}{}
	for _, option := range c.GlobalOptions {
		tokens := splitFlagTokens(option.Flag)
		isHelpOption := false
		for _, token := range tokens {
			if strings.Contains(token, "help") {
				isHelpOption = true
				break
			}
		}
		for _, token := range tokens {
			isValueFlag := strings.Contains(token, "=")
			base := token
			if isValueFlag {
				base = token[:strings.Index(token, "=")]
				c.globalValFlags[base] = struct{}{}
			} else {
				c.globalBoolFlags[token] = struct{}{}
			}
			if isHelpOption {
				c.helpFlags[base] = struct{}{}
			}
		}
	}
	for _, cmd := range c.Commands {
		path := normalizePath(cmd.Path)
		if path == "" {
			return nil, fmt.Errorf("invalid help yaml: command path is empty")
		}
		if cmd.Usage == "" {
			return nil, fmt.Errorf("invalid help yaml: missing usage for %q", path)
		}
		if _, ok := c.commandsByPath[path]; ok {
			return nil, fmt.Errorf("invalid help yaml: duplicate command path %q", path)
		}
		cmd.Path = path
		c.commandsByPath[path] = cmd
		c.commandPathList = append(c.commandPathList, path)
		for _, alias := range cmd.Aliases {
			normalizedAlias := normalizePath(alias)
			if normalizedAlias == "" {
				return nil, fmt.Errorf("invalid help yaml: empty alias for %q", path)
			}
			if existing, ok := c.aliasesToPath[normalizedAlias]; ok && existing != path {
				return nil, fmt.Errorf("invalid help yaml: duplicate alias %q", normalizedAlias)
			}
			c.aliasesToPath[normalizedAlias] = path
		}
	}
	sort.Slice(c.commandPathList, func(i, j int) bool {
		left := strings.Count(c.commandPathList[i], " ")
		right := strings.Count(c.commandPathList[j], " ")
		if left == right {
			return c.commandPathList[i] < c.commandPathList[j]
		}
		return left > right
	})
	return &c, nil
}

func (c *Catalog) ResolveHelp(args []string) (topic string, handled bool) {
	if len(args) == 0 {
		return "", true
	}
	tokens := c.stripGlobalArgs(args)
	if len(tokens) == 0 {
		return "", true
	}

	if tokens[0] == "help" {
		rest := tokens[1:]
		if len(rest) > 0 && rest[0] == "--" {
			rest = rest[1:]
		}
		path, hasTopic, _ := c.resolvePathFromTopic(rest)
		if !hasTopic {
			return "", true
		}
		return path, true
	}

	if !slices.ContainsFunc(tokens, c.isHelpFlag) {
		return "", false
	}
	pathTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if c.isHelpFlag(token) {
			continue
		}
		pathTokens = append(pathTokens, token)
	}
	path, hasTopic, _ := c.resolvePathFromArgs(pathTokens)
	if !hasTopic {
		return "", true
	}
	return path, true
}

func (c *Catalog) Render(version, topic string) (string, error) {
	if topic == "" {
		return c.renderGlobal(version), nil
	}
	command, ok := c.findCommand(topic)
	if !ok {
		return "", fmt.Errorf("unknown help topic %q\nRun 'ios --help' to list available commands.", topic)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n\n", c.Program, version)
	fmt.Fprintf(&b, "Command: %s\n", command.Path)
	if command.Summary != "" {
		fmt.Fprintf(&b, "%s\n\n", command.Summary)
	} else {
		b.WriteString("\n")
	}
	b.WriteString("Usage:\n")
	fmt.Fprintf(&b, "  %s\n\n", command.Usage)
	if len(command.Examples) > 0 {
		b.WriteString("Examples:\n")
		for _, example := range command.Examples {
			fmt.Fprintf(&b, "  %s\n", example)
		}
		b.WriteString("\n")
	}
	b.WriteString("Global options:\n")
	writeAlignedOptions(&b, c.GlobalOptions)
	return b.String(), nil
}

func (c *Catalog) WriteHelp(args []string, version string, stdout, stderr io.Writer) (handled bool, exitCode int) {
	topic, handled := c.ResolveHelp(args)
	if !handled {
		return false, 0
	}
	out, err := c.Render(version, topic)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		return true, 1
	}
	_, _ = io.WriteString(stdout, out)
	return true, 0
}

func (c *Catalog) renderGlobal(version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s\n\n", c.Program, version)
	if c.Intro != "" {
		b.WriteString(c.Intro)
		b.WriteString("\n\n")
	}
	b.WriteString("Usage:\n")
	for _, line := range c.Usage {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	b.WriteString("\nGlobal options:\n")
	writeAlignedOptions(&b, c.GlobalOptions)
	b.WriteString("\nCommands:\n")
	maxWidth := c.maxCommandPathWidth()
	for _, command := range c.sortedCommandsForDisplay() {
		summary := command.Summary
		if summary == "" {
			summary = "See command help."
		}
		writeAlignedPair(&b, command.Path, summary, maxWidth)
	}
	b.WriteString("\nRun 'ios help <command>' or 'ios <command> --help' for command details.\n")
	return b.String()
}

func (c *Catalog) sortedCommandsForDisplay() []Command {
	commands := make([]Command, len(c.Commands))
	copy(commands, c.Commands)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Path < commands[j].Path
	})
	return commands
}

func (c *Catalog) resolvePathFromTopic(tokens []string) (string, bool, bool) {
	if len(tokens) == 0 {
		return "", false, true
	}
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if strings.HasPrefix(token, "-") {
			continue
		}
		filtered = append(filtered, token)
	}
	if len(filtered) == 0 {
		return "", false, true
	}
	return c.resolvePathFromArgs(filtered)
}

func (c *Catalog) resolvePathFromArgs(tokens []string) (string, bool, bool) {
	candidateTokens := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if strings.HasPrefix(token, "-") {
			continue
		}
		candidateTokens = append(candidateTokens, token)
	}
	if len(candidateTokens) == 0 {
		return "", false, true
	}
	for _, path := range c.commandPathList {
		commandTokens := strings.Fields(path)
		if len(commandTokens) > len(candidateTokens) {
			continue
		}
		matched := true
		for i := range commandTokens {
			if commandTokens[i] != candidateTokens[i] {
				matched = false
				break
			}
		}
		if matched {
			return path, true, true
		}
	}
	return normalizePath(strings.Join(candidateTokens, " ")), true, false
}

func (c *Catalog) findCommand(path string) (Command, bool) {
	normalized := normalizePath(path)
	if command, ok := c.commandsByPath[normalized]; ok {
		return command, true
	}
	if target, ok := c.aliasesToPath[normalized]; ok {
		command, ok := c.commandsByPath[target]
		return command, ok
	}
	return Command{}, false
}

func normalizePath(path string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(path)), " ")
}

func splitFlagTokens(flag string) []string {
	parts := strings.Split(flag, ",")
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func (c *Catalog) isHelpFlag(token string) bool {
	_, ok := c.helpFlags[token]
	return ok
}

func (c *Catalog) stripGlobalArgs(args []string) []string {
	result := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if _, ok := c.globalBoolFlags[arg]; ok {
			if c.isHelpFlag(arg) {
				result = append(result, arg)
			}
			continue
		}
		if _, ok := c.globalValFlags[arg]; ok {
			if i+1 < len(args) {
				i++
			}
			continue
		}
		skipped := false
		for flag := range c.globalValFlags {
			if strings.HasPrefix(arg, flag+"=") {
				skipped = true
				break
			}
		}
		if skipped {
			continue
		}
		result = append(result, arg)
	}
	return result
}

func writeAlignedOptions(b *strings.Builder, options []Option) {
	maxWidth := 0
	for _, option := range options {
		width := utf8.RuneCountInString(option.Flag)
		if width > maxWidth {
			maxWidth = width
		}
	}
	for _, option := range options {
		writeAlignedPair(b, option.Flag, option.Description, maxWidth)
	}
}

func (c *Catalog) maxCommandPathWidth() int {
	maxWidth := 0
	for _, command := range c.Commands {
		width := utf8.RuneCountInString(command.Path)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

func writeAlignedPair(b *strings.Builder, left, right string, maxWidth int) {
	padding := max(maxWidth-utf8.RuneCountInString(left), 0)
	b.WriteString("  ")
	b.WriteString(left)
	b.WriteString(strings.Repeat(" ", padding+2))
	b.WriteString(right)
	b.WriteString("\n")
}
