package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath" // Importado para obter o nome base do programa
	"regexp"
	"strings"
)

// Define a flag fora de main para que a descrição esteja disponível para flag.Usage
var (
	removeMatches = flag.Bool("R", false, "Remove linhas que correspondem à regex do arquivo (sobrescreve o original)")
)

func main() {
	// Define a função de Usage personalizada ANTES de flag.Parse()
	flag.Usage = func() {
		// Usa flag.CommandLine.Output() que por padrão é os.Stderr
		output := flag.CommandLine.Output()
		// Nome do executável (requer "path/filepath")
		progName := filepath.Base(os.Args[0])

		fmt.Fprintf(output, "%s: Exibe ou remove linhas de um arquivo que correspondem a uma expressão regular (regex).\n\n", progName)
		fmt.Fprintf(output, "Uso: %s [opções] <regex> <arquivo>\n\n", progName)
		fmt.Fprintf(output, "Argumentos:\n")
		fmt.Fprintf(output, "  <regex>    A expressão regular Go (estilo PCRE) para procurar nas linhas.\n")
		fmt.Fprintf(output, "             Lembre-se de usar aspas (' ') se a regex contiver espaços ou caracteres especiais do shell.\n")
		fmt.Fprintf(output, "  <arquivo>  O caminho para o arquivo a ser processado.\n\n")
		fmt.Fprintf(output, "Opções:\n")
		// Imprime as descrições padrão das flags definidas (neste caso, -R)
		flag.PrintDefaults()
		fmt.Fprintf(output, "\nComportamento Padrão:\n")
		fmt.Fprintf(output, "  Por padrão, o programa apenas exibe as linhas que correspondem à <regex> no terminal (stdout).\n")
		fmt.Fprintf(output, "  O arquivo original não é modificado a menos que a opção -R seja usada.\n\n")
		fmt.Fprintf(output, "Exemplos:\n")
		fmt.Fprintf(output, "  # Exibir todas as linhas contendo 'WARN' ou 'ERROR' em app.log\n")
		fmt.Fprintf(output, "  %s '(WARN|ERROR)' app.log\n\n", progName)
		fmt.Fprintf(output, "  # Remover todas as linhas em branco (ou que só contêm espaços) de data.txt\n")
		fmt.Fprintf(output, "  %s -R '^\\s*$' data.txt\n", progName)
	}

	flag.Parse()

	// Verifica se os argumentos obrigatórios (regex e arquivo) foram fornecidos
	if flag.NArg() < 2 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: Os argumentos <regex> e <arquivo> são obrigatórios.\n\n")
		flag.Usage() // Mostra a mensagem de uso completa
		os.Exit(1)   // Sai com código de erro
	}
	// Verifica se foram fornecidos argumentos extras inesperados
	if flag.NArg() > 2 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: Argumentos extras fornecidos após <arquivo>.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	pattern := flag.Arg(0)
	filePath := flag.Arg(1)

	// Compila a regex - log.Fatalf é apropriado aqui, pois o programa não pode continuar
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Erro: Expressão regular inválida '%s': %v\n", pattern, err)
	}

	// Lê o arquivo - log.Fatalf é apropriado aqui
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		// Verifica se o erro é "arquivo não encontrado" para uma mensagem mais específica
		if os.IsNotExist(err) {
			log.Fatalf("Erro: Arquivo não encontrado '%s'\n", filePath)
		}
		log.Fatalf("Erro ao ler arquivo '%s': %v\n", filePath, err)
	}

	lines := strings.Split(string(originalData), "\n")
	var keptLines []string // Linhas que NÃO correspondem (para usar com -R)
	var matchedLines []string // Linhas que correspondem (para exibir)
	matchCount := 0

	// Itera pelas linhas, separando as que correspondem e as que não
	for _, line := range lines {
		// Trata a última linha vazia que pode surgir do Split final
		if line == "" && len(lines) > 1 && lines[len(lines)-1] == "" && line == lines[len(lines)-1] {
			// Se for a última linha e ela estiver vazia por causa do split,
			// adiciona-a às keptLines *apenas* se a flag -R estiver ativa,
			// para preservar o newline final se ele existia no arquivo original.
			if *removeMatches {
				keptLines = append(keptLines, line)
			}
			continue
		}

		if re.MatchString(line) {
			matchedLines = append(matchedLines, line)
			matchCount++
			// Se -R NÃO estiver ativo, não adicionamos a linha às keptLines
		} else {
			keptLines = append(keptLines, line)
		}
	}

	// Exibe as linhas correspondentes (comportamento padrão)
	if matchCount > 0 {
		fmt.Println("--- Linhas Correspondentes ---")
		for _, line := range matchedLines {
			fmt.Println(line)
		}
		fmt.Println("----------------------------")
	} else {
		fmt.Println("Nenhuma linha correspondeu à expressão regular.")
	}

	// Se -R foi setado, reescreve o arquivo SEM as linhas que deram match
	if *removeMatches {
		if matchCount > 0 { // Só reescreve se houve correspondências para remover
			fmt.Printf("Removendo %d linha(s) correspondente(s) do arquivo '%s'...\n", matchCount, filePath)
			// Junta as linhas mantidas com newline
			// Atenção: Se o arquivo original não terminava com \n, este join pode adicionar um.
			// Para controle mais fino, seria necessário analisar o `originalData`
			newContent := strings.Join(keptLines, "\n")

			// Usa permissões do arquivo original se possível, senão default 0644
			fileInfo, statErr := os.Stat(filePath)
			perms := os.FileMode(0644) // Default
			if statErr == nil {
				perms = fileInfo.Mode().Perm() // Pega permissões existentes
			} else {
				log.Printf("Aviso: Não foi possível ler as permissões originais de '%s', usando 0644: %v\n", filePath, statErr)
			}

			// Escreve o novo conteúdo - log.Fatalf é apropriado aqui
			if err := os.WriteFile(filePath, []byte(newContent), perms); err != nil {
				log.Fatalf("Erro ao escrever alterações no arquivo '%s': %v\n", filePath, err)
			}
			fmt.Println("Arquivo atualizado com sucesso.")
		} else {
			fmt.Println("Nenhuma linha para remover, o arquivo permanece inalterado.")
		}
	}
}