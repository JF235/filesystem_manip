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
	pathComment = "# Added by manip_organize"
)

var (
	// Flags
	buildAll  = flag.Bool("buildAll", false, "Compila todos os arquivos .go da raiz e os coloca em ./bin")
	setupPath = flag.Bool("setupPath", false, "Adiciona ./bin ao PATH no ~/.bashrc (se não existir)")
	listFuncs = flag.Bool("list", false, "Lista os arquivos .go na raiz e suas funções declaradas") // Nova flag
)

// Variável global para guardar o nome do arquivo fonte deste organizador
var organizerSourceName string

func main() {
	// --- Configuração Inicial e Flags ---
	log.SetFlags(0)

	// Determina o nome do arquivo fonte do organizador ANTES de definir Usage
	determineOrganizerSourceName()

	flag.Usage = func() {
		output := flag.CommandLine.Output()
		progName := filepath.Base(os.Args[0])
		if progName == "." || progName == "main" {
			progName = "manip_organize" // Nome pretendido
		}

		fmt.Fprintf(output, "%s: Ferramenta para compilar, configurar e listar funções do projeto filesystem-manip.\n\n", progName)
		fmt.Fprintf(output, "Uso: %s [flags]\n\n", progName)
		fmt.Fprintf(output, "IMPORTANTE (sobre -buildAll): Este script compila cada arquivo .go na raiz individualmente.\n")
		fmt.Fprintf(output, "            Ter múltiplos 'package main' na mesma pasta pode causar avisos em IDEs ('main redeclared').\n\n")
		fmt.Fprintf(output, "Flags disponíveis:\n")
		flag.PrintDefaults() // Incluirá -list automaticamente
		fmt.Fprintf(output, "\nExemplos:\n")
		fmt.Fprintf(output, "  ./%s -buildAll           # Compila *.go para ./bin\n", progName)
		fmt.Fprintf(output, "  ./%s -setupPath          # Adiciona ./bin ao PATH no ~/.bashrc\n", progName)
		fmt.Fprintf(output, "  ./%s -list               # Lista arquivos .go e suas funções\n", progName) // Novo exemplo
		fmt.Fprintf(output, "  ./%s -buildAll -setupPath # Compila e configura o PATH\n", progName)
	}

	flag.Parse()

	// Verifica se alguma flag *conhecida* foi passada
	if !*buildAll && !*setupPath && !*listFuncs {
		if flag.NFlag() > 0 { // NFlag conta quantas flags foram setadas na linha de comando
			log.Printf("Erro: Flag(s) desconhecida(s) fornecida(s) ou nenhuma ação válida especificada.")
		} else {
			log.Println("Nenhuma ação especificada.")
		}
		flag.Usage()
		os.Exit(1)
	}

	// --- Execução das Ações ---
	anyError := false

	// Executa -list primeiro se solicitado, pois é apenas informativo
	if *listFuncs {
		if err := runListFunctions(); err != nil {
			log.Printf("ERRO durante -list: %v", err)
			anyError = true
			// Continua para outras ações se solicitadas
		} else {
			log.Println("---------")
		}
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

// --- Funções Auxiliares ---

// Determina o nome do arquivo .go que contém o código deste programa
func determineOrganizerSourceName() {
	// Tenta obter o nome do executável
	exePath, err := os.Executable()
	if err == nil {
		// Se o executável tem o mesmo nome base de um arquivo .go na raiz, assume esse .go
		baseExe := filepath.Base(exePath)
		potentialSource := baseExe + ".go"
		if _, statErr := os.Stat(potentialSource); statErr == nil {
			organizerSourceName = potentialSource
			return
		}
	}

	// Fallback: Tenta pelo os.Args[0] (menos confiável, especialmente com 'go run')
	arg0 := filepath.Base(os.Args[0])
	if strings.HasSuffix(arg0, ".go") {
		if _, statErr := os.Stat(arg0); statErr == nil {
			organizerSourceName = arg0
			return
		}
	} else {
		potentialSource := arg0 + ".go"
		if _, statErr := os.Stat(potentialSource); statErr == nil {
			organizerSourceName = potentialSource
			return
		}
	}

	// Último recurso: Assume um nome padrão
	organizerSourceName = "manip_organize.go" // Ou o nome que você salvou o arquivo
	// log.Printf("Aviso: Não foi possível determinar o nome exato do arquivo fonte do organizador, assumindo '%s'", organizerSourceName)
}


func getProjectRoot() (string, error) {
    exePath, err := os.Executable()
    if err != nil {
        return "", fmt.Errorf("erro ao obter caminho do executável: %w", err)
    }
    
    // O diretório raiz é o diretório pai do executável
    return filepath.Dir(filepath.Dir(exePath)), nil
}

// --- Lógica para -list (Nova) ---

func runListFunctions() error {

	rootDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("erro ao determinar diretório raiz: %w", err)
	}

	entries ,err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("erro ao ler diretório raiz: %w", err)
	}

	listErrors := false
	goFilesListed := 0
	
	log.Println("---------")
	for _, entry := range entries {
		fileName := entry.Name()
		// Pula diretórios, não-go e o próprio organizador
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || fileName == organizerSourceName {
			continue
		}

		goFilesListed++
		fmt.Printf("%s\n", fileName)
	}

	if goFilesListed == 0 {
		log.Println("Nenhum arquivo .go encontrado na raiz para listar (ignorando o organizador).")
	}

	if listErrors {
		return fmt.Errorf("ocorreram erros ao analisar alguns arquivos .go")
	}

	return nil
}

// --- Lógica para -buildAll (Adaptada para ignorar organizador) ---

func runBuildAll() error {
	log.Printf("Garantindo que o diretório '%s' existe...", binDir)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório '%s': %w", binDir, err)
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("erro ao ler diretório atual: %w", err)
	}

	buildErrors := false
	builtCount := 0
	goFilesFound := 0

	log.Printf("Procurando por arquivos .go na raiz para compilar (ignorando '%s')...", organizerSourceName)

	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") {
			continue
		}
		goFilesFound++
		if fileName == organizerSourceName {
			// log.Printf("   Ignorando o próprio arquivo do organizador: %s", fileName) // Log mais verboso
			continue
		}

		programName := strings.TrimSuffix(fileName, ".go")
		sourcePath := fileName
		outputPath := filepath.Join(binDir, programName)

		log.Printf("Compilando '%s'...", sourcePath)
		cmd := exec.Command("go", "build", "-o", outputPath, sourcePath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Printf("ERRO ao compilar '%s': %v", sourcePath, err)
			buildErrors = true
		} else {
			log.Printf(" -> Compilado com sucesso para '%s'", outputPath)
			builtCount++
		}
	}

	if goFilesFound == 0 {
         log.Println("Nenhum arquivo .go encontrado na raiz para compilar (ignorando o organizador).")
     }


	if buildErrors {
		return fmt.Errorf("%d arquivo(s) compilado(s) com sucesso, mas ocorreram erros em outros", builtCount)
	}
	if builtCount == 0 && goFilesFound > 0 {
         log.Println("Nenhum arquivo .go (além do organizador) foi compilado com sucesso.")
    } else if builtCount > 0 {
		log.Printf("%d arquivo(s) .go compilados com sucesso para '%s'.", builtCount, binDir)
	}


	return nil
}

// --- Lógica para -setupPath (Sem alterações) ---
func runSetupPath() error {
	// 1. Obter diretório home do usuário
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("não foi possível obter o diretório home do usuário: %w", err)
	}
	bashrcPath := filepath.Join(homeDir, bashrcFile)

	// 2. Obter caminho absoluto para o diretório bin
	rootDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("não foi possível obter o diretório de trabalho atual: %w", err)
	}
	absBinPath, err := filepath.Abs(filepath.Join(rootDir, binDir))
	if err != nil {
		return fmt.Errorf("não foi possível obter o caminho absoluto para '%s': %w", binDir, err)
	}

	// Garante que o diretório bin existe
	if _, err := os.Stat(absBinPath); os.IsNotExist(err) {
		log.Printf("Aviso: O diretório '%s' não existe. Criando para adicionar ao PATH.", absBinPath)
		if err := os.MkdirAll(absBinPath, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório '%s' para adicionar ao PATH: %w", absBinPath, err)
		}
	}

	// 3. Verificar se o .bashrc existe e se o PATH já está configurado
	pathExists := false
	exportLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", absBinPath)

	if _, err := os.Stat(bashrcPath); err == nil {
		file, err := os.Open(bashrcPath)
		if err != nil {
			return fmt.Errorf("erro ao abrir '%s' para verificação: %w", bashrcPath, err)
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.Contains(line, exportLine) || strings.HasPrefix(line, pathComment) {
				pathExists = true
				break
			}
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("erro ao ler '%s' para verificação: %w", bashrcPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("erro ao verificar status de '%s': %w", bashrcPath, err)
	}

	// 4. Adicionar ao .bashrc se não existir
	if pathExists {
		log.Printf("O caminho '%s' já parece estar configurado em '%s'. Nenhuma alteração feita.", absBinPath, bashrcPath)
		return nil
	}

	log.Printf("Adicionando '%s' ao PATH em '%s'...", absBinPath, bashrcPath)

	file, err := os.OpenFile(bashrcPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("erro ao abrir/criar '%s' para escrita: %w", bashrcPath, err)
	}
	defer file.Close()

	timestamp := time.Now().Format(time.RFC1123)
	lineToAdd := fmt.Sprintf("\n%s (%s)\n%s\n", pathComment, timestamp, exportLine)

	if _, err := file.WriteString(lineToAdd); err != nil {
		return fmt.Errorf("erro ao escrever em '%s': %w", bashrcPath, err)
	}

	log.Printf("Caminho adicionado com sucesso a '%s'.", bashrcPath)
	log.Println("IMPORTANTE: Para aplicar as mudanças, reinicie seu terminal ou execute:")
	log.Printf("  source %s", bashrcPath)

	return nil
}