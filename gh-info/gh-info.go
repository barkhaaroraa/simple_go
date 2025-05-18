package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// -----------------------------------------------------------------------------
// GitHub API -------------------------------------------------------------------

type user struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Bio       string `json:"bio"`
	Followers int    `json:"followers"`
	PublicRepos int  `json:"public_repos"`
	HTMLURL   string `json:"html_url"`
	AvatarURL string `json:"avatar_url"`
}

func fetchUser(username string) (user, error) {
	var u user
	url := "https://api.github.com/users/" + username
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "go-ghinfo-client")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return u, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&u)
	return u, err
}

// -----------------------------------------------------------------------------
// Bubble Tea model -------------------------------------------------------------

type model struct {
	username string
	loading  bool
	u        user
	err      error
}

func initialModel() model {
	return model{username: "barkhaaroraa", loading: true}
}

// ----- messages -----
type tickMsg time.Time
type userMsg struct {
	d   user
	err error
}

// ----- init -----
func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), getUser(m.username))
}

// fetch cmd
func getUser(name string) tea.Cmd {
	return func() tea.Msg {
		u, err := fetchUser(name)
		return userMsg{u, err}
	}
}

// ticker so we can show a spinner or similar later
func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// ----- update -----
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case userMsg:
		m.loading = false
		m.u, m.err = msg.d, msg.err
		return m, nil

	case tickMsg:
		if m.loading {
			return m, tick()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.loading = true
			return m, getUser(m.username)
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// ----- view -----
func (m model) View() string {
	if m.loading {
		return " "
	}
	if m.err != nil {
		return fmt.Sprintf("\n  Error: %v\n", m.err)
	}

	return fmt.Sprintf(`
  ****************************
  %s (%s)

  Bio:  %s
  Repos: %d   Followers: %d

  Profile: %s

  [r] refresh   [q] quit
  *******************************
`,
		m.u.Name, m.u.Login,
		empty(m.u.Bio, "—"),
		m.u.PublicRepos, m.u.Followers,
		m.u.HTMLURL,
	)
}

func empty(s, repl string) string {
	if s == "" {
		return repl
	}
	return s
}

// -----------------------------------------------------------------------------
// main ------------------------------------------------------------------------
func main() {
	if err := tea.NewProgram(initialModel()).Start(); err != nil {
		fmt.Println("Error:", err)
	}
}
