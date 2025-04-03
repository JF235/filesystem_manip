package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

func showHelp() {
    fmt.Println("Uso: extractglobal_getallfiles <diretório> [opções]")
    fmt.Println("Opções:")
    fmt.Println("  -pre <prefixo>   Listar apenas arquivos que começam com <prefixo>")
    fmt.Println("  -post <sufixo>   Listar apenas arquivos que terminam com <sufixo>")
    fmt.Println("  -r, --recursive  Pesquisar também em subdiretórios")
    fmt.Println("Exemplo:")
    fmt.Println("  extractglobal_getallfiles /path/to/dir")
    fmt.Println("  extractglobal_getallfiles /path/to/dir -pre data_")
    fmt.Println("  extractglobal_getallfiles /path/to/dir -post .png")
    fmt.Println("  extractglobal_getallfiles /path/to/dir -r")
    os.Exit(1)
}

func main() {
    var prefix, suffix string
    var recursive bool

    flag.StringVar(&prefix, "pre", "", "Listar apenas arquivos com este prefixo")
    flag.StringVar(&suffix, "post", "", "Listar apenas arquivos com este sufixo")
    // Flag sem valor para -r
    boolRecursive := flag.Bool("r", false, "Pesquisar também em subdiretórios")
    flag.Parse()

    // Pode também capturar --recursive manualmente
    // Verificar se passamos --recursive sem -r
    for _, arg := range os.Args {
        if arg == "--recursive" {
            *boolRecursive = true
        }
    }

    recursive = *boolRecursive

    // Resgatar argumentos restantes (diretório fica no primeiro)
    args := flag.Args()
    if len(args) < 1 {
        showHelp()
    }
    dirPath := args[0]

    // Verificar se o diretório existe
    info, err := os.Stat(dirPath)
    if err != nil || !info.IsDir() {
        log.Fatalf("Erro: O diretório '%s' não existe ou não é diretório.\n", dirPath)
    }

    // Remover barra final
    dirPath = strings.TrimRight(dirPath, "/")

    // Armazenar resultados
    var matchedFiles []string

    // Função para verificar prefixo e sufixo
    matches := func(name string) bool {
        return strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix)
    }

    if recursive {
        // Caminho recursivo
        filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
            if err != nil {
                return nil
            }
            if !f.IsDir() {
                filename := f.Name()
                if matches(filename) {
                    matchedFiles = append(matchedFiles, path)
                }
            }
            return nil
        })
    } else {
        // Caminho não-recursivo
        entries, err := os.ReadDir(dirPath)
        if err != nil {
            log.Fatalf("Erro ao ler o diretório '%s': %v\n", dirPath, err)
        }
        for _, entry := range entries {
            if !entry.IsDir() {
                name := entry.Name()
                if matches(name) {
                    matchedFiles = append(matchedFiles, filepath.Join(dirPath, name))
                }
            }
        }
    }

    // Ordenar resultados
    sort.Strings(matchedFiles)

    // Exibir resultados
    for _, path := range matchedFiles {
        fmt.Println(path)
    }

    os.Exit(0)
}