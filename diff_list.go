package main

import (
	"bufio"         // Para leitura eficiente de arquivos linha por linha
	"flag"          // Para processar flags de linha de comando
	"fmt"           // Para formatação e impressão de saída
	"log"           // Para registrar erros fatais
	"os"            // Para interagir com o sistema operacional (arquivos, argumentos)
	"path/filepath" // Para obter o nome base do programa
	"strings"       // Para manipulação de strings
)

// Variáveis para as flags, definidas fora de main
var (
	countMode  = flag.Bool("count", false, "Exibir apenas a contagem de linhas diferentes, sem listar as linhas")
	pre1Flag   = flag.String("pre1", "", "Prefixo a remover das linhas do <arquivo1> antes da comparação")
	sufix1Flag = flag.String("sufix1", "", "Sufixo a remover das linhas do <arquivo1> antes da comparação")
	pre2Flag   = flag.String("pre2", "", "Prefixo a remover das linhas do <arquivo2> antes da comparação")
	sufix2Flag = flag.String("sufix2", "", "Sufixo a remover das linhas do <arquivo2> antes da comparação")
)

func main() {
	// Define a função de Usage personalizada ANTES de flag.Parse()
	flag.Usage = func() {
		output := flag.CommandLine.Output()
		progName := filepath.Base(os.Args[0])

		fmt.Fprintf(output, "%s: Compara dois arquivos de texto linha por linha e encontra linhas presentes no arquivo1 mas ausentes no arquivo2.\n", progName)
		fmt.Fprintf(output, "       Prefixos e sufixos podem ser removidos de cada linha antes da comparação.\n\n")
		fmt.Fprintf(output, "Uso: %s [opções] <arquivo1> <arquivo2>\n\n", progName)
		fmt.Fprintf(output, "Argumentos:\n")
		fmt.Fprintf(output, "  <arquivo1>  O arquivo principal cujas linhas serão verificadas.\n")
		fmt.Fprintf(output, "  <arquivo2>  O arquivo de referência contra o qual as linhas de arquivo1 serão comparadas.\n\n")
		fmt.Fprintf(output, "Opções:\n")
		// Imprime as descrições padrão de todas as flags definidas
		flag.PrintDefaults()
		fmt.Fprintf(output, "\nComportamento Padrão:\n")
		fmt.Fprintf(output, "  Por padrão, o programa exibe no terminal (stdout) cada linha (após modificações) que existe em <arquivo1> mas não em <arquivo2>.\n")
		fmt.Fprintf(output, "  Use -count para exibir apenas as contagens.\n\n")
		fmt.Fprintf(output, "Exemplo:\n")
		fmt.Fprintf(output, "  # Listar imagens de 'lista_completa.txt' que não estão em 'lista_processada.txt',\n")
		fmt.Fprintf(output, "  # removendo o prefixo 'img/' de lista_completa e o sufixo '.jpg' de lista_processada:\n")
		fmt.Fprintf(output, "  %s -pre1 'img/' -sufix2 '.jpg' lista_completa.txt lista_processada.txt\n\n", progName)
		fmt.Fprintf(output, "  # Apenas contar quantas linhas de 'fileA.log' não existem em 'fileB.log':\n")
		fmt.Fprintf(output, "  %s -count fileA.log fileB.log\n", progName)

	}

	flag.Parse() // Analisa os argumentos da linha de comando

	// Verifica se o número correto de argumentos (arquivos) foi fornecido
	args := flag.Args() // Obtém os argumentos que não são flags
	if len(args) != 2 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: São necessários exatamente dois nomes de arquivo como argumentos.\n\n")
		flag.Usage() // Mostra a mensagem de uso completa
		os.Exit(1)   // Sai com status de erro
	}
	file1Path := args[0]
	file2Path := args[1]

	// --- Leitura do Arquivo 2 ---
	// Usa um mapa para armazenar as linhas de file2 para busca rápida (O(1) em média).
	// O valor struct{} não ocupa memória adicional.
	linesFile2 := make(map[string]struct{})
	countFile2 := 0 // Contador de linhas lidas em file2

	// Abre file2
	file2, err := os.Open(file2Path)
	if err != nil {
		// log.Fatalf é apropriado para erros que impedem a execução
		log.Fatalf("Erro ao abrir o arquivo de referência '%s': %v\n", file2Path, err)
	}
	// Garante que file2 seja fechado no final de main
	// LIFO: file2 será fechado depois de file1 (se file1 for aberto com sucesso)
	defer file2.Close()

	// Lê file2 linha por linha
	scanner2 := bufio.NewScanner(file2)
	for scanner2.Scan() {
		line := scanner2.Text()
		// Modifica a linha ANTES de adicionar ao mapa
		modifiedLine := modifyLine(line, *pre2Flag, *sufix2Flag)
		linesFile2[modifiedLine] = struct{}{} // Adiciona a linha modificada ao set
		countFile2++
	}
	// Verifica erros durante a leitura de file2
	if err := scanner2.Err(); err != nil {
		log.Fatalf("Erro durante a leitura do arquivo de referência '%s': %v\n", file2Path, err)
	}

	// --- Leitura e Comparação do Arquivo 1 ---
	countFile1 := 0           // Contador de linhas lidas em file1
	missingCount := 0         // Contador de linhas de file1 não encontradas em file2
	var missingLines []string // Slice para armazenar linhas ausentes (usado apenas se !*countMode)

	// Abre file1
	file1, err := os.Open(file1Path)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo principal '%s': %v\n", file1Path, err)
	}
	// Garante que file1 seja fechado no final de main
	defer file1.Close()

	// Lê file1 linha por linha
	scanner1 := bufio.NewScanner(file1)
	for scanner1.Scan() {
		line := scanner1.Text()
		countFile1++
		// Modifica a linha ANTES da comparação
		modifiedLine := modifyLine(line, *pre1Flag, *sufix1Flag)

		// Verifica se a linha modificada existe no mapa de file2
		if _, exists := linesFile2[modifiedLine]; !exists {
			missingCount++
			// Armazena a linha ausente *apenas* se não estivermos no modo -count
			if !*countMode {
				// Armazena a linha *original* ou a *modificada*?
				// O exemplo original parecia imprimir a linha modificada. Vamos manter isso.
				// Se quisesse a original, seria append(missingLines, line)
				missingLines = append(missingLines, modifiedLine)
			}
		}
	}
	// Verifica erros durante a leitura de file1
	if err := scanner1.Err(); err != nil {
		log.Fatalf("Erro durante a leitura do arquivo principal '%s': %v\n", file1Path, err)
	}

	// --- Impressão do Resultado ---
	if *countMode {
		// Modo Contagem: Exibe estatísticas
		fmt.Printf("Linhas lidas em %s: %d\n", file1Path, countFile1)
		fmt.Printf("Linhas (únicas, após modificação) lidas em %s: %d\n", file2Path, len(linesFile2)) // len(map) dá linhas únicas
		fmt.Printf("Linhas de %s (após modificação) não encontradas em %s: %d\n", file1Path, file2Path, missingCount)
	} else {
		// Modo Padrão: Exibe as linhas ausentes
		if missingCount > 0 {
			fmt.Printf("--- Linhas de %s (após modificação) não encontradas em %s ---\n", file1Path, file2Path)
			for _, line := range missingLines {
				fmt.Println(line)
			}
			fmt.Println("-----------------------------------------------------------")
			fmt.Printf("Total de linhas ausentes: %d\n", missingCount)
		} else {
			fmt.Printf("Nenhuma linha de %s (após modificação) está ausente em %s.\n", file1Path, file2Path)
		}
	}
}

// Função auxiliar para aplicar modificações de prefixo/sufixo
func modifyLine(line, prefix, suffix string) string {
	modified := line
	if prefix != "" {
		modified = strings.TrimPrefix(modified, prefix)
	}
	if suffix != "" {
		modified = strings.TrimSuffix(modified, suffix)
	}
	return modified
}