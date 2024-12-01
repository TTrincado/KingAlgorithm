package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// Pedir al usuario la cantidad de nodos (n)
	fmt.Print("Ingrese la cantidad de nodos (n): ")
	nInput, _ := reader.ReadString('\n')
	n, err := strconv.Atoi(strings.TrimSpace(nInput))
	if err != nil || n <= 0 {
		fmt.Println("Error: Debe ingresar un número entero positivo para la cantidad de nodos.")
		return
	}

	// Pedir al usuario la cantidad de traidores (t)
	fmt.Print("Ingrese la cantidad de traidores (t): ")
	tInput, _ := reader.ReadString('\n')
	t, err := strconv.Atoi(strings.TrimSpace(tInput))
	if err != nil || t < 1 {
		fmt.Println("Error: Debe ingresar un número entero no negativo para la cantidad de traidores.")
		return
	}

	// Validar la condición n >= 4t + 1
	if n < 4*t+1 {
		fmt.Printf("Error: La cantidad de nodos (n=%d) debe ser al menos 4t+1 (t=%d).\n", n, t)
		return
	}

	// Pedir el plan para cada nodo
	plans := make([]string, n)
	for i := 1; i <= n; i++ {
		for {
			fmt.Printf("¿Cuál es el plan para el nodo %d? (R/A): ", i)
			planInput, _ := reader.ReadString('\n')
			plan := strings.TrimSpace(strings.ToUpper(planInput))
			if plan == "R" || plan == "A" {
				plans[i-1] = plan
				break
			} else {
				fmt.Println("Error: El plan debe ser 'R' (Retirada) o 'A' (Ataque). Intente nuevamente.")
			}
		}
	}

	// Seleccionar nodos traidores
	traitors := make([]int, 0, t)
	for i := 1; i <= t; i++ {
		for {
			fmt.Printf("Seleccione el nodo para el traidor %d: ", i)
			traitorInput, _ := reader.ReadString('\n')
			traitorID, err := strconv.Atoi(strings.TrimSpace(traitorInput))
			if err != nil || traitorID < 1 || traitorID > n {
				fmt.Println("Error: Debe ingresar un ID válido entre 1 y", n)
				continue
			}

			// Validar que el traidor no sea consecutivo a otro ya seleccionado
			isValid := true
			for _, existing := range traitors {
				if traitorID == existing || traitorID == existing-1 || traitorID == existing+1 {
					isValid = false
					break
				}
			}

			if isValid {
				traitors = append(traitors, traitorID)
				break
			} else {
				fmt.Println("Error: No puede seleccionar un nodo consecutivo a otro traidor ya elegido.")
			}
		}
	}

	printNodeStructure(n, plans, traitors)

	// Generar archivo docker-compose.yml
	fmt.Println("\nGenerando archivo docker-compose.yml...")
	err = generateDockerCompose(n, t, plans, traitors)
	if err != nil {
		fmt.Printf("Error generando el archivo docker-compose.yml: %s\n", err)
		return
	}

	// Ejecutar docker-compose up
	fmt.Println("\nLevantando contenedores con docker-compose...")
	err = runDockerCompose()
	if err != nil {
		fmt.Printf("Error ejecutando docker-compose: %s\n", err)
		return
	}
}

func generateDockerCompose(n int, t int, plans []string, traitors []int) error {
	// Crear el contenido del archivo docker-compose.yml
	services := ""
	for i := 1; i <= n; i++ {
		isTraitor := "NO"
		for _, traitorID := range traitors {
			if i == traitorID {
				isTraitor = "YES"
				break
			}
		}

		services += fmt.Sprintf(`
  node%d:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: node%d
    environment:
      - NODE_ID=%d
      - NODES=%d
      - NUM_TRAITORS=%d
      - PLAN=%s
      - TRAITOR=%s
    ports:
      - "%d:8080"
    stdin_open: true
    tty: true
    networks:
      - byzantine
`, i, i, i, n, t, plans[i-1], isTraitor, 8080+i)
	}

	dockerCompose := fmt.Sprintf(`version: "3.8"
services:%s
networks:
  byzantine:
    driver: bridge
`, services)

	// Escribir el archivo docker-compose.yml
	file, err := os.Create("docker-compose.yml")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(dockerCompose)
	return err
}

func runDockerCompose() error {
	cmd := exec.Command("docker-compose", "up", "--build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printNodeStructure(n int, plans []string, traitors []int) {
	fmt.Println("\nEstructura actual de nodos:")
	for i := 1; i <= n; i++ {
		traitorLabel := ""
		for _, traitorID := range traitors {
			if i == traitorID {
				traitorLabel = " (TRAITOR)"
				break
			}
		}
		fmt.Printf("Nodo %d -> %s%s\n", i, plans[i-1], traitorLabel)
	}
}
