package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	"Dashboard",
	"Tasks",
	"Areas",
	"Workspaces",
	"Session",
}

var taskPriorities = []string{"low", "medium", "high", "critical"}

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

	loginEmail    textinput.Model
	loginPassword textinput.Model
	loginFocus    int

	taskTitle    textinput.Model
	taskDesc     textinput.Model
	taskPriority int
	taskFocus    int

	areaName  textinput.Model
	areaSlug  textinput.Model
	areaFocus int

	lastTaskAction *taskAction

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
	email.Placeholder = "email"
	email.Prompt = "Email: "
	email.CharLimit = 120
	email.Width = 42

	password := textinput.New()
	password.Placeholder = "password"
	password.Prompt = "Password: "
	password.CharLimit = 120
	password.Width = 42
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '*'

	title := textinput.New()
	title.Placeholder = "Task title"
	title.Prompt = "Title: "
	title.CharLimit = 180
	title.Width = 60

	desc := textinput.New()
	desc.Placeholder = "Description (optional)"
	desc.Prompt = "Description: "
	desc.CharLimit = 300
	desc.Width = 60

	areaName := textinput.New()
	areaName.Placeholder = "Area name"
	areaName.Prompt = "Name: "
	areaName.CharLimit = 120
	areaName.Width = 60

	areaSlug := textinput.New()
	areaSlug.Placeholder = "area-slug"
	areaSlug.Prompt = "Slug: "
	areaSlug.CharLimit = 120
	areaSlug.Width = 60

	mode := modeLogin
	if client.session.IsAuthenticated() {
		mode = modeApp
	}

	m := tuiModel{
		ctx:            ctx,
		client:         client,
		mode:           mode,
		section:        sectionDashboard,
		sidebarCursor:  0,
		focusSidebar:   true,
		tasksCompleted: false,
		loginEmail:     email,
		loginPassword:  password,
		taskTitle:      title,
		taskDesc:       desc,
		taskPriority:   1,
		areaName:       areaName,
		areaSlug:       areaSlug,
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
		m.infoText = "Logged in"
		m.focusSidebar = true
		m.section = sectionDashboard
		m.sidebarCursor = 0
		m.contentCursor = 0
		m.loginPassword.SetValue("")
		m.lastTaskAction = nil
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
		m.infoText = "Session cleared"
		return m, nil
	case createTaskResultMsg:
		m.loading = false
		if msg.err != nil {
			m.errorText = msg.err.Error()
			return m, nil
		}
		m.infoText = fmt.Sprintf("Task created: %s", msg.task.Title)
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
		m.infoText = fmt.Sprintf("Area created: %s", msg.area.Name)
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
		verb := "completed"
		if msg.reopen {
			verb = "reopened"
		}
		if msg.undo {
			verb = "undone (" + verb + ")"
		}
		if msg.task != nil {
			m.infoText = fmt.Sprintf("Task %s: %s", verb, msg.task.Title)
			if msg.undo {
				m.lastTaskAction = nil
			} else {
				m.lastTaskAction = &taskAction{
					taskID: msg.task.ID,
					reopen: !msg.reopen,
				}
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
		m.infoText = "Workspace switched"
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
			m.errorText = "Email and password are required"
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
	case "tab", "down":
		m.taskFocus = (m.taskFocus + 1) % 3
		m.applyTaskFocus()
		return m, nil
	case "shift+tab", "up":
		m.taskFocus = (m.taskFocus + 2) % 3
		m.applyTaskFocus()
		return m, nil
	case "left", "h":
		if m.taskFocus == 2 {
			m.taskPriority = (m.taskPriority + len(taskPriorities) - 1) % len(taskPriorities)
		}
		return m, nil
	case "right", "l":
		if m.taskFocus == 2 {
			m.taskPriority = (m.taskPriority + 1) % len(taskPriorities)
		}
		return m, nil
	case "enter":
		if m.taskFocus < 2 {
			m.taskFocus++
			m.applyTaskFocus()
			return m, nil
		}
		title := strings.TrimSpace(m.taskTitle.Value())
		description := strings.TrimSpace(m.taskDesc.Value())
		if title == "" {
			m.errorText = "Task title is required"
			return m, nil
		}
		priority := taskPriorities[m.taskPriority]
		m.loading = true
		return m, createTaskCmd(m.ctx, m.client, title, description, priority)
	}

	var cmd tea.Cmd
	if m.taskFocus == 0 {
		m.taskTitle, cmd = m.taskTitle.Update(msg)
	} else if m.taskFocus == 1 {
		m.taskDesc, cmd = m.taskDesc.Update(msg)
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
			m.errorText = "Area name is required"
			return m, nil
		}
		if slug == "" {
			slug = slugify(name)
		}
		if slug == "" {
			m.errorText = "Invalid slug"
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
		m.mode = modeTaskForm
		m.taskFocus = 0
		m.applyTaskFocus()
		m.errorText = ""
		return m, textinput.Blink
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
	actions := []string{"Logout", "Quit"}
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
	sub := subHeaderStyle.Render("Interactive terminal workspace")
	line := strings.Repeat("-", max(10, panelWidth-4))

	content := []string{
		title,
		sub,
		line,
		m.loginEmail.View(),
		m.loginPassword.View(),
		"",
		footerStyle.Render("Enter: next/submit  Tab: switch field  q: quit"),
	}
	if m.loading {
		content = append(content, infoStyle.Render("Authenticating..."))
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
	content := []string{
		headerStyle.Render("New Task"),
		subHeaderStyle.Render("Create a task without leaving the keyboard"),
		strings.Repeat("-", max(10, panelWidth-4)),
		m.taskTitle.View(),
		m.taskDesc.View(),
		fmt.Sprintf("Priority: %s", renderPrioritySelector(m.taskPriority)),
		"",
		footerStyle.Render("Enter: next/submit  Left/Right: priority  Esc: cancel"),
	}
	if m.loading {
		content = append(content, infoStyle.Render("Creating task..."))
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
		headerStyle.Render("New Area"),
		subHeaderStyle.Render("Create an area from a focused form"),
		strings.Repeat("-", max(10, panelWidth-4)),
		m.areaName.View(),
		m.areaSlug.View(),
		"",
		footerStyle.Render("Enter: next/submit  Esc: cancel"),
	}
	if m.loading {
		content = append(content, infoStyle.Render("Creating area..."))
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

	userLabel := "unknown"
	if m.profile != nil && strings.TrimSpace(m.profile.Email) != "" {
		userLabel = m.profile.Email
	}
	workspaceLabel := "-"
	if m.workspace != nil {
		workspaceLabel = fmt.Sprintf("%s (%s)", m.workspace.Name, m.workspace.Slug)
	}

	header := headerStyle.Render("WIDIA CLI")
	subHeader := subHeaderStyle.Render(fmt.Sprintf("User: %s   Workspace: %s", userLabel, workspaceLabel))

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
		footer = "Loading...  " + footer
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
	out = append(out, "Sections")
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
	out = append(out, subHeaderStyle.Render("Tab switches focus"))
	out = append(out, subHeaderStyle.Render("1..5 jumps section"))
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
		return []string{"No dashboard data yet. Press r to refresh."}
	}
	rows := []string{
		fmt.Sprintf("Areas            %d / %d", m.usage.Counters.AreasCount, m.usage.Limits.MaxAreas),
		fmt.Sprintf("Goals            %d / %d", m.usage.Counters.GoalsCount, m.usage.Limits.MaxGoals),
		fmt.Sprintf("Habits           %d / %d", m.usage.Counters.HabitsCount, m.usage.Limits.MaxHabits),
		fmt.Sprintf("Projects         %d / %d", m.usage.Counters.ProjectsCount, m.usage.Limits.MaxProjects),
		fmt.Sprintf("Members          %d / %d", m.usage.Counters.MembersCount, m.usage.Limits.MaxMembers),
		fmt.Sprintf("Tasks Today      %d / %d", m.usage.Counters.TasksCreatedToday, m.usage.Limits.MaxTasksPerDay),
		fmt.Sprintf("Transactions Mo  %d / %d", m.usage.Counters.TransactionsMonth, m.usage.Limits.MaxTransactions),
	}
	return rows
}

func (m tuiModel) renderTasks() []string {
	filter := "Pending"
	if m.tasksCompleted {
		filter = "Completed"
	}
	lines := []string{fmt.Sprintf("Filter: %s", filter)}
	if len(m.tasks) == 0 {
		lines = append(lines, "", "No tasks found.")
		return lines
	}
	lines = append(lines, "")
	for i, task := range m.tasks {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		status := "todo"
		if task.IsCompleted {
			status = "done"
		}
		line := fmt.Sprintf("%s[%s] %-8s %s", cursor, status, task.Priority, task.Title)
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
		return []string{"No areas found. Press n to create one."}
	}
	lines := []string{}
	for i, area := range m.areas {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		active := "inactive"
		if area.IsActive {
			active = "active"
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
		return []string{"No workspaces found."}
	}
	lines := []string{}
	for i, workspace := range m.workspaces {
		cursor := "  "
		if i == m.contentCursor && !m.focusSidebar {
			cursor = "> "
		}
		defaultTag := ""
		if workspace.IsDefault {
			defaultTag = " default"
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
	actions := []string{"Logout", "Quit"}
	lines := []string{"Session actions", ""}
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

func (m tuiModel) renderFooter() string {
	if m.focusSidebar {
		return "up/down: navigate sections  enter: open  tab: focus content  r: refresh  q: quit"
	}
	switch m.section {
	case sectionTasks:
		return "up/down: select  enter/x: toggle done  f: filter  u: undo  n: new task  tab: sidebar  r: refresh"
	case sectionAreas:
		return "up/down: select  n/enter: new area  tab: sidebar  r: refresh"
	case sectionWorkspaces:
		return "up/down: select  enter: switch workspace  tab: sidebar  r: refresh"
	case sectionSession:
		return "up/down: select  enter: run action  tab: sidebar"
	default:
		return "tab: sidebar/content  r: refresh  q: quit"
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
	case 0:
		m.taskTitle.Focus()
		m.taskDesc.Blur()
	case 1:
		m.taskTitle.Blur()
		m.taskDesc.Focus()
	default:
		m.taskTitle.Blur()
		m.taskDesc.Blur()
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
	m.taskPriority = 1
	m.taskFocus = 0
	m.applyTaskFocus()
}

func (m *tuiModel) resetAreaForm() {
	m.areaName.SetValue("")
	m.areaSlug.SetValue("")
	m.areaFocus = 0
	m.applyAreaFocus()
}

func (m *tuiModel) handleAPIError(err error) {
	if err == nil {
		return
	}
	if isAuthError(err) {
		*m = resetToLogin(*m)
		m.errorText = "Session expired. Please login again"
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
	m.lastTaskAction = nil
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

func createTaskCmd(ctx context.Context, client *Client, title, description, priority string) tea.Cmd {
	return func() tea.Msg {
		task, err := client.CreateTask(ctx, title, description, priority)
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
	for i, value := range taskPriorities {
		if i == selected {
			parts = append(parts, lipgloss.NewStyle().Foreground(clrAccent).Bold(true).Render("["+value+"]"))
			continue
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, "  ")
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
