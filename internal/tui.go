package internal

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-redis/redis/v8"
)

type state int

const (
	defaultState state = iota
	searchState
	exportState
)

//nolint:govet
type model struct {
	list      list.Model
	textinput textinput.Model

	rdb           redis.UniversalClient
	searchValue   string
	statusMessage string
	now           string

	viewport viewport.Model

	keyMap
	spinner spinner.Model

	width, height int

	offset int64
	limit  int64 // scan size

	state
	ready bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.tickCmd(), m.spinner.Tick, m.scanCmd(), m.countCmd())
}

// keyMap defines the keybindings for the app.
type keyMap struct {
	reload key.Binding
	search key.Binding
	export key.Binding
}

// defaultKeyMap returns a set of default keybindings.
func defaultKeyMap() keyMap {
	return keyMap{
		reload: key.NewBinding(
			key.WithKeys("r"),
		),
		search: key.NewBinding(
			key.WithKeys("s"),
		),
		export: key.NewBinding(
			key.WithKeys("e"),
		),
	}
}

func Or(queryValues url.Values, key string, b string) string {
	a := queryValues[key]
	if len(a) == 0 {
		return b
	}

	return a[0]
}

func OrInt(queryValues url.Values, key string, b int) int {
	a := queryValues[key]
	if len(a) == 0 {
		return b
	}

	value, err := strconv.Atoi(a[0])
	if err != nil {
		return b
	}

	return value
}

func OrSlice(queryValues url.Values, key string, bb ...[]string) []string {
	a := queryValues[key]
	if len(a) > 0 {
		return a
	}

	for _, b := range bb {
		if len(b) > 0 {
			return b
		}
	}

	return nil
}

func New(config Config) (*model, error) {
	// export REDIS=addr&db=0&pwd=1qazzaq1
	redisQuery, _ := url.ParseQuery(os.Getenv("REDIS"))

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        OrSlice(redisQuery, "addr", config.Addrs, []string{"127.0.0.1:6379"}),
		DB:           OrInt(redisQuery, "db", config.DB),
		Username:     Or(redisQuery, "user", config.Username),
		Password:     Or(redisQuery, "pwd", config.Password),
		MaxRetries:   OrInt(redisQuery, "maxRetries", MaxRetries),
		MaxRedirects: OrInt(redisQuery, "maxRedirects", MaxRedirects),
		MasterName:   Or(redisQuery, "masterName", config.MasterName),
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("connect to redis failed: %w", err)
	}

	t := textinput.New()
	t.Prompt = "> "
	t.Placeholder = "Search Key"
	t.PlaceholderStyle = lipgloss.NewStyle()

	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Redis Viewer"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)

	s := spinner.New()
	s.Spinner = spinner.Dot

	return &model{
		list:      l,
		textinput: t,
		spinner:   s,

		rdb: rdb,

		limit: config.Limit,

		keyMap: defaultKeyMap(),
		state:  defaultState,
	}, nil
}
