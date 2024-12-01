package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	FromID int
	Plan   string
	Round  int
}

func main() {
	// Leer las variables de entorno
	nodeID, _ := strconv.Atoi(os.Getenv("NODE_ID"))
	totalNodes, _ := strconv.Atoi(os.Getenv("NODES"))
	numTraitors, _ := strconv.Atoi(os.Getenv("NUM_TRAITORS"))
	nodePlan := os.Getenv("PLAN") // R: Retirada, A: Ataque
	traitor := os.Getenv("TRAITOR")
	isTraitor := false

	if traitor == "YES" {
		isTraitor = true
	} else if traitor == "NO" {
		isTraitor = false
	}

	basePort := 8080
	port := strconv.Itoa(basePort)

	isKing := false
	currentRound := 1
	rounds := numTraitors + 1

	// mutex := &sync.Mutex{}

	// Cada nodo tendrá una lista de puertos de los demás nodos
	var ports []string
	for i := 1; i <= totalNodes; i++ {
		address := fmt.Sprintf("node%d:%s", i, port) // Dirección completa -> nodei:8080
		if i != nodeID {                             // Excluir el nodo actual
			ports = append(ports, address)
		}
	}

	plans := make(map[string]string) // Direcciones de otros nodos y sus planes (inicialmente vacío) -> plans[nodei]= ""

	for i := 1; i <= totalNodes; i++ {
		address := fmt.Sprintf("node%d", i)
		if i != nodeID { // Excluir el nodo actual
			plans[address] = ""
		} else {
			plans[address] = string(nodePlan)
		}
	}

	fmt.Println("Mapa inicial de planes:")
	for address, plan := range plans {
		fmt.Printf("%s -> %s\n", address, plan)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go startPeerListener(port, wg, nodeID, plans)

	time.Sleep(1 * time.Second) // Esperar a que el listener inicie

	connections := startConnections(ports) // Conectar con los demás nodos
	defer endConnections(connections)

	time.Sleep(1 * time.Second) // Esperar a que las conexiones se establezcan entre nodos

	// Iniciamos algoritmo del rey
	for currentRound <= rounds {

		if getKing(currentRound, totalNodes) == nodeID {
			isKing = true
		}

		if isKing {
			fmt.Printf("Ronda %d\n", currentRound)
			fmt.Println("Fase 1: Generales intercambian información")
		}

		time.Sleep(1 * time.Second) // Tiempo antes de fase 1

		// Primera fase -> Todos se envían sus planes
		sendPlan(connections, nodeID, nodePlan, currentRound, isTraitor)
		// RecievePlan in handleIncomingConnection
		time.Sleep(1 * time.Second) // Esperar a que los planes sean envíados y recibidos

		fmt.Printf("Mapa de planes actualizado del nodo%d:\n", nodeID)
		for address, plan := range plans {
			fmt.Printf("%s -> %s\n", address, plan)
		}
		time.Sleep(1 * time.Second) // Tiempo antes de fase 2

		// Segunda fase -> Solo el rey envía su plan a los demás generales

		if isKing {
			fmt.Println("Fase 2")
			kingPlan := plans[fmt.Sprintf("node%d", nodeID)]
			fmt.Printf("Rey (Nodo %d) elige plan '%s' y lo envía a todos los generales\n", nodeID, kingPlan)
			sendPlan(connections, nodeID, kingPlan, currentRound, isTraitor)
			time.Sleep(1 * time.Second) //
		} else {
			// Recibir plan del rey, comparar con voto mayoría de R vs A, si es
			// abrumador, (> n/2 + t), cambiar el plan del general al del rey
			// si no, mantener el plan del general
			time.Sleep(1 * time.Second) // Tiempo para esperar al rey
			kingPlan := plans[fmt.Sprintf("node%d", getKing(currentRound, totalNodes))]

			if validateKingPlan(plans, totalNodes, numTraitors) {
				plans[fmt.Sprintf("node%d", nodeID)] = nodePlan
			} else {
				nodePlan = kingPlan
				plans[fmt.Sprintf("node%d", nodeID)] = kingPlan
			}
		}
		currentRound++
		time.Sleep(1 * time.Second) // Tiempo para esperar a la segunda ronda
		isKing = false
	}

	time.Sleep(2 * time.Second) // Esperamos a que todo nodo tome una decisión

	fmt.Printf("Decisión final %s\n", plans[fmt.Sprintf("node%d", nodeID)])

	time.Sleep(1 * time.Second) // Esperamos a que los nodos printeen antes de cerrar programa

	wg.Done()
	wg.Wait()
}

func handleIncomingConnection(conn net.Conn, plans map[string]string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		messageBytes, err := reader.ReadString('\n') // Lee un mensaje completo (hasta '\n')
		if err != nil {
			return
		}

		// Decodifica el mensaje
		var message Message
		err = json.Unmarshal([]byte(messageBytes), &message)
		if err != nil {
			fmt.Printf("Error al decodificar mensaje '%s': %s\n", messageBytes, err)
			continue
		}

		// fmt.Printf("Recibió mensaje de nodo%d: %+v\n", message.FromID, message)

		// Actualiza el estado local con el plan recibido
		// mutex.Lock() // Evita condiciones de carrera al acceder a `plans`
		plans[fmt.Sprintf("node%d", message.FromID)] = message.Plan
		// mutex.Unlock()
	}
}

func sendPlan(connections map[string]net.Conn, fromID int, plan string, round int, isTraitor bool) {

	for port, conn := range connections {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

		currentPlan := plan

		if isTraitor && randGen.Float64() < 0.5 { // 50% chance de que haya falla bizantina
			if currentPlan == "R" {
				currentPlan = "A"
				fmt.Println("Falla bizantina: Cambiando un envío de plan de R a A")
			} else {
				currentPlan = "R"
				fmt.Println("Falla bizantina: Cambiando un envío de plan de A a R")
			}
		}

		// Crear el mensaje
		message := Message{FromID: fromID, Plan: currentPlan, Round: round}
		data, err := json.Marshal(message)

		if err != nil {
			fmt.Printf("Error al serializar mensaje: %s\n", err)
			continue
		}

		// Enviar el mensaje
		_, err = conn.Write(append(data, '\n'))
		if err != nil {
			fmt.Printf("Error al enviar mensaje a %s: %s\n", port, err)
		}
	}
}

func getKing(round int, totalNodes int) int {
	return round % totalNodes // Nodo que será el Rey para esta ronda
}

func validateKingPlan(plans map[string]string, totalNodes, numTraitors int) bool {
	threshold := totalNodes/2 + numTraitors

	planVotes := make(map[string]int)

	for _, plan := range plans {
		planVotes[plan]++
	}

	var majorityPlan string
	maxVotes := 0
	for plan, count := range planVotes {
		if count > maxVotes {
			maxVotes = count
			majorityPlan = plan
		}
	}

	fmt.Printf("La votación mayoritaría fue de %d votos sobre %d para '%s'\n", maxVotes, totalNodes, majorityPlan)

	if maxVotes > threshold { // abrumador
		fmt.Println("El plan de la mayoría es abrumadora. Se ignora el plan del rey.")
		return true
	} else { // maxVotes <= threshold, No abrumador
		fmt.Println("El plan de la mayoría NO fue abrumadora, se seguirá el plan del rey.")
		return false
	}
}

func startPeerListener(port string, wg *sync.WaitGroup, nodeID int, plans map[string]string) {
	defer wg.Done()
	own_address := fmt.Sprintf("node%d:%s", nodeID, port)

	fmt.Printf("Listener activado: Escuchando en %s\n", own_address)

	listener, err := net.Listen("tcp", own_address)
	if err != nil {
		fmt.Printf("Error al iniciar listener: %s\n", err)
		return
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		go handleIncomingConnection(conn, plans)
	}
}

func startConnections(ports []string) map[string]net.Conn {
	connections := make(map[string]net.Conn)

	for _, remotePort := range ports {
		conn, err := net.Dial("tcp", remotePort)
		if err != nil {
			fmt.Printf("Error al conectar con %s: %s\n", remotePort, err)
			continue
		}

		connections[remotePort] = conn
		// fmt.Printf("Conexión establecida con %s\n", remotePort)
	}
	return connections
}

func endConnections(connections map[string]net.Conn) {
	for _, conn := range connections {
		// fmt.Printf("Cerrando conexión con %s\n", port)
		conn.Close()
	}
}
