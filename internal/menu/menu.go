package menu

import (
	"bytes"
	"fun/internal/typer"
	"log"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MenuModel struct {
	tmpl *template.Template
	tm   *typer.TyperModel
}

func (mm *MenuModel) ViewScore() string {
	return mm.tm.ViewScore()
}

const (
	menuTemplateName = "MenuTemplate"
	menuTemplate     = `
{{.ViewScore}}

Press s to start
`
)

func NewMenuModel() (*MenuModel, error) {
	tmpl, err := template.New(menuTemplate).Parse(menuTemplate)

	if err != nil {
		return nil, err
	}

	mm := &MenuModel{
		tmpl: tmpl,
	}

	mm.tm, err = typer.NewTyperModel(mm)
	if err != nil {
		return nil, err
	}

	return mm, nil
}

func (mm *MenuModel) Init() tea.Cmd {
	return nil
}

func (mm *MenuModel) Update(msg tea.Msg) (m tea.Model, cmd tea.Cmd) {
	m = mm
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			cmd = tea.Quit

		case "s":
			mm.tm.Prepare()
			m = mm.tm
		}
	}

	return
}

var viewStyle = lipgloss.NewStyle().Padding(1, 2)

func (mm *MenuModel) View() string {
	buf := new(bytes.Buffer)
	if err := mm.tmpl.Execute(buf, mm); err != nil {
		log.Fatalf("From MenuModel.View(): %s\n", err.Error())
	}

	return viewStyle.Render(buf.String())
}
