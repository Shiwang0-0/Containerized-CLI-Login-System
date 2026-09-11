// internals/prompt/prompt.go
package prompt

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

var ErrInterrupted = errors.New("interrupted")

type Prompt struct {
	rl *readline.Instance
}

var preLoginCommands = []string{"register", "login", "help", "exit"}
var postLoginCommands = []string{"whoami", "enable-2fa", "disable-2fa", "logout", "help", "exit"}

func New() (*Prompt, error) {
	historyPath := historyFilePath()

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     historyPath,
		AutoComplete:    completerFor(preLoginCommands),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}
	return &Prompt{rl: rl}, nil
}

func historyFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cli_history"
	}
	dir := filepath.Join(home, ".clilogin")
	_ = os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "history")
}

func completerFor(commands []string) readline.AutoCompleter {
	items := make([]readline.PrefixCompleterInterface, len(commands))
	for i, c := range commands {
		items[i] = readline.PcItem(c)
	}
	return readline.NewPrefixCompleter(items...)
}

// SetPreLoginMode switches tab-completion to the logged-out command set.
func (p *Prompt) SetPreLoginMode() {
	p.rl.Config.AutoComplete = completerFor(preLoginCommands)
}

// SetPostLoginMode switches tab-completion to the logged-in command set.
func (p *Prompt) SetPostLoginMode() {
	p.rl.Config.AutoComplete = completerFor(postLoginCommands)
}

// ReadLine shows the given prompt, returns trimmed input.
// Handles Ctrl+C / Ctrl+D distinctly from a real error.
func (p *Prompt) ReadLine(promptText string) (string, error) {
	p.rl.SetPrompt(promptText)
	line, err := p.rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			return "", ErrInterrupted
		}
		return "", err // io.EOF on Ctrl+D, or real error
	}
	return strings.TrimSpace(line), nil
}

// ReadPassword reads masked input (no echo), still line-buffered.
// Not added to history.
func (p *Prompt) ReadPassword(promptText string) (string, error) {
	pw, err := p.rl.ReadPassword(promptText)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(pw)), nil
}

func (p *Prompt) Close() error {
	return p.rl.Close()
}
