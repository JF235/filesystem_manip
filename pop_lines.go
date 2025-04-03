package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "regexp"
    "strings"
)

func main() {
    removeMatches := flag.Bool("R", false, "Remove linhas que encaixam na expressão regular do arquivo")
    flag.Parse()

    if flag.NArg() < 2 {
        log.Fatal("Uso: pop_lines [opções] <regex> <arquivo>")
    }

    pattern := flag.Arg(0)
    filePath := flag.Arg(1)

    re, err := regexp.Compile(pattern)
    if err != nil {
        log.Fatalf("Expressão regular inválida: %v", err)
    }

    originalData, err := os.ReadFile(filePath)
    if err != nil {
        log.Fatalf("Erro ao ler arquivo '%s': %v", filePath, err)
    }

    lines := strings.Split(string(originalData), "\n")
    var keptLines []string

    // Mostra no stdout as linhas que dão match
    for _, line := range lines {
        if re.MatchString(line) {
            fmt.Println(line)
        } else {
            keptLines = append(keptLines, line)
        }
    }

    // Se -R for setado, reescreve o arquivo sem as linhas que deram match
    if *removeMatches {
        newContent := strings.Join(keptLines, "\n")
        if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
            log.Fatalf("Erro ao escrever no arquivo '%s': %v", filePath, err)
        }
    }
}