package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	cli "github.com/widia-io/widia-omni/internal/cli"
)

const defaultAPIURL = "https://api.meufoco.app"
const defaultVersion = "desenvolvimento"

var (
	version = defaultVersion
	commit  = ""
	date    = ""
)

func main() {
	ctx := context.Background()

	apiURL := firstNonEmpty(
		os.Getenv("MEUFOCO_API_URL"),
		os.Getenv("WIDIA_API_URL"),
		defaultAPIURL,
	)
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
	versionFlag := flag.Bool("version", false, "exibe versão do CLI")
	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()

	if *help {
		printUsage()
		return
	}
	if *versionFlag {
		fmt.Println(fullVersion())
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
	case "version":
		fmt.Println(fullVersion())
		return nil
	default:
		printUsage()
		return fmt.Errorf("comando desconhecido: %s", args[0])
	}
}

func fullVersion() string {
	buildDate := date
	if buildDate == "" {
		buildDate = "local"
	}
	gitCommit := "local"
	if commit != "" {
		gitCommit = commit
	}
	return fmt.Sprintf("meufoco-cli %s (%s) em %s", version, gitCommit, buildDate)
}

func printUsage() {
	fmt.Println("meufoco-cli", fullVersion())
	fmt.Println("Uso:")
	fmt.Println("  meufoco               abre a interface interativa")
	fmt.Println("  meufoco login         entrar no sistema")
	fmt.Println("  meufoco logout        sair da sessao")
	fmt.Println("  meufoco status        checar autenticacao atual")
	fmt.Println("  meufoco version       ver informacoes da versao")
	fmt.Println("  meufoco -help         exibir esta mensagem")
	fmt.Println("  meufoco -version      exibir versao do binary")
	fmt.Println()
	fmt.Println("Configuracao:")
	fmt.Println("  MEUFOCO_API_URL     URL da API (padrao:", defaultAPIURL, ")")
}
