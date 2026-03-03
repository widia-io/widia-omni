package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

func RunInteractive(ctx context.Context, reader *bufio.Reader, client *Client) error {
	_ = reader
	return RunTUI(ctx, client)
}

func loginFlow(ctx context.Context, reader *bufio.Reader, client *Client) error {
	email, err := readLine(reader, "E-mail: ")
	if err != nil {
		return err
	}
	password, err := readLine(reader, "Senha: ")
	if err != nil {
		return err
	}
	return client.Login(ctx, email, password)
}

func showDashboard(ctx context.Context, client *Client) error {
	usage, err := client.GetWorkspaceUsage(ctx)
	if err != nil {
		return err
	}
	fmt.Println("\n--- Dashboard ---")
	fmt.Printf("Áreas: %d / %d\n", usage.Counters.AreasCount, usage.Limits.MaxAreas)
	fmt.Printf("Metas: %d / %d\n", usage.Counters.GoalsCount, usage.Limits.MaxGoals)
	fmt.Printf("Hábitos: %d / %d\n", usage.Counters.HabitsCount, usage.Limits.MaxHabits)
	fmt.Printf("Projetos: %d / %d\n", usage.Counters.ProjectsCount, usage.Limits.MaxProjects)
	fmt.Printf("Membros: %d / %d\n", usage.Counters.MembersCount, usage.Limits.MaxMembers)
	fmt.Printf("Tarefas hoje: %d / %d\n", usage.Counters.TasksCreatedToday, usage.Limits.MaxTasksPerDay)
	fmt.Printf("Transações (mês): %d / %d\n", usage.Counters.TransactionsMonth, usage.Limits.MaxTransactions)
	return nil
}

func workspaceMenu(ctx context.Context, reader *bufio.Reader, client *Client) error {
	for {
		choice, err := selectOption(reader, "Workspaces", []string{
			"Listar",
			"Trocar workspace",
			"Voltar",
		})
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			items, err := client.ListWorkspaces(ctx)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				fmt.Println("Nenhum workspace encontrado.")
				continue
			}
			fmt.Println("\nWorkspaces:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Nome\tSlug\tPapel\tPadrão\t")
			for _, it := range items {
				def := "Não"
				if it.IsDefault {
					def = "Sim"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", it.Name, it.Slug, it.Role, def)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			waitForEnter(reader)
		case 1:
			if err := switchWorkspaceFlow(ctx, reader, client); err != nil {
				return err
			}
		case 2:
			return nil
		}
	}
}

func switchWorkspaceFlow(ctx context.Context, reader *bufio.Reader, client *Client) error {
	items, err := client.ListWorkspaces(ctx)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		fmt.Println("Nenhum workspace para alternar.")
		return nil
	}
	fmt.Println()
	opts := make([]string, 0, len(items))
	for _, it := range items {
		defaultLabel := " "
		if it.IsDefault {
			defaultLabel = "*"
		}
		opts = append(opts, fmt.Sprintf("%s %s (%s)", defaultLabel, it.Name, it.Role))
	}
	choice, err := selectOption(reader, "Escolha o workspace", opts)
	if err != nil {
		return err
	}
	return client.SwitchWorkspace(ctx, items[choice].ID)
}

func tasksMenu(ctx context.Context, reader *bufio.Reader, client *Client) error {
	for {
		choice, err := selectOption(reader, "Tarefas", []string{
			"Listar pendentes",
			"Listar concluídas",
			"Nova tarefa",
			"Concluir tarefa",
			"Reabrir tarefa",
			"Voltar",
		})
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			if err := listTasksFlow(ctx, client, false); err != nil {
				return err
			}
		case 1:
			if err := listTasksFlow(ctx, client, true); err != nil {
				return err
			}
		case 2:
			if err := createTaskFlow(ctx, reader, client); err != nil {
				return err
			}
		case 3:
			if err := completeTaskFlow(ctx, reader, client, false); err != nil {
				return err
			}
		case 4:
			if err := completeTaskFlow(ctx, reader, client, true); err != nil {
				return err
			}
		case 5:
			return nil
		}
		waitForEnter(reader)
	}
}

func listTasksFlow(ctx context.Context, client *Client, completed bool) error {
	tasks, err := client.ListTasks(ctx, &completed)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("Nenhuma tarefa encontrada.")
		return nil
	}
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tStatus\tPrioridade\tTítulo")
	for idx, t := range tasks {
		status := "Pendente"
		if t.IsCompleted {
			status = "Concluída"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", idx+1, status, t.Priority, t.Title)
	}
	return w.Flush()
}

func createTaskFlow(ctx context.Context, reader *bufio.Reader, client *Client) error {
	title, err := readLine(reader, "Título da tarefa: ")
	if err != nil {
		return err
	}
	if strings.TrimSpace(title) == "" {
		return errors.New("título vazio")
	}

	description, err := readLine(reader, "Descrição (opcional): ")
	if err != nil {
		return err
	}

	priority, err := selectOption(reader, "Prioridade", []string{
		"baixa",
		"média",
		"alta",
		"crítica",
	})
	if err != nil {
		return err
	}

	priorities := []string{"low", "medium", "high", "critical"}
	req := TaskCreateRequest{
		Title:       title,
		Description: &description,
		Priority:    priorities[priority],
	}
	task, err := client.CreateTask(ctx, req)
	if err != nil {
		return err
	}
	fmt.Printf("Tarefa criada: %s (%s)\n", task.Title, task.ID)
	return nil
}

func completeTaskFlow(ctx context.Context, reader *bufio.Reader, client *Client, reopen bool) error {
	completedFilter := false
	actionVerb := "concluir"
	if reopen {
		completedFilter = true
		actionVerb = "reabrir"
	}

	tasks, err := client.ListTasks(ctx, &completedFilter)
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("Nenhuma tarefa para essa ação.")
		return nil
	}

	opts := make([]string, 0, len(tasks))
	for _, task := range tasks {
		opts = append(opts, task.Title)
	}
	idx, err := selectOption(reader, "Escolha a tarefa", opts)
	if err != nil {
		return err
	}

	var updated *Task
	if reopen {
		updated, err = client.ReopenTask(ctx, tasks[idx].ID)
	} else {
		updated, err = client.CompleteTask(ctx, tasks[idx].ID)
	}
	if err != nil {
		return err
	}
	if updated != nil {
		fmt.Printf("Tarefa %s: %s\n", actionVerb, updated.Title)
	}
	return nil
}

func areasMenu(ctx context.Context, reader *bufio.Reader, client *Client) error {
	for {
		choice, err := selectOption(reader, "Áreas", []string{
			"Listar",
			"Criar",
			"Voltar",
		})
		if err != nil {
			return err
		}

		switch choice {
		case 0:
			areas, err := client.ListAreas(ctx)
			if err != nil {
				return err
			}
			if len(areas) == 0 {
				fmt.Println("Nenhuma área cadastrada.")
				continue
			}
			fmt.Println("\nÁreas:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "Nome\tSlug\tÍcone\tCor\tAtiva")
			for _, a := range areas {
				active := "Não"
				if a.IsActive {
					active = "Sim"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", a.Name, a.Slug, a.Icon, a.Color, active)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			waitForEnter(reader)
		case 1:
			if err := createAreaFlow(ctx, reader, client); err != nil {
				return err
			}
		case 2:
			return nil
		}
	}
}

func createAreaFlow(ctx context.Context, reader *bufio.Reader, client *Client) error {
	name, err := readLine(reader, "Nome da área: ")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("nome vazio")
	}

	slug, err := readLine(reader, "Slug (opcional): ")
	if err != nil {
		return err
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		slug = slugify(name)
	}
	if slug == "" {
		return errors.New("slug vazio")
	}

	area, err := client.CreateArea(ctx, name, slug)
	if err != nil {
		return err
	}
	fmt.Printf("Área criada: %s (%s)\n", area.Name, area.ID)
	return nil
}

func selectOption(reader *bufio.Reader, title string, options []string) (int, error) {
	for {
		fmt.Printf("\n%s\n", title)
		for i, option := range options {
			fmt.Printf("  %d) %s\n", i+1, option)
		}
		line, err := readLine(reader, "Escolha: ")
		if err != nil {
			return 0, err
		}
		idx, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || idx < 1 || idx > len(options) {
			fmt.Println("Opção inválida.")
			continue
		}
		return idx - 1, nil
	}
}

func waitForEnter(reader *bufio.Reader) {
	_, _ = readLine(reader, "Pressione Enter para continuar...")
}

func readLine(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func slugify(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	var builder strings.Builder
	wasDash := false
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			wasDash = false
			continue
		}
		if r == ' ' || r == '-' || r == '_' {
			if !wasDash {
				builder.WriteByte('-')
				wasDash = true
			}
			continue
		}
	}
	out := strings.Trim(builder.String(), "-")
	return out
}

func RunInteractiveLogin(ctx context.Context, reader *bufio.Reader, client *Client) error {
	if reader == nil {
		return io.EOF
	}
	return loginFlow(ctx, reader, client)
}

func Logout(ctx context.Context, client *Client) error {
	return client.Logout(ctx)
}

func ShowStatus(ctx context.Context, client *Client) error {
	if !client.session.IsAuthenticated() {
		fmt.Println("Sem sessão ativa")
		return nil
	}
	profile, err := client.Me(ctx)
	if err != nil {
		return err
	}
	ws, err := client.CurrentWorkspace(ctx)
	if err == nil {
		fmt.Printf("Sessão ativa: %s (%s)\nWorkspace: %s (%s)\n", profile.Email, profile.DisplayName, ws.Name, ws.Slug)
		return nil
	}
	fmt.Printf("Sessão ativa: %s (%s)\n", profile.Email, profile.DisplayName)
	return nil
}
