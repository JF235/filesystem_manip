package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

const (
    binDir      = "bin"
    bashrcFile  = ".bashrc"
    pathComment = "# fsgo" // <comentário> (<date>) será montado na hora
)

var (
    // Flags de linha de comando
    buildAll  = flag.Bool("buildAll", false, "Compila todos os arquivos .go da raiz e os coloca em ./bin")
    setupPath = flag.Bool("setupPath", false, "Adiciona ./bin ao PATH no ~/.bashrc (se não existir)")
    listFuncs = flag.Bool("list", false, "Lista os arquivos .go na raiz e suas funções declaradas")
)

// Guarda o nome do arquivo fonte deste organizador
var organizerSourceName string

func main() {
    log.SetFlags(0)
    determineOrganizerSourceName()

    flag.Usage = usage
    flag.Parse()

    // Nenhuma flag válida? Mostra usage e sai.
    if !*buildAll && !*setupPath && !*listFuncs {
        if flag.NFlag() > 0 {
            log.Printf("Erro: Flag(s) desconhecida(s) fornecida(s) ou nenhuma ação válida especificada.")
        } else {
            log.Println("Nenhuma ação especificada.")
        }
        flag.Usage()
        os.Exit(1)
    }

    anyError := false

    if *listFuncs {
        if err := runListFunctions(); err != nil {
            log.Printf("ERRO durante -list: %v", err)
            anyError = true
        }
        log.Println("---------")
    }

    if *buildAll {
        log.Println("--- Iniciando compilação dos arquivos .go da raiz ---")
        if err := runBuildAll(); err != nil {
            log.Printf("ERRO durante -buildAll: %v", err)
            anyError = true
        } else {
            log.Println("--- Compilação concluída ---")
        }
    }

    if *setupPath {
        log.Println("--- Configurando PATH no ~/.bashrc ---")
        if err := runSetupPath(); err != nil {
            log.Printf("ERRO durante -setupPath: %v", err)
            anyError = true
        } else {
            log.Println("--- Configuração do PATH concluída ---")
        }
    }

    if anyError {
        log.Println("\nAVISO: Uma ou mais operações falharam.")
        os.Exit(1)
    }
}

// -------------------- Helpers gerais --------------------

func usage() {
    output := flag.CommandLine.Output()
    progName := filepath.Base(os.Args[0])
    if progName == "." || progName == "main" {
        progName = "manip_organize"
    }

    fmt.Fprintf(output, "%s: Ferramenta para compilar, configurar e listar funções do projeto filesystem‑manip.\n\n", progName)
    fmt.Fprintf(output, "Uso: %s [flags]\n\n", progName)
    fmt.Fprintf(output, "IMPORTANTE (sobre -buildAll): Este script compila cada arquivo .go na raiz individualmente.\n")
    fmt.Fprintf(output, "            Ter múltiplos 'package main' na mesma pasta pode causar avisos em IDEs ('main redeclared').\n\n")
    fmt.Fprintf(output, "Flags disponíveis:\n")
    flag.PrintDefaults()
    fmt.Fprintf(output, "\nExemplos:\n")
    fmt.Fprintf(output, "  ./%s -buildAll           # Compila *.go para ./bin\n", progName)
    fmt.Fprintf(output, "  ./%s -setupPath          # Adiciona ./bin ao PATH no ~/.bashrc\n", progName)
    fmt.Fprintf(output, "  ./%s -list               # Lista arquivos .go e suas funções\n", progName)
    fmt.Fprintf(output, "  ./%s -buildAll -setupPath # Compila e configura o PATH\n", progName)
}

// Determina o nome do arquivo .go que contém este código para ignorá‑lo
func determineOrganizerSourceName() {
    if exePath, err := os.Executable(); err == nil {
        baseExe := filepath.Base(exePath)
        potential := baseExe + ".go"
        if _, err := os.Stat(potential); err == nil {
            organizerSourceName = potential
            return
        }
    }

    arg0 := filepath.Base(os.Args[0])
    if strings.HasSuffix(arg0, ".go") {
        if _, err := os.Stat(arg0); err == nil {
            organizerSourceName = arg0
            return
        }
    } else {
        potential := arg0 + ".go"
        if _, err := os.Stat(potential); err == nil {
            organizerSourceName = potential
            return
        }
    }

    organizerSourceName = "manip_organize.go"
}

func getProjectRoot() (string, error) {
    exePath, err := os.Executable()
    if err != nil {
        return "", fmt.Errorf("erro ao obter caminho do executável: %w", err)
    }
    return filepath.Dir(filepath.Dir(exePath)), nil
}

// -------------------- -list --------------------

func runListFunctions() error {
    rootDir, err := getProjectRoot()
    if err != nil {
        return fmt.Errorf("erro ao determinar diretório raiz: %w", err)
    }

    entries, err := os.ReadDir(rootDir)
    if err != nil {
        return fmt.Errorf("erro ao ler diretório raiz: %w", err)
    }

    goFilesListed := 0
    log.Println("---------")
    for _, entry := range entries {
        fileName := entry.Name()
        if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || fileName == organizerSourceName {
            continue
        }
        goFilesListed++
        fmt.Println(fileName)
    }

    if goFilesListed == 0 {
        log.Printf("Nenhum arquivo .go encontrado em '%s' (ignorando o organizador).", rootDir)
    }
    return nil
}

// -------------------- -buildAll --------------------

func runBuildAll() error {
    absBinDir, err := filepath.Abs(binDir)
    if err != nil {
        return fmt.Errorf("não foi possível resolver caminho absoluto de '%s': %w", binDir, err)
    }

    log.Printf("Garantindo que o diretório de saída '%s' existe...", absBinDir)
    if err := os.MkdirAll(absBinDir, 0755); err != nil {
        return fmt.Errorf("erro ao criar diretório '%s': %w", absBinDir, err)
    }

    entries, err := os.ReadDir(".")
    if err != nil {
        return fmt.Errorf("erro ao ler diretório atual: %w", err)
    }

    buildErrors := false
    builtCount := 0
    goFilesFound := 0

    log.Printf("Procurando por arquivos .go na raiz '%s' para compilar (ignorando '%s')...", filepath.Dir(absBinDir), organizerSourceName)

    for _, entry := range entries {
        fileName := entry.Name()
        if entry.IsDir() || !strings.HasSuffix(fileName, ".go") {
            continue
        }
        goFilesFound++
        if fileName == organizerSourceName {
            continue
        }

        programName := strings.TrimSuffix(fileName, ".go")
        outputPath := filepath.Join(absBinDir, programName)

        log.Printf("Compilando '%s' → '%s'...", fileName, outputPath)
        cmd := exec.Command("go", "build", "-o", outputPath, fileName)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        if err := cmd.Run(); err != nil {
            log.Printf("ERRO ao compilar '%s': %v", fileName, err)
            buildErrors = true
        } else {
            log.Printf("✔ Compilado com sucesso: %s", outputPath)
            builtCount++
        }
    }

    if goFilesFound == 0 {
        log.Printf("Nenhum arquivo .go encontrado no diretório '%s' (além do organizador).", filepath.Dir(absBinDir))
    }

    if buildErrors {
        return fmt.Errorf("%d arquivo(s) compilado(s) com sucesso, mas ocorreram erros em outros", builtCount)
    }

    if builtCount > 0 {
        log.Printf("Executáveis gerados em '%s'.", absBinDir)
    }
    return nil
}

// -------------------- -setupPath --------------------

func runSetupPath() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("não foi possível obter o diretório home do usuário: %w", err)
    }
    bashrcPath := filepath.Join(homeDir, bashrcFile)

    rootDir, err := os.Getwd()
    if err != nil {
        return fmt.Errorf("não foi possível obter o diretório de trabalho atual: %w", err)
    }
    absBinPath, err := filepath.Abs(filepath.Join(rootDir, binDir))
    if err != nil {
        return fmt.Errorf("não foi possível obter o caminho absoluto para '%s': %w", binDir, err)
    }

    if _, err := os.Stat(absBinPath); os.IsNotExist(err) {
        log.Printf("Aviso: O diretório '%s' não existe. Criando para adicionar ao PATH.", absBinPath)
        if err := os.MkdirAll(absBinPath, 0755); err != nil {
            return fmt.Errorf("erro ao criar diretório '%s': %w", absBinPath, err)
        }
    }

    exportLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", absBinPath)
    entryExists, err := bashrcContainsLine(bashrcPath, exportLine)
    if err != nil {
        return err
    }

    if entryExists {
        log.Printf("O caminho '%s' já está configurado em '%s'. Nenhuma alteração feita.", absBinPath, bashrcPath)
        return nil
    }

    timestamp := time.Now().Format(time.RFC1123)
    lineToAdd := fmt.Sprintf("\n%s (%s)\n%s\n", pathComment, timestamp, exportLine)

    if err := appendToFile(bashrcPath, lineToAdd); err != nil {
        return err
    }

    log.Printf("Adicionado '%s' ao PATH em '%s'.", absBinPath, bashrcPath)
    log.Printf("Linha adicionada:\n%s (%s)\n%s", pathComment, timestamp, exportLine)
    log.Println("IMPORTANTE: Para aplicar as mudanças, reinicie seu terminal ou execute: source", bashrcPath)
    return nil
}

func bashrcContainsLine(path, target string) (bool, error) {
    file, err := os.Open(path)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil // .bashrc ainda não existe
        }
        return false, fmt.Errorf("erro ao abrir '%s': %w", path, err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if strings.Contains(scanner.Text(), target) || strings.HasPrefix(scanner.Text(), pathComment) {
            return true, nil
        }
    }
    if err := scanner.Err(); err != nil {
        return false, fmt.Errorf("erro ao ler '%s': %w", path, err)
    }
    return false, nil
}

func appendToFile(path, content string) error {
    file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
    if err != nil {
        return fmt.Errorf("erro ao abrir/criar '%s': %w", path, err)
    }
    defer file.Close()

    if _, err := file.WriteString(content); err != nil {
        return fmt.Errorf("erro ao escrever em '%s': %w", path, err)
    }
    return nil
}
