package typer

import (
	"bytes"
	"fmt"
	"fun/internal/storage"
	"log"
	"math"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TyperModel struct {
	parent tea.Model
	tmpl   *template.Template
	s      *storage.Storage

	target    []rune
	typed     []rune
	start     time.Time
	end       time.Time
	ch        chan struct{}
	correct   int
	wrong     int
	backspace int
}

func (tm *TyperModel) ViewScore() string {
	sum := tm.correct + tm.wrong
	if sum == 0 {
		return ""
	}

	accuracy := tm.correct * 100 / sum
	spent := tm.end.Sub(tm.start).Seconds()
	score := (tm.correct - tm.wrong - tm.backspace) / int(math.Ceil(spent))

	return fmt.Sprintf("correct: %d\nwrong: %d\naccuracy: %d%%\nscore: %d",
		tm.correct, tm.wrong, accuracy, score,
	)
}

func (tm *TyperModel) ViewTime() int {
	if tm.start.IsZero() {
		return int(typerTimerDuration.Seconds())
	}

	return int(math.Ceil(tm.start.Add(typerTimerDuration).Sub(time.Now()).Seconds()))
}

const (
	targetColor  = "7"
	correctColor = "2"
	wrongColor   = "1"
)

var (
	targetStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(targetColor))
	correctStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(correctColor))
	wrongStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(wrongColor))
	cursorStyle     = targetStyle.Underline(true)
	wrongSpaceStyle = cursorStyle.Foreground(lipgloss.Color(wrongColor))
)

func (tm *TyperModel) ViewText() (s string) {
	for i := 0; i < len(tm.typed); i++ {
		style := correctStyle
		if tm.target[i] != tm.typed[i] {
			if tm.typed[i] == ' ' {
				style = wrongSpaceStyle
			} else {
				style = wrongStyle
			}
		}
		s += style.Render(string(tm.typed[i]))
	}

	if len(tm.target) > len(tm.typed) {
		s += cursorStyle.Render(string(tm.target[len(tm.typed)]))
	}

	if len(tm.target)+1 > len(tm.typed) {
		s += targetStyle.Render(string(tm.target[len(tm.typed)+1:]))
	}

	return
}

const (
	typerTemplateName = "TyperTemplate"
	typerTemplate     = `
Time left: {{printf "%2d" .ViewTime}}

{{.ViewText}}
`
)

func NewTyperModel(parent tea.Model) (*TyperModel, error) {
	tmpl, err := template.New(typerTemplateName).Parse(typerTemplate)
	if err != nil {
		return nil, err
	}

	s, err := storage.NewStorage()
	if err != nil {
		return nil, err
	}

	return &TyperModel{
		parent: parent,
		tmpl:   tmpl,
		s:      s,
	}, nil
}

func (tm *TyperModel) Prepare() {
	text := tm.s.GetText()
	tm.target = []rune(text)
	tm.typed = []rune{}
	tm.start = time.Time{}
	tm.end = time.Time{}
	tm.ch = make(chan struct{}, 1)
	tm.correct = 0
	tm.wrong = 0
	tm.backspace = 0
}

func (tm *TyperModel) Init() tea.Cmd {
	return nil
}

const (
	typerTimerDuration  = time.Second * 30
	typerUpdateDuration = time.Second
)

type typerStop struct{}
type typerUpdate struct{}

func timerUpdateCmd() tea.Msg {
	time.Sleep(typerUpdateDuration)
	return typerUpdate{}
}

func (tm *TyperModel) timerDurationCmd() tea.Msg {
	for {
		select {
		case <-time.After(typerTimerDuration):
			return typerStop{}
		case <-tm.ch:
			return nil
		}
	}
}

func (tm *TyperModel) Update(msg tea.Msg) (m tea.Model, cmd tea.Cmd) {
	m = tm
	stop := func() {
		close(tm.ch)
		tm.end = time.Now()
		m = tm.parent
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			cmd = tea.Quit

		case tea.KeyCtrlQ:
			stop()

		case tea.KeyRunes, tea.KeySpace:
			if tm.start.IsZero() {
				tm.start = time.Now()
				cmd = tea.Batch(
					timerUpdateCmd,
					tm.timerDurationCmd,
				)
			}

			tm.typed = append(tm.typed, msg.Runes[0])
			if i := len(tm.typed) - 1; tm.target[i] == tm.typed[i] {
				tm.correct++
			} else {
				tm.wrong++
			}
			if len(tm.target) == len(tm.typed) {
				stop()
			}

		case tea.KeyBackspace:
			if len(tm.typed) > 0 {
				tm.typed = tm.typed[:len(tm.typed)-1]
				tm.backspace++
			}
		}

	case typerStop:
		stop()

	case typerUpdate:
		cmd = timerUpdateCmd
	}

	return
}

var viewStyle = lipgloss.NewStyle().Padding(1, 2)

func (tm *TyperModel) View() string {
	buf := new(bytes.Buffer)
	if err := tm.tmpl.Execute(buf, tm); err != nil {
		log.Fatalf("From TyperModel.View(): %s\n", err.Error())
	}

	return viewStyle.Render(buf.String())
}
