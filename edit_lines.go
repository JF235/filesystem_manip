package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath" // Importado para obter o nome base do programa
	"strings"
)

// Define as flags fora de main
var (
	rmpre   = flag.String("rmpre", "", "String a remover do início de cada linha")
	rmpos   = flag.String("rmpos", "", "String a remover do final de cada linha")
	addpre  = flag.String("addpre", "", "String a adicionar no início de cada linha")
	addpos  = flag.String("addpos", "", "String a adicionar no final de cada linha")
	inplace = flag.Bool("I", false, "Edita o arquivo in-place (sobrescreve o original)")
)

func main() {
	// Define a função de Usage personalizada ANTES de flag.Parse()
	flag.Usage = func() {
		output := flag.CommandLine.Output()
		progName := filepath.Base(os.Args[0])

		fmt.Fprintf(output, "%s: Edita cada linha de um arquivo adicionando/removendo prefixos/sufixos.\n\n", progName)
		fmt.Fprintf(output, "Uso: %s [opções] <arquivo>\n\n", progName)
		fmt.Fprintf(output, "Argumento:\n")
		fmt.Fprintf(output, "  <arquivo>  O caminho para o arquivo a ser processado.\n\n")
		fmt.Fprintf(output, "Opções:\n")
		// Imprime as descrições padrão de todas as flags definidas
		flag.PrintDefaults()
		fmt.Fprintf(output, "\nComportamento Padrão:\n")
		fmt.Fprintf(output, "  Por padrão, o programa exibe o conteúdo modificado no terminal (stdout).\n")
		fmt.Fprintf(output, "  O arquivo original não é alterado a menos que a opção -I seja usada.\n\n")
		fmt.Fprintf(output, "Exemplos:\n")
		fmt.Fprintf(output, "  # Adicionar '// ' no início de cada linha de config.txt (mostrar resultado)\n")
		fmt.Fprintf(output, "  %s -addpre '// ' config.txt\n\n", progName)
		fmt.Fprintf(output, "  # Remover o sufixo '.tmp' de cada linha e sobrescrever nomes.txt\n")
		fmt.Fprintf(output, "  %s -I -rmpos '.tmp' nomes.txt\n\n", progName)
		fmt.Fprintf(output, "  # Remover prefixo 'old_' e adicionar sufixo '.new' em cada linha de list.dat (mostrar resultado)\n")
		fmt.Fprintf(output, "  %s -rmpre 'old_' -addpos '.new' list.dat\n", progName)
	}

	flag.Parse()

	// Verifica se o argumento obrigatório <arquivo> foi fornecido
	if flag.NArg() < 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: O argumento <arquivo> é obrigatório.\n\n")
		flag.Usage() // Mostra a mensagem de uso completa
		os.Exit(1)   // Sai com código de erro
	}
	// Verifica se foram fornecidos argumentos extras inesperados
	if flag.NArg() > 1 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: Fornecido mais de um argumento de arquivo.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	filePath := flag.Arg(0)

	// Tenta obter informações do arquivo para permissões e verificação de existência
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Erro: Arquivo não encontrado '%s'\n", filePath)
		}
		log.Fatalf("Erro ao acessar informações do arquivo '%s': %v\n", filePath, err)
	}
	// Verifica se é um diretório (não suportado)
	if fileInfo.IsDir() {
		log.Fatalf("Erro: O caminho '%s' é um diretório, não um arquivo.\n", filePath)
	}

	// Lê o arquivo - log.Fatalf apropriado se falhar após Stat ter sucesso
	originalData, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo '%s': %v\n", filePath, err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(originalData)))
	var newLines []string
	linesProcessed := 0

	// Processa linha por linha
	for scanner.Scan() {
		line := scanner.Text()
		modifiedLine := line // Começa com a linha original

		// Aplica modificações apenas se as flags correspondentes foram fornecidas
		if *rmpre != "" {
			modifiedLine = strings.TrimPrefix(modifiedLine, *rmpre)
		}
		if *rmpos != "" {
			modifiedLine = strings.TrimSuffix(modifiedLine, *rmpos)
		}
		if *addpre != "" {
			modifiedLine = *addpre + modifiedLine
		}
		if *addpos != "" {
			modifiedLine = modifiedLine + *addpos
		}
		newLines = append(newLines, modifiedLine)
		linesProcessed++
	}

	// Verifica erros do scanner
	if scannerErr := scanner.Err(); scannerErr != nil {
		log.Fatalf("Erro durante o processamento das linhas do arquivo '%s': %v\n", filePath, scannerErr)
	}

	// Junta as linhas modificadas. O Join cuida dos newlines entre as linhas.
	// Não adiciona um \n extra no final, a menos que a última linha lida estivesse vazia
	// (preservando o comportamento do scanner/split).
	newContent := strings.Join(newLines, "\n")

	if *inplace {
		// Edição in-place: sobrescreve o arquivo original
		fmt.Printf("Modificando arquivo '%s' in-place...\n", filePath)

		// Usa as permissões originais do arquivo
		perms := fileInfo.Mode().Perm()

		// Escreve o novo conteúdo
		if err := os.WriteFile(filePath, []byte(newContent), perms); err != nil {
			log.Fatalf("Erro ao escrever modificações no arquivo '%s': %v\n", filePath, err)
		}
		fmt.Printf("Arquivo '%s' modificado com sucesso (%d linhas processadas).\n", filePath, linesProcessed)
	} else {
		// Comportamento padrão: imprime o resultado no stdout
		// Adiciona um newline final na saída do terminal se o conteúdo não estiver vazio
		// para melhor formatação no shell.
		if len(newContent) > 0 {
			fmt.Println(newContent)
		}
	}
}