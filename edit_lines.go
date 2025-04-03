package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
    "strings"
)

func main() {
    rmpre := flag.String("rmpre", "", "String a remover do início de cada linha")
    rmpos := flag.String("rmpos", "", "String a remover do final de cada linha")
    addpre := flag.String("addpre", "", "String a adicionar no início de cada linha")
    addpos := flag.String("addpos", "", "String a adicionar no final de cada linha")
    inplace := flag.Bool("I", false, "Edição in-place (sobrescrever arquivo)")
    flag.Parse()

    if flag.NArg() < 1 {
        log.Fatal("Uso: edit_lines [opções] <arquivo>")
    }
    filePath := flag.Arg(0)

    originalContent, err := os.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Erro ao ler arquivo: %v", err)
    }

    scanner := bufio.NewScanner(strings.NewReader(string(originalContent)))
    var newLines []string
    for scanner.Scan() {
        line := scanner.Text()
        line = strings.TrimPrefix(line, *rmpre)
        line = strings.TrimSuffix(line, *rmpos)
        line = *addpre + line + *addpos
        newLines = append(newLines, line)
    }

    if scannerErr := scanner.Err(); scannerErr != nil {
        log.Fatalf("Erro ao processar linhas: %v", scannerErr)
    }

    newContent := strings.Join(newLines, "\n") + "\n"

    if *inplace {
        // Loga o conteúdo antigo no stdout
        fmt.Print(string(originalContent))

        // Substitui o arquivo com o novo conteúdo
        if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
            log.Fatalf("Erro ao escrever arquivo: %v", err)
        }
    } else {
        fmt.Print(newContent)
    }
}