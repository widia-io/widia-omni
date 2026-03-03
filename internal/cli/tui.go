package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appMode int

type appSection int

const (
	modeLogin appMode = iota
	modeApp
	modeTaskForm
	modeAreaForm
)

const (
	sectionDashboard appSection = iota
	sectionTasks
	sectionAreas
	sectionWorkspaces
	sectionSession
)

var sectionNames = []string{
	"Painel",
	"Tarefas",
	"Áreas",
	"Espaços de trabalho",
	"Sessão",
}

var taskPriorities = []string{"low", "medium", "high", "critical"}
var taskPriorityLabels = []string{"Baixa", "Média", "Alta", "Crítica"}

const (
	taskFocusTitle = iota
	taskFocusDescription
	taskFocusDueDate
	taskFocusPriority
	taskFocusArea
	taskFocusProject
	taskFocusDuration
	taskFocusLabels
	taskFocusCount = 8
)

var (
	clrTextPrimary   = lipgloss.Color("252")
	clrTextMuted     = lipgloss.Color("245")
	clrBorder        = lipgloss.Color("238")
	clrPanel         = lipgloss.Color("236")
	clrPanelSelected = lipgloss.Color("238")
	clrAccent        = lipgloss.Color("179")
	clrError         = lipgloss.Color("204")
	clrInfo          = lipgloss.Color("115")
)

var (
	rootStyle = lipgloss.NewStyle().Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(clrTextPrimary).
			Bold(true).
			Padding(0, 1)

	subHeaderStyle = lipgloss.NewStyle().
			Foreground(clrTextMuted).
			Padding(0, 1)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrBorder).
			Background(clrPanel).
			Padding(1, 1)

	mainStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrBorder).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(clrTextMuted).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(clrError).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(clrInfo).
			Padding(0, 1)
)

type userStateMsg struct {
	profile   *UserProfile
	workspace *Workspace
	err       error
}

type usageStateMsg struct {
	usage *WorkspaceUsage
	err   error
}

type tasksStateMsg struct {
	tasks     []Task
	completed bool
	err       error
}

type areasStateMsg struct {
	areas []Area
	err   error
}

type projectsStateMsg struct {
	projects []Project
	err      error
}

type labelsStateMsg struct {
	labels []Label
	err    error
}

type workspacesStateMsg struct {
	workspaces []WorkspaceListItem
	err        error
}

type loginResultMsg struct {
	err error
}

type logoutResultMsg struct {
	err error
}

type createTaskResultMsg struct {
	task *Task
	err  error
}

type createAreaResultMsg struct {
	area *Area
	err  error
}

type taskActionResultMsg struct {
	task   *Task
	reopen bool
	undo   bool
	err    error
}

type workspaceSwitchResultMsg struct {
	err error
}

type taskAction struct {
	taskID string
	reopen bool
}

type tuiModel struct {
	ctx    context.Context
	client *Client

	mode           appMode
	section        appSection
	sidebarCursor  int
	contentCursor  int
	focusSidebar   bool
	tasksCompleted bool
	loading        bool

	width  int
	height int

	profile    *UserProfile
	workspace  *Workspace
	usage      *WorkspaceUsage
	tasks      []Task
	areas      []Area
	workspaces []WorkspaceListItem
	projects   []Project
	labels     []Label

	loginEmail    textinput.Model
	loginPassword textinput.Model
	loginFocus    int

	taskTitle          textinput.Model
	taskDesc           textinput.Model
	taskDueDate        textinput.Model
	taskDuration       textinput.Model
	taskPriority       int
	taskAreaIdx        int
	taskProjectIdx     int
	taskLabelCursor    int
	taskSelectedLabels map[string]bool
	taskFocus          int

	areaName  textinput.Model
	areaSlug  textinput.Model
	areaFocus int

	lastTaskAction      *taskAction
	lastTaskActionLabel string

	errorText string
	infoText  string
}

func RunTUI(ctx context.Context, client *Client) error {
	m := newTUIModel(ctx, client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newTUIModel(ctx context.Context, client *Client) tuiModel {
	email := textinput.New()
	email.Placeholder = "e-mail"
	email.Prompt = "E-mail: "
	email.CharLimit = 120
	email.Width = 42

	password := textinput.New()
	password.Placeholder = "senha"
	password.Prompt = "Senha: "
	password.CharLimit = 120
	password.Width = 42
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '*'

	title := textinput.New()
	title.Placeholder = "Título da tarefa"
	title.Prompt = "Título: "
	title.CharLimit = 180
	title.Width = 60

	desc := textinput.New()
	desc.Placeholder = "Descrição (opcional)"
	desc.Prompt = "Descrição: "
	desc.CharLimit = 300
	desc.Width = 60

	dueDate := textinput.New()
	dueDate.Placeholder = "AAAA-MM-DD (opcional)"
	dueDate.Prompt = "Data: "
	dueDate.CharLimit = 16
	dueDate.Width = 60

	duration := textinput.New()
	duration.Placeholder = "ex: 30, 45m, 1h30m"
	duration.Prompt = "Duração (min): "
	duration.CharLimit = 16
	duration.Width = 60

	areaName := textinput.New()
	areaName.Placeholder = "Nome da área"
	areaName.Prompt = "Nome: "
	areaName.CharLimit = 120
	areaName.Width = 60

	areaSlug := textinput.New()
	areaSlug.Placeholder = "slug-da-área"
	areaSlug.Prompt = "Slug: "
	areaSlug.CharLimit = 120
	areaSlug.Width = 60

	mode := modeLogin
	if client.session.IsAuthenticated() {
		mode = modeApp
	}

	m := tuiModel{
		ctx:                ctx,
		client:             client,
		mode:               mode,
		section:            sectionDashboard,
		sidebarCursor:      0,
		focusSidebar:       true,
		tasksCompleted:     false,
		loginEmail:         email,
		loginPassword:      password,
		taskTitle:          title,
		taskDesc:           desc,
		taskDueDate:        dueDate,
		taskDuration:       duration,
		taskPriority:       1,
		taskAreaIdx:        -1,
		taskProjectIdx:     -1,
		taskLabelCursor:    0,
		taskSelectedLabels: map[string]bool{},
		areaName:           areaName,
		areaSlug:           areaSlug,
	}
	if mode == modeApp {
		m.loading = true
	}
	if m.mode == modeLogin {
		m.loginEmail.Focus()
	}
	return m
}

func (m tuiModel) Init() tea.Cmd {
	if m.mode == modeLogin {
		return textinput.Blink
	}
	m.loading = true
	return tea.Batch(
		loadUserStateCmd(m.ctx, m.client),
		loadUsageStateCmd(m.ctx, m.client),
	)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case userStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.profile = msg.profile
		m.workspace = msg.workspace
		if m.mode != modeApp {
			m.mode = modeApp
		}
		return m, nil
	case usageStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.usage = msg.usage
		return m, nil
	case tasksStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.tasks = msg.tasks
		m.tasksCompleted = msg.completed
		m.contentCursor = clampCursor(m.contentCursor, len(m.tasks))
		return m, nil
	case areasStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.areas = msg.areas
		m.contentCursor = clampCursor(m.contentCursor, len(m.areas))
		m.ensureTaskFormAreaDefaults()
		return m, nil
	case projectsStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.projects = msg.projects
		m.ensureTaskFormProjectDefaults()
		return m, nil
	case labelsStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.labels = msg.labels
		m.ensureTaskFormLabelCursor()
		return m, nil
	case workspacesStateMsg:
		m.loading = false
		if msg.err != nil {
			m.handleAPIError(msg.err)
			return m, nil
		}
		m.workspaces = msg.workspaces
		m.contentCursor = clampCursor(m.contentCursor, len(m.workspaces))
		return m, nil
	case loginResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			m.infoText = ""
			return m, nil
		}
		m.mode = modeApp
		m.errorText = ""
		m.infoText = "Login realizado"
		m.focusSidebar = true
		m.section = sectionDashboard
		m.sidebarCursor = 0
		m.contentCursor = 0
		m.loginPassword.SetValue("")
		m.lastTaskAction = nil
		m.lastTaskActionLabel = ""
		m.loading = true
		return m, tea.Batch(
			loadUserStateCmd(m.ctx, m.client),
			loadUsageStateCmd(m.ctx, m.client),
		)
	case logoutResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m = resetToLogin(m)
		m.infoText = "Sessão encerrada"
		return m, nil
	case createTaskResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.infoText = fmt.Sprintf("Tarefa criada: %s", msg.task.Title)
		m.errorText = ""
		m.resetTaskForm()
		m.mode = modeApp
		m.section = sectionTasks
		m.sidebarCursor = int(sectionTasks)
		m.focusSidebar = false
		m.loading = true
		return m, loadTasksStateCmd(m.ctx, m.client, false)
	case createAreaResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.infoText = fmt.Sprintf("Área criada: %s", msg.area.Name)
		m.errorText = ""
		m.resetAreaForm()
		m.mode = modeApp
		m.section = sectionAreas
		m.sidebarCursor = int(sectionAreas)
		m.focusSidebar = false
		m.loading = true
		return m, loadAreasStateCmd(m.ctx, m.client)
	case taskActionResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		verb := "concluída"
		if msg.reopen {
			verb = "reaberta"
		}
		if msg.undo {
			verb = "desfeita (" + verb + ")"
		}
		if msg.task != nil {
			m.infoText = fmt.Sprintf("Tarefa %s: %s", verb, msg.task.Title)
			if msg.undo {
				m.lastTaskAction = nil
				m.lastTaskActionLabel = ""
			} else {
				m.lastTaskAction = &taskAction{
					taskID: msg.task.ID,
					reopen: !msg.reopen,
				}
				actionLabel := "concluiu"
				if msg.reopen {
					actionLabel = "reabriu"
				}
				m.lastTaskActionLabel = fmt.Sprintf("%s %s", actionLabel, msg.task.Title)
			}
		}
		m.errorText = ""
		m.loading = true
		return m, loadTasksStateCmd(m.ctx, m.client, m.tasksCompleted)
	case workspaceSwitchResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.errorText = ""
		m.infoText = "Espaço de trabalho alterado"
		m.loading = true
		return m, tea.Batch(
			loadUserStateCmd(m.ctx, m.client),
			loadUsageStateCmd(m.ctx, m.client),
			loadWorkspacesStateCmd(m.ctx, m.client),
		)
	case tea.KeyMsg:
		if m.mode == modeLogin {
			return m.updateLogin(msg)
		}
		if m.mode == modeTaskForm {
			return m.updateTaskForm(msg)
		}
		if m.mode == modeAreaForm {
			return m.updateAreaForm(msg)
		}
		return m.updateApp(msg)
	}
	return m, nil
}

func (m tuiModel) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorText = ""
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab", "down":
		m.loginFocus = (m.loginFocus + 1) % 2
		m.applyLoginFocus()
		return m, nil
	case "shift+tab", "up":
		m.loginFocus = (m.loginFocus + 1) % 2
		m.applyLoginFocus()
		return m, nil
	case "enter":
		if m.loginFocus == 0 {
			m.loginFocus = 1
			m.applyLoginFocus()
			return m, nil
		}
		email := strings.TrimSpace(m.loginEmail.Value())
		password := strings.TrimSpace(m.loginPassword.Value())
		if email == "" || password == "" {
			m.errorText = "E-mail e senha são obrigatórios"
			return m, nil
		}
		m.loading = true
		m.infoText = ""
		return m, loginCmd(m.ctx, m.client, email, password)
	}

	var cmd tea.Cmd
	if m.loginFocus == 0 {
		m.loginEmail, cmd = m.loginEmail.Update(msg)
	} else {
		m.loginPassword, cmd = m.loginPassword.Update(msg)
	}
	return m, cmd
}

func (m tuiModel) updateTaskForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorText = ""
	switch msg.String() {
	case "esc":
		m.mode = modeApp
		m.resetTaskForm()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "tab":
		m.taskFocus = (m.taskFocus + 1) % taskFocusCount
		m.applyTaskFocus()
		return m, nil
	case "shift+tab":
		m.taskFocus = (m.taskFocus + taskFocusCount - 1) % taskFocusCount
		m.applyTaskFocus()
		return m, nil
	case "left", "h":
		switch m.taskFocus {
		case taskFocusPriority:
			m.taskPriority = (m.taskPriority + len(taskPriorities) - 1) % len(taskPriorities)
			return m, nil
		case taskFocusArea:
			m.taskAreaIdx = prevTaskSelectionIdx(m.taskAreaIdx, len(m.areas))
			return m, nil
		case taskFocusProject:
			m.taskProjectIdx = prevTaskSelectionIdx(m.taskProjectIdx, len(m.projects))
			return m, nil
		}
	case "right", "l":
		switch m.taskFocus {
		case taskFocusPriority:
			m.taskPriority = (m.taskPriority + 1) % len(taskPriorities)
			return m, nil
		case taskFocusArea:
			m.taskAreaIdx = nextTaskSelectionIdx(m.taskAreaIdx, len(m.areas))
			return m, nil
		case taskFocusProject:
			m.taskProjectIdx = nextTaskSelectionIdx(m.taskProjectIdx, len(m.projects))
			return m, nil
		}
	case "up":
		switch m.taskFocus {
		case taskFocusArea:
			m.taskAreaIdx = prevTaskSelectionIdx(m.taskAreaIdx, len(m.areas))
			return m, nil
		case taskFocusProject:
			m.taskProjectIdx = prevTaskSelectionIdx(m.taskProjectIdx, len(m.projects))
			return m, nil
		case taskFocusLabels:
			m.taskLabelCursor = wrapCursorUp(m.taskLabelCursor, len(m.labels))
			return m, nil
		}
	case "down":
		switch m.taskFocus {
		case taskFocusArea:
			m.taskAreaIdx = nextTaskSelectionIdx(m.taskAreaIdx, len(m.areas))
			return m, nil
		case taskFocusProject:
			m.taskProjectIdx = nextTaskSelectionIdx(m.taskProjectIdx, len(m.projects))
			return m, nil
		case taskFocusLabels:
			m.taskLabelCursor = wrapCursorDown(m.taskLabelCursor, len(m.labels))
			return m, nil
		}
	case " ":
		if m.taskFocus == taskFocusLabels {
			m.toggleTaskLabelSelection()
			return m, nil
		}
	case "enter":
		if m.taskFocus < taskFocusLabels {
			m.taskFocus++
			m.applyTaskFocus()
			return m, nil
		}
		title := strings.TrimSpace(m.taskTitle.Value())
		description := strings.TrimSpace(m.taskDesc.Value())
		dueDate, err := parseTaskDueDate(strings.TrimSpace(m.taskDueDate.Value()))
		if err != nil {
			m.errorText = err.Error()
			return m, nil
		}
		duration, err := parseTaskDuration(strings.TrimSpace(m.taskDuration.Value()))
		if err != nil {
			m.errorText = err.Error()
			return m, nil
		}
		if title == "" {
			m.errorText = "O título da tarefa é obrigatório"
			return m, nil
		}
		priority := taskPriorities[m.taskPriority]
		var desc *string
		if description != "" {
			desc = &description
		}
		areaID := taskSelectionID(m.taskAreaIdx, m.areas)
		projectID := taskProjectSelectionID(m.taskProjectIdx, m.projects)
		m.loading = true
		return m, createTaskCmd(
			m.ctx,
			m.client,
			TaskCreateRequest{
				Title:           title,
				Description:     desc,
				Priority:        priority,
				AreaID:          areaID,
				ProjectID:       projectID,
				DueDate:         dueDate,
				DurationMinutes: duration,
				LabelIDs:        m.taskSelectedLabelIDs(),
			},
		)
	}

	var cmd tea.Cmd
	switch m.taskFocus {
	case taskFocusTitle:
		m.taskTitle, cmd = m.taskTitle.Update(msg)
	case taskFocusDescription:
		m.taskDesc, cmd = m.taskDesc.Update(msg)
	case taskFocusDueDate:
		m.taskDueDate, cmd = m.taskDueDate.Update(msg)
	case taskFocusDuration:
		m.taskDuration, cmd = m.taskDuration.Update(msg)
	case taskFocusLabels:
		return m, nil
	}
	return m, cmd
}

func (m tuiModel) updateAreaForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.errorText = ""
	switch msg.String() {
	case "esc":
		m.mode = modeApp
		m.resetAreaForm()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	case "tab", "down":
		m.areaFocus = (m.areaFocus + 1) % 2
		m.applyAreaFocus()
		return m, nil
	case "shift+tab", "up":
		m.areaFocus = (m.areaFocus + 1) % 2
		m.applyAreaFocus()
		return m, nil
	case "enter":
		if m.areaFocus == 0 {
			m.areaFocus = 1
			m.applyAreaFocus()
			return m, nil
		}
		name := strings.TrimSpace(m.areaName.Value())
		slug := strings.TrimSpace(m.areaSlug.Value())
		if name == "" {
			m.errorText = "Nome da área é obrigatório"
			return m, nil
		}
		if slug == "" {
			slug = slugify(name)
		}
		if slug == "" {
			m.errorText = "Slug inválido"
			return m, nil
		}
		m.loading = true
		return m, createAreaCmd(m.ctx, m.client, name, slug)
	}

	var cmd tea.Cmd
	if m.areaFocus == 0 {
		m.areaName, cmd = m.areaName.Update(msg)
	} else {
		m.areaSlug, cmd = m.areaSlug.Update(msg)
	}
	return m, cmd
}

func (m tuiModel) updateApp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "tab":
		m.focusSidebar = !m.focusSidebar
		return m, nil
	case "r":
		m.loading = true
		m.errorText = ""
		return m, tea.Batch(loadUserStateCmd(m.ctx, m.client), m.loadSectionCmd())
	case "1", "2", "3", "4", "5":
		idx := int(msg.String()[0] - '1')
		if idx >= 0 && idx < len(sectionNames) {
			m.sidebarCursor = idx
			m.focusSidebar = true
			return m.selectSection(idx)
		}
	}

	if m.focusSidebar {
		switch msg.String() {
		case "up", "k":
			m.sidebarCursor = wrapCursorUp(m.sidebarCursor, len(sectionNames))
		case "down", "j":
			m.sidebarCursor = wrapCursorDown(m.sidebarCursor, len(sectionNames))
		case "enter", " ":
			return m.selectSection(m.sidebarCursor)
		}
		return m, nil
	}

	switch m.section {
	case sectionTasks:
		return m.updateTasksView(msg)
	case sectionAreas:
		return m.updateAreasView(msg)
	case sectionWorkspaces:
		return m.updateWorkspacesView(msg)
	case sectionSession:
		return m.updateSessionView(msg)
	}
	return m, nil
}

func (m tuiModel) updateTasksView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.contentCursor = wrapCursorUp(m.contentCursor, len(m.tasks))
	case "down", "j":
		m.contentCursor = wrapCursorDown(m.contentCursor, len(m.tasks))
	case "u":
		if m.lastTaskAction == nil {
			m.errorText = "Nada para desfazer"
			return m, nil
		}
		m.loading = true
		return m, undoTaskCmd(m.ctx, m.client, *m.lastTaskAction)
	case "f":
		m.loading = true
		m.tasksCompleted = !m.tasksCompleted
		return m, loadTasksStateCmd(m.ctx, m.client, m.tasksCompleted)
	case "n":
		m.resetTaskForm()
		m.mode = modeTaskForm
		m.taskFocus = taskFocusTitle
		m.applyTaskFocus()
		m.errorText = ""
		return m, tea.Batch(
			textinput.Blink,
			loadAreasStateCmd(m.ctx, m.client),
			loadProjectsStateCmd(m.ctx, m.client),
			loadLabelsStateCmd(m.ctx, m.client),
		)
	case "x", "enter":
		if len(m.tasks) == 0 {
			return m, nil
		}
		task := m.tasks[m.contentCursor]
		m.loading = true
		if m.tasksCompleted {
			return m, reopenTaskCmd(m.ctx, m.client, task.ID)
		}
		return m, completeTaskCmd(m.ctx, m.client, task.ID)
	}
	return m, nil
}

func (m tuiModel) updateAreasView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.contentCursor = wrapCursorUp(m.contentCursor, len(m.areas))
	case "down", "j":
		m.contentCursor = wrapCursorDown(m.contentCursor, len(m.areas))
	case "n", "enter":
		m.mode = modeAreaForm
		m.areaFocus = 0
		m.applyAreaFocus()
		m.errorText = ""
		return m, textinput.Blink
	}
	return m, nil
}

func (m tuiModel) updateWorkspacesView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.contentCursor = wrapCursorUp(m.contentCursor, len(m.workspaces))
	case "down", "j":
		m.contentCursor = wrapCursorDown(m.contentCursor, len(m.workspaces))
	case "enter", "s":
		if len(m.workspaces) == 0 {
			return m, nil
		}
		m.loading = true
		return m, switchWorkspaceCmd(m.ctx, m.client, m.workspaces[m.contentCursor].ID)
	}
	return m, nil
}

func (m tuiModel) updateSessionView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := []string{"Sair da sessão", "Sair"}
	switch msg.String() {
	case "up", "k":
		m.contentCursor = wrapCursorUp(m.contentCursor, len(actions))
	case "down", "j":
		m.contentCursor = wrapCursorDown(m.contentCursor, len(actions))
	case "enter":
		if m.contentCursor == 0 {
			m.loading = true
			return m, logoutCmd(m.ctx, m.client)
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m tuiModel) selectSection(idx int) (tea.Model, tea.Cmd) {
	m.section = appSection(idx)
	m.sidebarCursor = idx
	m.contentCursor = 0
	m.errorText = ""
	m.infoText = ""
	m.focusSidebar = false
	m.loading = true
	return m, m.loadSectionCmd()
}

func (m tuiModel) loadSectionCmd() tea.Cmd {
	switch m.section {
	case sectionDashboard:
		return loadUsageStateCmd(m.ctx, m.client)
	case sectionTasks:
		return loadTasksStateCmd(m.ctx, m.client, m.tasksCompleted)
	case sectionAreas:
		return loadAreasStateCmd(m.ctx, m.client)
	case sectionWorkspaces:
		return loadWorkspacesStateCmd(m.ctx, m.client)
	case sectionSession:
		return nil
	default:
		return nil
	}
}

func (m tuiModel) View() string {
	if m.mode == modeLogin {
		return m.viewLogin()
	}
	if m.mode == modeTaskForm {
		return m.viewTaskForm()
	}
	if m.mode == modeAreaForm {
		return m.viewAreaForm()
	}
	return m.viewApp()
}

func (m tuiModel) viewLogin() string {
	if m.width == 0 {
		m.width = 100
	}
	panelWidth := min(76, max(56, m.width-8))
	title := headerStyle.Render("WIDIA CLI")
	sub := subHeaderStyle.Render("Espaço de trabalho interativo no terminal")
	line := strings.Repeat("-", max(10, panelWidth-4))

	content := []string{
		title,
		sub,
		line,
		m.loginEmail.View(),
		m.loginPassword.View(),
		"",
		footerStyle.Render("Enter: próximo/confirmar  Tab: trocar campo  q: sair"),
	}
	if m.loading {
		content = append(content, infoStyle.Render("Autenticando..."))
	}
	if m.errorText != "" {
		content = append(content, errorStyle.Render(m.errorText))
	}
	if m.infoText != "" {
		content = append(content, infoStyle.Render(m.infoText))
	}

	box := mainStyle.Width(panelWidth).Render(strings.Join(content, "\n"))
	return lipgloss.Place(m.width, max(14, m.height), lipgloss.Center, lipgloss.Center, box)
}

func (m tuiModel) viewTaskForm() string {
	if m.width == 0 {
		m.width = 100
	}
	panelWidth := min(88, max(62, m.width-8))

	areaName := taskSelectionDisplayName(m.taskAreaIdx, m.areas, "Nenhuma")
	projectName := taskProjectSelectionDisplayName(m.taskProjectIdx, m.projects, "Nenhum")
	selectedLabels := taskSelectedLabelsSummary(m.labels, m.taskSelectedLabels)
	content := []string{
		headerStyle.Render("Nova tarefa"),
		subHeaderStyle.Render("Crie uma tarefa sem sair do teclado"),
		strings.Repeat("-", max(10, panelWidth-4)),
		m.taskTitle.View(),
		m.taskDesc.View(),
		m.taskDueDate.View(),
		fmt.Sprintf("Prioridade: %s", renderPrioritySelector(m.taskPriority)),
		fmt.Sprintf("Área: %s", areaName),
		fmt.Sprintf("Projeto: %s", projectName),
		m.taskDuration.View(),
		fmt.Sprintf("Etiquetas: %s", selectedLabels),
		"",
	}
	content = append(content, m.renderTaskAreaSelection()...)
	content = append(content, m.renderTaskProjectSelection()...)
	content = append(content, m.renderTaskLabelsSelection()...)
	content = append(content, footerStyle.Render("Tab: próximo  Shift+Tab: anterior  Enter: avançar/salvar  Esc: cancelar"))
	if m.loading {
		content = append(content, infoStyle.Render("Criando tarefa..."))
	}
	if m.errorText != "" {
		content = append(content, errorStyle.Render(m.errorText))
	}
	box := mainStyle.Width(panelWidth).Render(strings.Join(content, "\n"))
	return lipgloss.Place(m.width, max(14, m.height), lipgloss.Center, lipgloss.Center, box)
}

func (m tuiModel) viewAreaForm() string {
	if m.width == 0 {
		m.width = 100
	}
	panelWidth := min(88, max(62, m.width-8))
	content := []string{
		headerStyle.Render("Nova área"),
		subHeaderStyle.Render("Crie uma área no formulário em foco"),
		strings.Repeat("-", max(10, panelWidth-4)),
		m.areaName.View(),
		m.areaSlug.View(),
		"",
		footerStyle.Render("Enter: próximo/confirmar  Esc: cancelar"),
	}
	if m.loading {
		content = append(content, infoStyle.Render("Criando área..."))
	}
	if m.errorText != "" {
		content = append(content, errorStyle.Render(m.errorText))
	}
	box := mainStyle.Width(panelWidth).Render(strings.Join(content, "\n"))
	return lipgloss.Place(m.width, max(14, m.height), lipgloss.Center, lipgloss.Center, box)
}

func (m tuiModel) viewApp() string {
	if m.width == 0 {
		m.width = 120
	}
	if m.height == 0 {
		m.height = 32
	}

	userLabel := "desconhecido"
	if m.profile != nil && strings.TrimSpace(m.profile.Email) != "" {
		userLabel = m.profile.Email
	}
	workspaceLabel := "-"
	if m.workspace != nil {
		workspaceLabel = fmt.Sprintf("%s (%s)", m.workspace.Name, m.workspace.Slug)
	}

	header := headerStyle.Render("WIDIA CLI")
	subHeader := subHeaderStyle.Render(fmt.Sprintf("Usuário: %s   Espaço: %s", userLabel, workspaceLabel))

	sidebarW := 26
	if m.width < 86 {
		sidebarW = 22
	}
	bodyHeight := max(8, m.height-6)
	mainW := max(30, m.width-sidebarW-5)

	sidebar := sidebarStyle.Width(sidebarW).Height(bodyHeight).Render(m.renderSidebar())
	mainPanel := mainStyle.Width(mainW).Height(bodyHeight).Render(m.renderMainContent(mainW - 4))

	footer := m.renderFooter()
	if m.loading {
		footer = "Carregando...  " + footer
	}

	return rootStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			subHeader,
			lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPanel),
			footerStyle.Render(footer),
		),
	)
}

func (m tuiModel) renderSidebar() string {
	var out []string
	out = append(out, "Seções")
	out = append(out, "")
	for i, name := range sectionNames {
		marker := "  "
		if i == m.sidebarCursor {
			marker = "> "
		}
		line := marker + name
		style := lipgloss.NewStyle().Foreground(clrTextPrimary)
		if appSection(i) == m.section {
			style = style.Foreground(clrAccent).Bold(true)
		}
		if m.focusSidebar && i == m.sidebarCursor {
			style = style.Background(clrPanelSelected)
		}
		out = append(out, style.Render(line))
	}
	out = append(out, "")
	out = append(out, subHeaderStyle.Render("Tab troca foco"))
	out = append(out, subHeaderStyle.Render("1..5 muda de seção"))
	return strings.Join(out, "\n")
}

func (m tuiModel) renderMainContent(width int) string {
	title := headerStyle.Foreground(clrAccent).Render(sectionNames[m.section])
	lines := []string{title, ""}

	switch m.section {
	case sectionDashboard:
		lines = append(lines, m.renderDashboard()...)
	case sectionTasks:
		lines = append(lines, m.renderTasks()...)
	case sectionAreas:
		lines = append(lines, m.renderAreas()...)
	case sectionWorkspaces:
		lines = append(lines, m.renderWorkspaces()...)
	case sectionSession:
		lines = append(lines, m.renderSession()...)
	}

	if m.errorText != "" {
		lines = append(lines, "", errorStyle.Render(m.errorText))
	}
	if m.infoText != "" {
		lines = append(lines, "", infoStyle.Render(m.infoText))
	}

	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

func (m tuiModel) renderDashboard() []string {
	if m.usage == nil {
		return []string{"Sem dados do painel ainda. Pressione r para atualizar."}
	}
	rows := []string{
		fmt.Sprintf("Áreas             %d / %d", m.usage.Counters.AreasCount, m.usage.Limits.MaxAreas),
		fmt.Sprintf("Metas             %d / %d", m.usage.Counters.GoalsCount, m.usage.Limits.MaxGoals),
		fmt.Sprintf("Hábitos           %d / %d", m.usage.Counters.HabitsCount, m.usage.Limits.MaxHabits),
		fmt.Sprintf("Projetos          %d / %d", m.usage.Counters.ProjectsCount, m.usage.Limits.MaxProjects),
		fmt.Sprintf("Membros           %d / %d", m.usage.Counters.MembersCount, m.usage.Limits.MaxMembers),
		fmt.Sprintf("Tarefas hoje      %d / %d", m.usage.Counters.TasksCreatedToday, m.usage.Limits.MaxTasksPerDay),
		fmt.Sprintf("Transações/mês  %d / %d", m.usage.Counters.TransactionsMonth, m.usage.Limits.MaxTransactions),
	}
	return rows
}

func (m tuiModel) renderTasks() []string {
	filter := "Pendentes"
	if m.tasksCompleted {
		filter = "Concluídas"
	}
	lines := []string{fmt.Sprintf("Filtro: %s", filter)}
	if len(m.tasks) == 0 {
		lines = append(lines, "", "Nenhuma tarefa encontrada.")
		return lines
	}
	lines = append(lines, "")
	for i, task := range m.tasks {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		status := "pendente"
		if task.IsCompleted {
			status = "concluída"
		}
		line := fmt.Sprintf("%s[%s] %-9s %s", cursor, status, renderTaskPriority(task.Priority), task.Title)
		style := lipgloss.NewStyle().Foreground(clrTextPrimary)
		if i == m.contentCursor && !m.focusSidebar {
			style = style.Background(clrPanelSelected)
		}
		lines = append(lines, style.Render(line))
	}
	return lines
}

func (m tuiModel) renderAreas() []string {
	if len(m.areas) == 0 {
		return []string{"Nenhuma área encontrada. Pressione n para criar uma."}
	}
	lines := []string{}
	for i, area := range m.areas {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		active := "inativa"
		if area.IsActive {
			active = "ativa"
		}
		line := fmt.Sprintf("%s%-20s %-18s %s", cursor, area.Name, area.Slug, active)
		style := lipgloss.NewStyle().Foreground(clrTextPrimary)
		if i == m.contentCursor && !m.focusSidebar {
			style = style.Background(clrPanelSelected)
		}
		lines = append(lines, style.Render(line))
	}
	return lines
}

func (m tuiModel) renderWorkspaces() []string {
	if len(m.workspaces) == 0 {
		return []string{"Nenhum espaço de trabalho encontrado."}
	}
	lines := []string{}
	for i, workspace := range m.workspaces {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		defaultTag := ""
		if workspace.IsDefault {
			defaultTag = " padrão"
		}
		line := fmt.Sprintf("%s%-20s %-10s %s%s", cursor, workspace.Name, workspace.Role, workspace.Slug, defaultTag)
		style := lipgloss.NewStyle().Foreground(clrTextPrimary)
		if i == m.contentCursor && !m.focusSidebar {
			style = style.Background(clrPanelSelected)
		}
		lines = append(lines, style.Render(line))
	}
	return lines
}

func (m tuiModel) renderSession() []string {
	actions := []string{"Sair da sessão", "Sair"}
	lines := []string{"Ações da sessão", ""}
	for i, action := range actions {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		style := lipgloss.NewStyle().Foreground(clrTextPrimary)
		if i == m.contentCursor && !m.focusSidebar {
			style = style.Background(clrPanelSelected)
		}
		lines = append(lines, style.Render(cursor+action))
	}
	return lines
}

func (m tuiModel) renderTaskAreaSelection() []string {
	if m.taskFocus != taskFocusArea {
		return nil
	}
	lines := []string{"", "Área (←/→/↑/↓):"}
	lines = append(lines, "  [Sem área]")
	for i, area := range m.areas {
		cursor := "  "
		if i == m.taskAreaIdx {
			cursor = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursor, area.Name))
	}
	return lines
}

func (m tuiModel) renderTaskProjectSelection() []string {
	if m.taskFocus != taskFocusProject {
		return nil
	}
	lines := []string{"", "Projeto (←/→/↑/↓):"}
	lines = append(lines, "  [Sem projeto]")
	for i, project := range m.projects {
		cursor := "  "
		if i == m.taskProjectIdx {
			cursor = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s", cursor, project.Title))
	}
	return lines
}

func (m tuiModel) renderTaskLabelsSelection() []string {
	if m.taskFocus != taskFocusLabels {
		return nil
	}
	lines := []string{"", "Etiquetas (espaço: alternar):"}
	if len(m.labels) == 0 {
		return append(lines, "  Nenhuma etiqueta disponível.")
	}
	for i, label := range m.labels {
		selected := "[ ]"
		if m.taskSelectedLabels != nil && m.taskSelectedLabels[label.ID] {
			selected = "[x]"
		}
		cursor := "  "
		if i == m.taskLabelCursor {
			cursor = "> "
		}
		lines = append(lines, fmt.Sprintf("%s%s %s", cursor, selected, label.Name))
	}
	return lines
}

func (m tuiModel) renderFooter() string {
	if m.focusSidebar {
		return "up/down: navegar seções  enter: abrir  tab: focar menu lateral  r: atualizar  q: sair"
	}
	switch m.section {
	case sectionTasks:
		footer := "up/down: selecionar  enter/x: alternar concluída  f: filtro  u: desfazer  n: nova tarefa  tab: menu lateral  r: atualizar"
		if m.lastTaskActionLabel != "" {
			return fmt.Sprintf("%s  |  Última ação: %s [u]", footer, m.lastTaskActionLabel)
		}
		return footer
	case sectionAreas:
		return "up/down: selecionar  n/enter: nova área  tab: menu lateral  r: atualizar"
	case sectionWorkspaces:
		return "up/down: selecionar  enter: trocar espaço de trabalho  tab: menu lateral  r: atualizar"
	case sectionSession:
		return "up/down: selecionar  enter: executar ação  tab: menu lateral"
	default:
		return "tab: menu lateral/conteúdo  r: atualizar  q: sair"
	}
}

func (m *tuiModel) applyLoginFocus() {
	if m.loginFocus == 0 {
		m.loginEmail.Focus()
		m.loginPassword.Blur()
		return
	}
	m.loginEmail.Blur()
	m.loginPassword.Focus()
}

func (m *tuiModel) applyTaskFocus() {
	switch m.taskFocus {
	case taskFocusTitle:
		m.taskTitle.Focus()
		m.taskDesc.Blur()
		m.taskDueDate.Blur()
		m.taskDuration.Blur()
	case taskFocusDescription:
		m.taskTitle.Blur()
		m.taskDesc.Focus()
		m.taskDueDate.Blur()
		m.taskDuration.Blur()
	case taskFocusDueDate:
		m.taskTitle.Blur()
		m.taskDesc.Blur()
		m.taskDueDate.Focus()
		m.taskDuration.Blur()
	case taskFocusDuration:
		m.taskTitle.Blur()
		m.taskDesc.Blur()
		m.taskDueDate.Blur()
		m.taskDuration.Focus()
	default:
		m.taskTitle.Blur()
		m.taskDesc.Blur()
		m.taskDueDate.Blur()
		m.taskDuration.Blur()
	}
}

func (m *tuiModel) applyAreaFocus() {
	if m.areaFocus == 0 {
		m.areaName.Focus()
		m.areaSlug.Blur()
		return
	}
	m.areaName.Blur()
	m.areaSlug.Focus()
}

func (m *tuiModel) resetTaskForm() {
	m.taskTitle.SetValue("")
	m.taskDesc.SetValue("")
	m.taskDueDate.SetValue("")
	m.taskDuration.SetValue("")
	m.taskPriority = 1
	m.taskAreaIdx = -1
	m.taskProjectIdx = -1
	m.taskLabelCursor = 0
	m.taskSelectedLabels = map[string]bool{}
	m.taskFocus = taskFocusTitle
	m.applyTaskFocus()
}

func (m *tuiModel) resetAreaForm() {
	m.areaName.SetValue("")
	m.areaSlug.SetValue("")
	m.areaFocus = 0
	m.applyAreaFocus()
}

func (m *tuiModel) ensureTaskFormAreaDefaults() {
	if m.taskAreaIdx >= len(m.areas) {
		m.taskAreaIdx = -1
	}
}

func (m *tuiModel) ensureTaskFormProjectDefaults() {
	if m.taskProjectIdx >= len(m.projects) {
		m.taskProjectIdx = -1
	}
}

func (m *tuiModel) ensureTaskFormLabelCursor() {
	if len(m.labels) == 0 {
		m.taskLabelCursor = 0
		return
	}
	if m.taskLabelCursor < 0 || m.taskLabelCursor >= len(m.labels) {
		m.taskLabelCursor = 0
	}
	for _, lbl := range m.labels {
		if m.taskSelectedLabels != nil && m.taskSelectedLabels[lbl.ID] {
			return
		}
	}
}

func (m *tuiModel) taskSelectedLabelIDs() []string {
	if len(m.labels) == 0 || len(m.taskSelectedLabels) == 0 {
		return nil
	}
	ids := make([]string, 0, len(m.taskSelectedLabels))
	for _, label := range m.labels {
		if m.taskSelectedLabels[label.ID] {
			ids = append(ids, label.ID)
		}
	}
	return ids
}

func (m *tuiModel) toggleTaskLabelSelection() {
	if len(m.labels) == 0 {
		return
	}
	if m.taskLabelCursor < 0 || m.taskLabelCursor >= len(m.labels) {
		m.taskLabelCursor = 0
	}
	id := m.labels[m.taskLabelCursor].ID
	if m.taskSelectedLabels == nil {
		m.taskSelectedLabels = map[string]bool{}
	}
	if m.taskSelectedLabels[id] {
		delete(m.taskSelectedLabels, id)
		return
	}
	m.taskSelectedLabels[id] = true
}

func taskSelectionDisplayName(idx int, areas []Area, noneLabel string) string {
	if idx < 0 || idx >= len(areas) {
		return noneLabel
	}
	return areas[idx].Name
}

func taskProjectSelectionDisplayName(idx int, projects []Project, noneLabel string) string {
	if idx < 0 || idx >= len(projects) {
		return noneLabel
	}
	return projects[idx].Title
}

func taskSelectedLabelsSummary(labels []Label, selected map[string]bool) string {
	items := make([]string, 0)
	for _, label := range labels {
		if selected != nil && selected[label.ID] {
			items = append(items, label.Name)
		}
	}
	if len(items) == 0 {
		return "nenhuma"
	}
	return strings.Join(items, ", ")
}

func taskSelectionID(idx int, areas []Area) *string {
	if idx < 0 || idx >= len(areas) {
		return nil
	}
	id := areas[idx].ID
	return &id
}

func taskProjectSelectionID(idx int, projects []Project) *string {
	if idx < 0 || idx >= len(projects) {
		return nil
	}
	id := projects[idx].ID
	return &id
}

func nextTaskSelectionIdx(current, total int) int {
	if total <= 0 {
		return -1
	}
	if current >= total-1 {
		return -1
	}
	return current + 1
}

func prevTaskSelectionIdx(current, total int) int {
	if total <= 0 {
		return -1
	}
	if current <= 0 {
		return total - 1
	}
	if current == -1 {
		return total - 1
	}
	return current - 1
}

func parseTaskDueDate(raw string) (*string, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, errors.New("data inválida, use AAAA-MM-DD")
	}
	value := parsed.Format("2006-01-02")
	return &value, nil
}

func parseTaskDuration(raw string) (*int, error) {
	if raw == "" {
		return nil, nil
	}
	if raw[0] == '+' {
		raw = raw[1:]
	}
	if strings.Contains(raw, ":") {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return nil, errors.New("duração inválida, use minutos ou formato HH:MM")
		}
		hours, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, errors.New("duração inválida, use minutos ou formato HH:MM")
		}
		minutes, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.New("duração inválida, use minutos ou formato HH:MM")
		}
		if hours < 0 || minutes < 0 {
			return nil, errors.New("duração não pode ser negativa")
		}
		total := hours*60 + minutes
		return &total, nil
	}

	value, err := time.ParseDuration(raw)
	if err == nil {
		total := int(value.Minutes())
		if total < 0 {
			return nil, errors.New("duração não pode ser negativa")
		}
		return &total, nil
	}
	minutes, err := strconv.Atoi(raw)
	if err != nil {
		return nil, errors.New("duração inválida, use minutos ou formato HH:MM")
	}
	if minutes < 0 {
		return nil, errors.New("duração não pode ser negativa")
	}
	return &minutes, nil
}

func (m *tuiModel) handleAPIError(err error) {
	if err == nil {
		return
	}
	if isAuthError(err) {
		*m = resetToLogin(*m)
		m.errorText = "Sessão expirada. Faça login novamente"
		return
	}
	m.errorText = err.Error()
}

func resetToLogin(m tuiModel) tuiModel {
	m.mode = modeLogin
	m.section = sectionDashboard
	m.sidebarCursor = 0
	m.contentCursor = 0
	m.focusSidebar = true
	m.tasks = nil
	m.areas = nil
	m.workspaces = nil
	m.usage = nil
	m.profile = nil
	m.workspace = nil
	m.projects = nil
	m.labels = nil
	m.lastTaskAction = nil
	m.lastTaskActionLabel = ""
	m.taskSelectedLabels = nil
	m.loading = false
	m.loginEmail.Focus()
	m.loginPassword.Blur()
	m.loginPassword.SetValue("")
	return m
}

func loadUserStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.Me(ctx)
		if err != nil {
			return userStateMsg{err: err}
		}
		workspace, _ := client.CurrentWorkspace(ctx)
		return userStateMsg{profile: profile, workspace: workspace}
	}
}

func loadUsageStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		usage, err := client.GetWorkspaceUsage(ctx)
		return usageStateMsg{usage: usage, err: err}
	}
}

func loadTasksStateCmd(ctx context.Context, client *Client, completed bool) tea.Cmd {
	return func() tea.Msg {
		tasks, err := client.ListTasks(ctx, &completed)
		return tasksStateMsg{tasks: tasks, completed: completed, err: err}
	}
}

func loadAreasStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		areas, err := client.ListAreas(ctx)
		return areasStateMsg{areas: areas, err: err}
	}
}

func loadProjectsStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		projects, err := client.ListProjects(ctx)
		return projectsStateMsg{projects: projects, err: err}
	}
}

func loadLabelsStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		labels, err := client.ListLabels(ctx)
		return labelsStateMsg{labels: labels, err: err}
	}
}

func loadWorkspacesStateCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		items, err := client.ListWorkspaces(ctx)
		return workspacesStateMsg{workspaces: items, err: err}
	}
}

func loginCmd(ctx context.Context, client *Client, email, password string) tea.Cmd {
	return func() tea.Msg {
		err := client.Login(ctx, email, password)
		return loginResultMsg{err: err}
	}
}

func logoutCmd(ctx context.Context, client *Client) tea.Cmd {
	return func() tea.Msg {
		err := client.Logout(ctx)
		return logoutResultMsg{err: err}
	}
}

func createTaskCmd(ctx context.Context, client *Client, req TaskCreateRequest) tea.Cmd {
	return func() tea.Msg {
		task, err := client.CreateTask(ctx, req)
		return createTaskResultMsg{task: task, err: err}
	}
}

func createAreaCmd(ctx context.Context, client *Client, name, slug string) tea.Cmd {
	return func() tea.Msg {
		area, err := client.CreateArea(ctx, name, slug)
		return createAreaResultMsg{area: area, err: err}
	}
}

func completeTaskCmd(ctx context.Context, client *Client, id string) tea.Cmd {
	return func() tea.Msg {
		task, err := client.CompleteTask(ctx, id)
		return taskActionResultMsg{task: task, reopen: false, err: err}
	}
}

func reopenTaskCmd(ctx context.Context, client *Client, id string) tea.Cmd {
	return func() tea.Msg {
		task, err := client.ReopenTask(ctx, id)
		return taskActionResultMsg{task: task, reopen: true, err: err}
	}
}

func undoTaskCmd(ctx context.Context, client *Client, action taskAction) tea.Cmd {
	return func() tea.Msg {
		if action.reopen {
			task, err := client.ReopenTask(ctx, action.taskID)
			return taskActionResultMsg{task: task, reopen: true, undo: true, err: err}
		}
		task, err := client.CompleteTask(ctx, action.taskID)
		return taskActionResultMsg{task: task, reopen: false, undo: true, err: err}
	}
}

func switchWorkspaceCmd(ctx context.Context, client *Client, workspaceID string) tea.Cmd {
	return func() tea.Msg {
		err := client.SwitchWorkspace(ctx, workspaceID)
		return workspaceSwitchResultMsg{err: err}
	}
}

func renderPrioritySelector(selected int) string {
	parts := make([]string, 0, len(taskPriorities))
	for i := range taskPriorities {
		value := taskPriorityLabels[i]
		if i == selected {
			parts = append(parts, lipgloss.NewStyle().Foreground(clrAccent).Bold(true).Render("["+value+"]"))
			continue
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, "  ")
}

func renderTaskPriority(value string) string {
	for i, priority := range taskPriorities {
		if value == priority && i < len(taskPriorityLabels) {
			return taskPriorityLabels[i]
		}
	}
	return value
}

func clampCursor(cursor, total int) int {
	if total <= 0 {
		return 0
	}
	if cursor >= total {
		return total - 1
	}
	if cursor < 0 {
		return 0
	}
	return cursor
}

func wrapCursorUp(cursor, total int) int {
	if total <= 0 {
		return 0
	}
	if cursor <= 0 {
		return total - 1
	}
	return cursor - 1
}

func wrapCursorDown(cursor, total int) int {
	if total <= 0 {
		return 0
	}
	return (cursor + 1) % total
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotAuthenticated) || errors.Is(err, ErrNoRefreshToken) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Status == 401
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
