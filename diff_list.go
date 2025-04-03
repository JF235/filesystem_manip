package main

import (
	"bufio" 	// Para leitura eficiente de arquivos linha por linha
	"flag"  	// Para processar flags de linha de comando como -count
	"fmt"   	// Para formatação e impressão de saída
	"log"   	// Para registrar erros fatais
	"os"    	// Para interagir com o sistema operacional (arquivos, argumentos)
	"strings" 	// Para manipulação de strings
)

// 00000/0/0_041A5F35-2ECD-4DB9-867A-952162F4E732_6_1600872760306.wsq
// extracted/00000/0/0_041A5F35-2ECD-4DB9-867A-952162F4E732_6_1600872760306.tpt
// ./list_diff -count /storage/fingerprints/multibio/multibio_images.list /storage/fingerprints/multibio/multibio_tpt.list -sufix1 .tpt -pre2 extracted/ -sufix2 .wsq

// Compara o conteúdo em dois arquivos de texto

func main() {
	// 1. Definir e analisar a flag -count
	countMode := flag.Bool("count", false, "Exibir apenas a contagem de linhas diferentes")
	pre1Flag := flag.String("pre1", "", "Prefixo a remover das linhas do arquivo1")
    sufix1Flag := flag.String("sufix1", "", "Sufixo a remover das linhas do arquivo1")
    pre2Flag := flag.String("pre2", "", "Prefixo a remover das linhas do arquivo2")
    sufix2Flag := flag.String("sufix2", "", "Sufixo a remover das linhas do arquivo2")

	flag.Parse() // Analisa os argumentos da linha de comando

	// 2. Obter os nomes dos arquivos dos argumentos restantes
	args := flag.Args() // Obtém os argumentos que não são flags
	if len(args) != 2 {
		// Exibe mensagem de uso e sai se o número de argumentos estiver incorreto
		fmt.Fprintf(os.Stderr, "Erro: São necessários exatamente dois nomes de arquivo.\n")
		fmt.Fprintf(os.Stderr, "Uso: %s [-count] <arquivo1> <arquivo2>\n", os.Args[0])
		os.Exit(1) // Sai com status de erro
	}
	file1Path := args[0]
	file2Path := args[1]

	// 3. Ler todas as linhas do arquivo2 em um mapa (para busca eficiente)
	linesFile2 := make(map[string]struct{}) // Usamos struct{} como valor para economizar memória (set)
	countFile2 := 0

	file2, err := os.Open(file2Path)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo %s: %v", file2Path, err)
	}
	defer file2.Close() // Garante que o arquivo seja fechado ao final da função main


    scanner2 := bufio.NewScanner(file2)
    for scanner2.Scan() {
        line := scanner2.Text()
        // Remove prefixo e sufixo de cada linha do arquivo2
        line = strings.TrimPrefix(line, *pre2Flag)
        line = strings.TrimSuffix(line, *sufix2Flag)
        linesFile2[line] = struct{}{}
        countFile2++
    }
    if err := scanner2.Err(); err != nil {
        log.Fatalf("Erro ao ler o arquivo %s: %v", file2Path, err)
    }

	// 4. Ler o arquivo1 linha por linha e verificar a existência no mapa do arquivo2
	countFile1 := 0
	missingCount := 0
	var missingLines []string // Slice para armazenar as linhas ausentes (apenas se não for -count)

	file1, err := os.Open(file1Path)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo %s: %v", file1Path, err)
	}
	defer file1.Close() // Garante que o arquivo seja fechado

    scanner1 := bufio.NewScanner(file1)
    for scanner1.Scan() {
        line := scanner1.Text()
        // Remove prefixo e sufixo de cada linha do arquivo1
        line = strings.TrimPrefix(line, *pre1Flag)
        line = strings.TrimSuffix(line, *sufix1Flag)

        countFile1++
        if _, exists := linesFile2[line]; !exists {
            missingCount++
            if !*countMode {
                missingLines = append(missingLines, line)
            }
        }
    }
    if err := scanner1.Err(); err != nil {
        log.Fatalf("Erro ao ler o arquivo %s: %v", file1Path, err)
    }

	// 5. Imprimir o resultado com base na flag -count
	if *countMode {
		fmt.Printf("lines in %s: %d\n", file1Path, countFile1)
		fmt.Printf("lines in %s: %d\n", file2Path, countFile2)
		// Renomeado para clareza na saída, conforme solicitado
		fmt.Printf("lines in %s missing in %s: %d\n", file1Path, file2Path, missingCount)
	} else {
		// Imprime cada linha de arquivo1 que não foi encontrada em arquivo2
		for _, line := range missingLines {
			fmt.Println(line)
		}
	}
}


