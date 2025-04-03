package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Definindo as flags no escopo do pacote para serem acessíveis na função Usage
var (
	rmpre   = flag.String("rmpre", "", "String a remover do início do nome de arquivo")
	rmpos   = flag.String("rmpos", "", "String a remover do final do nome de arquivo")
	addpre  = flag.String("addpre", "", "String a adicionar no início do nome de arquivo")
	addpos  = flag.String("addpos", "", "String a adicionar no final do nome de arquivo")
	inplace = flag.Bool("I", false, "Renomear in-place (sobrescreve o arquivo antigo)")
	dirMode = flag.String("dir", "", "Se especificado, percorre todo este `diretório` para renomear arquivos")
)

func main() {
	// Define a função de Usage personalizada ANTES de flag.Parse()
	flag.Usage = func() {
		// Usa flag.CommandLine.Output() que por padrão é os.Stderr
		output := flag.CommandLine.Output()

		// Nome do executável
		progName := filepath.Base(os.Args[0])

		fmt.Fprintf(output, "%s: Renomeia arquivos removendo/adicionando prefixos/sufixos.\n\n", progName)
		fmt.Fprintf(output, "Uso:\n")
		fmt.Fprintf(output, "  1. %s [opções] <arquivo1> [arquivo2...]\n", progName)
		fmt.Fprintf(output, "  2. %s -dir <diretório> [opções]\n\n", progName)
		fmt.Fprintf(output, "Opções:\n")

		// Imprime as descrições padrão das flags definidas
		flag.PrintDefaults()
		fmt.Fprintf(output, "\nExemplos:\n")
		fmt.Fprintf(output, "  %s -rmpre 'temp_' -addpos '.bkp' arquivo1.txt\n", progName)
		fmt.Fprintf(output, "     (mostra: arquivo1.txt -> arquivo1.txt.bkp)\n")
		fmt.Fprintf(output, "  %s -I -rmpre 'draft-' -dir ./documentos\n", progName)
		fmt.Fprintf(output, "     (renomeia todos os arquivos em ./documentos que começam com 'draft-', removendo o prefixo)\n")
	}

	flag.Parse()

	// Verifica se a combinação de argumentos é válida
	// Precisa de um diretório (-dir) OU de pelo menos um arquivo como argumento
	if *dirMode == "" && flag.NArg() == 0 {
		fmt.Fprintf(flag.CommandLine.Output(), "Erro: Nenhum arquivo ou diretório especificado.\n\n")
		flag.Usage() // Mostra a mensagem de uso completa
		os.Exit(1)   // Sai com código de erro
	}

	// Se a flag -dir foi fornecida, percorre o diretório
	if *dirMode != "" {
		info, err := os.Stat(*dirMode)
		if err != nil || !info.IsDir() {
			// Mantém log.Fatalf aqui pois é um erro fatal específico da operação
			log.Fatalf("Erro: O caminho especificado em -dir '%s' não é um diretório válido ou acessível.\n", *dirMode)
		}

		fmt.Printf("Percorrendo diretório: %s\n", *dirMode)
		err = filepath.Walk(*dirMode, func(path string, f os.FileInfo, err error) error {
			if err != nil {
				log.Printf("Aviso: Erro ao acessar '%s', pulando: %v\n", path, err)
				return nil // Continua a percorrer outros arquivos/subdiretórios
			}
			// Processa apenas se for um arquivo e não o próprio diretório raiz percorrido
			if !f.IsDir() && path != *dirMode {
				renameFile(path, *rmpre, *rmpos, *addpre, *addpos, *inplace)
			}
			return nil
		})
		if err != nil {
			// Erro durante o Walk (raro se os erros individuais forem tratados)
			log.Fatalf("Erro fatal ao percorrer o diretório '%s': %v", *dirMode, err)
		}
		fmt.Println("Processamento do diretório concluído.")
		return // Termina a execução após processar o diretório
	}

	// Caso contrário (não usou -dir), processa os arquivos passados diretamente
	fmt.Println("Processando arquivos individuais:")
	for _, oldPath := range flag.Args() {
		// Verifica se o arquivo existe antes de tentar renomear
		if _, err := os.Stat(oldPath); os.IsNotExist(err) {
			log.Printf("Erro: Arquivo '%s' não encontrado.\n", oldPath)
			continue // Pula para o próximo arquivo
		}
		renameFile(oldPath, *rmpre, *rmpos, *addpre, *addpos, *inplace)
	}
	fmt.Println("Processamento de arquivos individuais concluído.")
}

func renameFile(oldPath, rmpre, rmpos, addpre, addpos string, inplace bool) {
	dir := filepath.Dir(oldPath)
	base := filepath.Base(oldPath)

	// Aplica as transformações apenas na parte do nome (base)
	newName := base
	if rmpre != "" {
		newName = strings.TrimPrefix(newName, rmpre)
	}
	if rmpos != "" {
		newName = strings.TrimSuffix(newName, rmpos)
	}
	if addpre != "" {
		newName = addpre + newName
	}
	if addpos != "" {
		newName = newName + addpos
	}

	// Se o nome não mudou, não faz nada
	if newName == base {
		// log.Printf("Info: Nome de '%s' não alterado pelas regras.\n", oldPath)
		return
	}

	newPath := filepath.Join(dir, newName)

	if inplace {
		if err := os.Rename(oldPath, newPath); err != nil {
			// Usa log.Printf para erros não fatais durante o processo
			log.Printf("Erro ao renomear '%s' para '%s': %v\n", oldPath, newPath, err)
		} else {
			fmt.Printf("Renomeado: %s -> %s\n", oldPath, newPath)
		}
	} else {
		// Apenas mostra o que seria feito
		fmt.Printf("Simulação: %s -> %s\n", oldPath, newPath)
	}
}