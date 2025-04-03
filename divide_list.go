package main

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

func main() {
    // Verifica se os argumentos foram passados corretamente
    if len(os.Args) != 4 {
        fmt.Fprintf(os.Stderr, "Uso: %s <arquivo> <num_parts> <dir_de_saida>\n", os.Args[0])
        os.Exit(1)
    }

    inputFile := os.Args[1]
    numParts, err := strconv.Atoi(os.Args[2])
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erro: num_parts inválido.\n")
        os.Exit(1)
    }
    outputDir := os.Args[3]

    // Verifica se o arquivo de entrada existe
    info, err := os.Stat(inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erro ao acessar '%s': %v\n", inputFile, err)
        os.Exit(1)
    }

    // Verifica se numParts é positivo
    if numParts < 1 {
        fmt.Fprintln(os.Stderr, "Erro: número de partes deve ser >= 1")
        os.Exit(1)
    }

    // Cria o diretório de saída, se necessário
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        fmt.Fprintf(os.Stderr, "Erro ao criar diretório '%s': %v\n", outputDir, err)
        os.Exit(1)
    }

    // Conta o total de linhas no arquivo
    var totalLines int
    {
        f, err := os.Open(inputFile)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Erro ao abrir '%s': %v\n", inputFile, err)
            os.Exit(1)
        }
        defer f.Close()

        sc := bufio.NewScanner(f)
        for sc.Scan() {
            totalLines++
        }
        if err := sc.Err(); err != nil {
            fmt.Fprintf(os.Stderr, "Erro ao ler '%s': %v\n", inputFile, err)
            os.Exit(1)
        }
    }

    // Calcula quantas linhas por arquivo
    linesPerFile := (totalLines + numParts - 1) / numParts

    // Nome base (sem extensão)
    base := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

    // Realiza a divisão
    f, err := os.Open(inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erro ao abrir '%s': %v\n", inputFile, err)
        os.Exit(1)
    }
    defer f.Close()

    sc := bufio.NewScanner(f)
    var fileCount, lineCount int

    for i := 0; i < numParts; i++ {
        outName := filepath.Join(outputDir, fmt.Sprintf("%s_parte_%d.txt", base, i+1))
        out, err := os.Create(outName)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Erro ao criar '%s': %v\n", outName, err)
            os.Exit(1)
        }

        w := bufio.NewWriter(out)
        for j := 0; j < linesPerFile && sc.Scan(); j++ {
            _, _ = w.WriteString(sc.Text() + "\n")
            lineCount++
        }
        w.Flush()
        out.Close()
        fileCount++
        if !sc.Scan() {
            break
        }
        // A linha atual já foi lida para o próximo arquivo
        _, _ = w.WriteString(sc.Text() + "\n")
        lineCount++
    }

    // Descarta possíveis sobras
    for sc.Scan() {
    }

    fmt.Printf("Divisão concluída: %d arquivos criados em '%s'.\n", fileCount, outputDir)
}