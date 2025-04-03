package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Nenhuma flag opcional definida por enquanto, mas preparamos para o futuro.

func main() {
	// Define a função de Usage personalizada ANTES de flag.Parse()
	flag.Usage = func() {
		output := flag.CommandLine.Output()
		progName := filepath.Base(os.Args[0])

		fmt.Fprintf(output, "%s: Divide um arquivo de texto em um número especificado de partes menores.\n\n", progName)
		fmt.Fprintf(output, "Uso: %s <arquivo_entrada> <num_partes> <diretorio_saida>\n\n", progName)
		fmt.Fprintf(output, "Argumentos:\n")
		fmt.Fprintf(output, "  <arquivo_entrada>  O caminho para o arquivo de texto a ser dividido.\n")
		fmt.Fprintf(output, "  <num_partes>       O número de arquivos menores a serem criados (deve ser >= 1).\n")
		fmt.Fprintf(output, "  <diretorio_saida>  O diretório onde os arquivos divididos serão salvos.\n")
		fmt.Fprintf(output, "                     O diretório será criado se não existir.\n\n")
		fmt.Fprintf(output, "Opções:\n")
		// Imprime as opções padrão (como -h/--help)
		flag.PrintDefaults()
		fmt.Fprintf(output, "\nExemplo:\n")
		fmt.Fprintf(output, "  # Dividir 'grande_lista.txt' em 10 partes no diretório './partes':\n")
		fmt.Fprintf(output, "  %s grande_lista.txt 10 ./partes\n", progName)
	}

	// Analisa flags (como -h). Não temos flags customizadas aqui, mas mantém o padrão.
	flag.Parse()

	// Verifica se o número correto de argumentos posicionais foi fornecido
	if flag.NArg() != 3 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: Número incorreto de argumentos fornecidos.\n\n")
		flag.Usage() // Mostra a mensagem de uso completa
		os.Exit(1)   // Sai com código de erro
	}

	// Obtém os argumentos posicionais
	inputFile := flag.Arg(0)
	numPartsStr := flag.Arg(1)
	outputDir := flag.Arg(2)

	// Valida num_partes
	numParts, err := strconv.Atoi(numPartsStr)
	if err != nil {
		// Usar log.Fatalf para erros fatais simplifica o código
		log.Fatalf("Erro: <num_partes> ('%s') não é um número inteiro válido.\n", numPartsStr)
	}
	if numParts < 1 {
		log.Fatalf("Erro: <num_partes> (%d) deve ser maior ou igual a 1.\n", numParts)
	}

	// Verifica se o arquivo de entrada existe e é um arquivo
	inputFileInfo, err := os.Stat(inputFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatalf("Erro: Arquivo de entrada '%s' não encontrado.\n", inputFile)
		}
		log.Fatalf("Erro ao acessar o arquivo de entrada '%s': %v\n", inputFile, err)
	}
	if inputFileInfo.IsDir() {
		log.Fatalf("Erro: O caminho de entrada '%s' é um diretório, não um arquivo.\n", inputFile)
	}

	// Cria o diretório de saída, se necessário (ignora erro se já existir)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Erro ao criar o diretório de saída '%s': %v\n", outputDir, err)
	}

	// --- Primeira Passagem: Contar Linhas ---
	// Nota: Ler o arquivo duas vezes pode ser ineficiente para arquivos gigantescos.
	// Uma alternativa seria ler em chunks e estimar, mas a contagem exata é mais simples.
	totalLines := 0
	{ // Bloco para limitar o escopo de 'f' e 'sc' da contagem
		fileCounter, err := os.Open(inputFile)
		if err != nil {
			log.Fatalf("Erro ao abrir '%s' para contagem de linhas: %v\n", inputFile, err)
		}
		// defer fileCounter.Close() // Não precisa de defer dentro do bloco se fecharmos explicitamente

		sc := bufio.NewScanner(fileCounter)
		for sc.Scan() {
			totalLines++
		}
		// Verifica erro do scanner após o loop
		if err := sc.Err(); err != nil {
			fileCounter.Close() // Fecha antes de sair
			log.Fatalf("Erro durante a contagem de linhas em '%s': %v\n", inputFile, err)
		}
		fileCounter.Close() // Fecha o arquivo após a contagem
	} // Fim do bloco de contagem

	if totalLines == 0 && inputFileInfo.Size() > 0 {
		log.Printf("Aviso: O arquivo de entrada '%s' não está vazio, mas nenhuma linha foi contada (verifique o formato).\n", inputFile)
	} else if totalLines == 0 {
		log.Printf("Aviso: O arquivo de entrada '%s' está vazio ou não contém linhas.\n", inputFile)
		// Mesmo com 0 linhas, pode-se querer criar arquivos vazios. Continuamos.
	}

	// Calcula quantas linhas por arquivo de saída (arredondando para cima)
	linesPerFile := 0
	if totalLines > 0 { // Evita divisão por zero se numParts for grande e totalLines for 0
		linesPerFile = (totalLines + numParts - 1) / numParts
	}
	if linesPerFile == 0 && totalLines > 0 {
		linesPerFile = 1 // Garante pelo menos 1 se houver linhas e muitas partes
	}

	log.Printf("Arquivo de entrada: '%s' (%d linhas)\n", inputFile, totalLines)
	log.Printf("Dividindo em %d partes (aprox. %d linhas por parte) no diretório '%s'\n", numParts, linesPerFile, outputDir)

	// --- Segunda Passagem: Dividir e Escrever ---
	fileSplitter, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Erro ao reabrir '%s' para divisão: %v\n", inputFile, err)
	}
	defer fileSplitter.Close() // Garante fechamento no final de main

	scanner := bufio.NewScanner(fileSplitter)
	filesCreated := 0
	baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	// Loop para criar cada arquivo de parte
	for i := 0; i < numParts; i++ {
		// Define o nome do arquivo de saída
		outputFileName := filepath.Join(outputDir, fmt.Sprintf("%s_parte_%0*d.txt", baseName, len(strconv.Itoa(numParts)), i+1)) // Adiciona padding ao número

		// Cria o arquivo de saída
		outFile, err := os.Create(outputFileName)
		if err != nil {
			// Loga o erro mas tenta continuar se possível? Ou falha tudo? Vamos falhar tudo.
			log.Fatalf("Erro ao criar o arquivo de saída '%s': %v\n", outputFileName, err)
		}

		// Usa bufio.Writer para escrita eficiente
		writer := bufio.NewWriter(outFile)
		linesWrittenInCurrentFile := 0

		// Escreve linesPerFile linhas (ou até o fim do arquivo de entrada)
		for j := 0; j < linesPerFile; j++ {
			// Tenta ler a próxima linha do arquivo de entrada
			if !scanner.Scan() {
				// Se não há mais linhas, sai do loop interno (e potencialmente do externo)
				goto EndOfInput // Usamos goto para sair de ambos os loops se a entrada acabar
			}
			// Escreve a linha lida no arquivo de saída atual
			_, err := writer.WriteString(scanner.Text() + "\n")
			if err != nil {
				outFile.Close() // Tenta fechar antes de sair
				log.Fatalf("Erro ao escrever no arquivo de saída '%s': %v\n", outputFileName, err)
			}
			linesWrittenInCurrentFile++
		}

		// Garante que tudo foi escrito para o disco para este arquivo
		if err := writer.Flush(); err != nil {
			outFile.Close()
			log.Fatalf("Erro ao fazer flush no arquivo de saída '%s': %v\n", outputFileName, err)
		}
		// Fecha o arquivo de saída atual
		if err := outFile.Close(); err != nil {
			log.Fatalf("Erro ao fechar o arquivo de saída '%s': %v\n", outputFileName, err)
		}

		// Incrementa o contador de arquivos criados *somente* se linhas foram escritas
		// ou se for um dos numParts e totalLines for 0 (para criar arquivos vazios)
		if linesWrittenInCurrentFile > 0 || totalLines == 0 {
			filesCreated++
		}
	}

EndOfInput: // Label para onde pular quando a entrada acabar

	// Verifica erro final do scanner da segunda passagem
	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro durante a leitura de '%s' na segunda passagem: %v\n", inputFile, err)
	}

	log.Printf("Divisão concluída. %d arquivos criados em '%s'.\n", filesCreated, outputDir)
}