package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	cli "github.com/widia-io/widia-omni/internal/cli"
)

const defaultAPIURL = "http://localhost:8080"

func main() {
	ctx := context.Background()

	apiURL := firstNonEmpty(os.Getenv("WIDIA_API_URL"), defaultAPIURL)
	sessionPath, err := cli.DefaultSessionPath()
	if err != nil {
		fmt.Printf("Não foi possível determinar o caminho de sessão: %v\n", err)
		os.Exit(1)
	}

	session, err := cli.NewSessionStore(sessionPath)
	if err != nil {
		fmt.Printf("Erro ao carregar sessão local: %v\n", err)
		os.Exit(1)
	}
	session.SetBaseURL(apiURL)

	client := cli.NewClient(session)

	help := flag.Bool("help", false, "exibe ajuda")
	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()

	if *help {
		printUsage()
		return
	}

	if len(args) > 0 {
		if err := handleCommand(ctx, client, args); err != nil {
			fmt.Println("Erro:", err)
			os.Exit(1)
		}
		return
	}

	if err := cli.RunInteractive(ctx, bufio.NewReader(os.Stdin), client); err != nil {
		fmt.Println("Erro:", err)
		os.Exit(1)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func handleCommand(ctx context.Context, client *cli.Client, args []string) error {
	switch args[0] {
	case "login":
		return cli.RunInteractiveLogin(ctx, bufio.NewReader(os.Stdin), client)
	case "logout":
		return cli.Logout(ctx, client)
	case "status":
		return cli.ShowStatus(ctx, client)
	default:
		printUsage()
		return fmt.Errorf("comando desconhecido: %s", args[0])
	}
}

func printUsage() {
	fmt.Println("Uso:")
	fmt.Println("  widia               abre a interface interativa")
	fmt.Println("  widia login         entrar no sistema")
	fmt.Println("  widia logout        sair da sessao")
	fmt.Println("  widia status        checar autenticacao atual")
	fmt.Println("  widia -help         exibir esta mensagem")
	fmt.Println()
	fmt.Println("Configuracao:")
	fmt.Println("  WIDIA_API_URL       URL da API (padrao:", defaultAPIURL, ")")
}
