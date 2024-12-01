Instrucciones:

Para ejecutar la tarea primero hay que correr el siguiente comando desde la carpeta tarea-3-ttrincado:

    go run setup.go

Una vez ejecutado, debería aparecer en la terminal una solicitud para introducir el número de nodos:

    "Ingrese la cantidad de nodos (n):"

Para esto ingrese una cantidad >= 5 (se asume que la cantidad mínima de traidores es 1), para poder trabajar con esa cantidad de nodos. 

Posteriormente, se solicitará el número de traidores:

    "Ingrese la cantidad de traidores (t):" 

Luego, se preguntará por el plan que desea que cada nodo tome:

    "¿Cuál es el plan para el nodo i? (R/A):"

De esta forma se asignarán los planes para cada nodo en el docker compose.

Finalmente, dependiendo del número de traidores, se preguntará i veces qué nodo se quiere que actúe como traidor:

    "Seleccione el nodo para el traidor i:"

Por simplicidad (También, para cumplir con restricción puesta en las diapositivas de clases), no podrán haber 2 nodos traidores seguidos.

Una vez se hayan ingresado los inputs que cumplan con las condiciones especificadas anteriormente, se generará el archivo docker-compose.yml con las características de cada nodo, los contenedores se levantarán y correrán el archivo main.go que contiene el flujo principal del algoritmo del rey (comando en Dockerfile), donde se irá mostrando en consola el estado y progreso de cada nodo. 

A lo largo del main.go, habrá una cierta cantidad de time.sleeps localizados estratégicamente para poder hacer que los nodos no partan procesos sin que otros nodos hayan terminados, y que los outputs sean más ordenados de ver. 

Importante a tener en cuenta, para reflejar el comportamiento de un traidores, se simulará la inconsistencia de la siguiente manera:

    50% chance de cambiar plan
        Plan -> If A then B
        Plan -> If B then A

Además, las rondas serán lideradas por los primeros t+1 nodos, por lo que para asignar un traidor como rey, este deberá tener un id < t+1.
    
Consideraciones: 
    - Generalmente probando muchas veces seguidas con imagenes existentes empiecen a fallar conexiones, para solucionar esto solamente hay que limpiar las imagenes o reiniciar docker.
    - El programa en mi computador empieza a caerse con una mayor cantidad de nodos! >15