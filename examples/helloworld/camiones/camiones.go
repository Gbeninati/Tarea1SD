/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */


package main

import (
	"context"
	"log"
	"net"
	"fmt"
	"time"
	"math/rand"
	"strconv"
	"encoding/csv"
	"os"


	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50052"
	address     = "10.6.40.228:50051"
	defaultName = "world"
)

var tiempo_espera int
var tiempo_viaje int

//Representan las colas del enunciado
var cola_normal [][]string
var cola_prioritaria [][]string
var cola_retail [][]string

//Inidcan si el camion_i esta disponible o se encuentra realizando entregas
var disponibilidad_camion1 bool = true
var disponibilidad_camion2 bool = true
var disponibilidad_camion3 bool = true

// Se utiliza para saber si el camnion_i espero el timepo necesario por un segundo paquete
var espera_camion1 bool = false
var espera_camion2 bool = false
var espera_camion3 bool = false

var camion2_listo = false

//Camion1 y Camion2 corresponden a camiones de retail, Camion3 corresponde a un camion normal
var camion1 [][]string
var camion2 [][]string
var camion3 [][]string

//Retorna true o false de manera aleatoria con probabilidades 0.8 y 0.2 respectivamente
func aleatorio() bool{
	rand.Seed(time.Now().UnixNano())
	randomChoice := rand.Intn(10)
	//fmt.Println(randomChoice)
	if(randomChoice <= 1){
		return false
	}else{
		return true
	}
}

//Recube una cola y realiza pop de esta
func pop(lis [][]string) [][]string{
	//fmt.Println("Se recibio la lista: ", lis)
	if(len(lis) == 1){
		lis = lis[:0]
	}else if(len(lis) >= 2){
		lis = lis[1:]
	}
	return lis
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

//Se encarga de guardar el registro de camion_i
func guardarRegistro(id string, tipo string, valor string, origen string, destino string, intentos string, estado string, nombre_archivo string){
	file, err := os.OpenFile(nombre_archivo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    checkError("Cannot create file", err)
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

		tiempo := time.Now()
		var registro []string
		if(estado == "Recibido"){
			registro = []string{id, tipo, valor, origen, destino, intentos, tiempo.Format("2006-01-02 15:04:05")}
		}else{
			registro = []string{id, tipo, valor, origen, destino, intentos, "0"}
		}

    err = writer.Write(registro)
    checkError("Cannot write to file", err)
}


type server struct {
	pb.UnimplementedGreeterServer
}


func esperarPaqueteCamion(cam int){
	time.Sleep(time.Duration(tiempo_espera) * time.Second)
	if(cam == 1){
		espera_camion1 = true
	}else if(cam == 2){
		espera_camion2 = true
	}else if(cam == 3){
		espera_camion3 = true
	}
}

// Envia un mensaje al centro de logistica con un numero de seguimiento, el estado de ese paquete y la cantidad de intentos. Se utilizar para que el centro de logistica pueda actualizar el estado de los paquetes
func UpdatateEstado(id_prod string, new_estado string, intentos string){
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx := context.Background()

	_, err = c.ActualizarEstado(ctx, &pb.NuevoEstado{Id:id_prod, Estado:new_estado, Intentos:intentos})
	if err != nil {
				log.Fatalf("could not greet: %v", err)
	}
}

//Funcion que recibe un camion con 1 o 2 paquetes listos para entregar e intenta hacer la entrega segun reestricciones del enunciado
func entregarPaquetes(camion [][]string) [][]string{
	//var camion_aux
	//Se verifica si el camion tiene 1 o 2 paquetes
	if(len(camion) == 1){
		//Mientras el estado del paquete sea distinto de Recibido o No Recibido se intenta entregar el paquete
		for {
			if (camion[0][5] == "Recibido" || camion[0][5] == "No Recibido"){
				break
			}
			//Se intenta entregar el paquete con 0.8 de exito y 0.2 de fracaso
			if(aleatorio()){
				//Simula el tiempo de viaje y luego se cambia su estado a Recibido
				time.Sleep(time.Duration(tiempo_viaje) * time.Second)
				camion[0][5] = "Recibido"
				fmt.Println("Se ha entregado el paquete con seguimiento", camion[0][1])
			}else{
				//Simula el tiempo de viaje
				time.Sleep(time.Duration(tiempo_viaje) * time.Second)
				//Se aumenta el numero de intentos
				intentos, _ := strconv.Atoi(camion[0][4])
				intentos = intentos + 1
				camion[0][4] = strconv.Itoa((intentos))
				fmt.Println("Se ha INTENTADO entregar el paquete con seguimiento", camion[0][1])
				//Si el numero de intentos es 3 y es de retail, entonces se cambia el estado a No Recibido
				if(camion[0][4] == "3" && camion[0][2] == "retail"){
					camion[0][5] = "No Recibido"
					fmt.Println("No se ha recciido el paquete con seguimiento", camion[0][1])
				}else if(camion[0][2] != "retail"){//Si el paquete es de Pymes entonces se verifica si continua siendo retable hacer la entrega o si el # intentos es 2
					valor_producto, _ := strconv.Atoi(camion[0][3])
					if(intentos == 2){
						camion[0][5] = "No Recibido"
						fmt.Println("No se ha recciido el paquete con seguimiento", camion[0][1])
					}else if(intentos*10 >= valor_producto){
						camion[0][5] = "No Recibido"
						fmt.Println("No se ha recciido el paquete con seguimiento", camion[0][1])
					}
				}
			}
		}
	}else if(len(camion) >= 2){ //Caso en que el camion tiene 2 paquetes
		//Se ordenan los paquetes dentro del camion segun su valor para entregar primero el paquete con mas valor
		valor1, _ := strconv.Atoi(camion[1][3])
		valor2, _ := strconv.Atoi(camion[0][3])
		if(valor1 > valor2){
			var aux []string
			aux = camion[0]
			camion[0] = camion[1]
			camion[1] = aux
		}
		for {
			//Mientras los estados de los paquetes sean distintos de Recibido o No Recibido se intenta entregar los paquetes
			if ((camion[0][5] == "Recibido" || camion[0][5] == "No Recibido") && (camion[1][5] == "Recibido" || camion[1][5] == "No Recibido")){
				break
			}
			// Se realiza lo mismo que el codigo para un solo paquete pero en este caso para ambos paquetes del camion
			for i, paquete := range camion{
				if(paquete[5] != "Recibido" || paquete[5] != "No Recibido"){
					//Se intenta entregar el paquete con 0.8 de exito y 0.2 de fracaso
					if(aleatorio()){
						//Simula el tiempo de viaje y luego de ese tiempo se cambia el estado a Recibido
						time.Sleep(time.Duration(tiempo_viaje) * time.Second)
						camion[i][5] = "Recibido"
						fmt.Println("Se ha entregado el paquete con seguimiento", camion[i][1])
						//fmt.Println("Dentro de EtregarPaquete() el camion es: ", camion)
						}else{
							//Simula el tiempo de viaje
							time.Sleep(time.Duration(tiempo_viaje) * time.Second)
							//Se aumenta el numero de intentos
							intentos, _ := strconv.Atoi(camion[i][4])
							intentos = intentos + 1
							camion[i][4] = strconv.Itoa(intentos)
							fmt.Println("Se ha INTENTADO entregar el paquete con seguimiento", camion[i][1])
							//Si el numero de intentos es 3 y es de retail, entonces se cambia el estado a No Recibido
							if(camion[i][4] == "3" && camion[i][2] == "retail"){//Si el paquete es de Pymes entonces se verifica si continua siendo retable hacer la entrega o si el # intentos es 2
								camion[i][5] = "No Recibido"
								fmt.Println("No se ha recubido el paquete con seguimiento", camion[i][1])
							}else if(camion[i][2] != "retail"){
								valor_producto, _ := strconv.Atoi(camion[i][3])
								if(intentos == 2){
									camion[i][5] = "No Recibido"
									fmt.Println("No se ha recciido el paquete con seguimiento", camion[i][1])
								}else if(intentos*10 >= valor_producto){
									camion[i][5] = "No Recibido"
									fmt.Println("No se ha recciido el paquete con seguimiento", camion[i][1])
								}
							}
						}
					}
				}
			}
}
return camion
}

//Funcion que simula el funcionamiento del Camion1 de Retail
func Camion1() {
	//Loop infinito que verifica si el camion esta disponible y si lo esta verifica si las colas tiene elementos o no
	for{
		if(disponibilidad_camion1){
			//Primero se verifica si la cola de retail tiene elementos, si es el caso entonces se agrga el paquete al camion y se elimina de la cola
			if(len(cola_retail) >= 1){
				camion1 = append(camion1, cola_retail[0])
				cola_retail= pop(cola_retail)
				camion2_listo = true
				//Si el camion luego de agregar el paquete tiene 2 paquetes en total, esta listo para realizar la entrega
				if(len(camion1) == 2){
					disponibilidad_camion1 = false
				}
			}else if(len(cola_prioritaria) >=1){ //Si la cola de retail no tiene elementos entonces se verifica que la cola prioritaria tenga y en caso de hacerlo, se agrega al camion
				camion1 = append(camion1, cola_prioritaria[0])
				cola_prioritaria = pop(cola_prioritaria)
				camion2_listo = true
				//Si el camion luego de agregar el paquete tiene 2 paquetes en total, esta listo para realizar la entrega
				if(len(camion1) == 2){
					disponibilidad_camion1 = false
				}
			}
		}
		//Si el camion tiene 1 paquete, no hay mas paquetes en las colas y ademas aun no espero el tiempo por un segundo paquete, espera este tiempo.
		if(len(camion1) == 1 && len(cola_retail) == 0 && len(cola_prioritaria) == 0 && espera_camion1 == false){
			time.Sleep(time.Duration(tiempo_espera) * time.Second)
			//Luego de esperar indica que si espero por un segundo paquete
			espera_camion1 = true
		}else if(len(camion1) == 2 || espera_camion1){//Si el camion tiene 2 paquetes o tiene uno y espero el tiempo por un segundo paquete que no llego, entonces se hace la entrega
			disponibilidad_camion1 = false
			//Se ve si el camion tiene 1 o 2 paquetes
			if(len(camion1) == 1){
				//Se actualiza el estado del paquete a En Camino y con 0 intetos por el momento
				UpdatateEstado(camion1[0][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion1)
				//Se realiza la entrega
				camion1 = entregarPaquetes(camion1)
				//Se actualiza el estado del paquete luego de la entrega
				UpdatateEstado(camion1[0][1], camion1[0][5], camion1[0][4])
				//fmt.Println("El camion luego de volver es ", camion1)
				//Se guarda el registro del camion
				guardarRegistro(camion1[0][0],camion1[0][2],camion1[0][3],camion1[0][6],camion1[0][7], camion1[0][4], camion1[0][5], "camion1.csv")
				//Se vacia el camion
				camion1 = pop(camion1)
			}else{
				//Se actualiza el estado de los paquetes a En Camino y con 0 intetos por el momento
				UpdatateEstado(camion1[0][1], "En Camino", "0")
				UpdatateEstado(camion1[1][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion1)
				//Se realiza la entrega
				camion1 = entregarPaquetes(camion1)
				//Se actualiza el estado de los paquetes luego de la entrega
				UpdatateEstado(camion1[0][1], camion1[0][5], camion1[0][4])
				UpdatateEstado(camion1[1][1], camion1[1][5], camion1[1][4])
				//fmt.Println("El camion luego de volver es ", camion1)
				//Se guarda el registro del camion
				guardarRegistro(camion1[0][0],camion1[0][2],camion1[0][3],camion1[0][6],camion1[0][7], camion1[0][4], camion1[0][5], "camion1.csv")
				guardarRegistro(camion1[1][0],camion1[1][2],camion1[1][3],camion1[1][6],camion1[1][7], camion1[1][4], camion1[1][5], "camion1.csv")
				//Se vacia el camion
				camion1 = pop(camion1)
				camion1 = pop(camion1)
		}
		//El camion pasa a estar nuevamente disponible
			disponibilidad_camion1 = true
			espera_camion1 = false
		}
	}
}

//Para las funciones Camion2 y Camion3, el codigo es igual al de Camion1 pero trabajando sobre la lista de camion respectiva.
//En el caso de Camion3 en vez de trabajar con las colas de retail y prioritaria, se trabaja con prioritaria y normal.

func Camion2() {
	//var yaHay1Prioritario bool = false
	//time.Sleep(1 * time.Second)
	for{
		if(disponibilidad_camion2 && camion2_listo){
			if(len(cola_retail) >= 1){
				camion2 = append(camion2, cola_retail[0])
				cola_retail= pop(cola_retail)
				if(len(camion2) == 2){
					disponibilidad_camion2 = false
				}
			}else if(len(cola_prioritaria) >=1){
				camion2 = append(camion2, cola_prioritaria[0])
				cola_prioritaria = pop(cola_prioritaria)
				//yaHay1Prioritario = true
				if(len(camion2) == 2){
					disponibilidad_camion2 = false
				}
			}
		}

		if(len(camion2) == 1 && len(cola_retail) == 0 && len(cola_prioritaria) == 0 && espera_camion2 == false){
			time.Sleep(time.Duration(tiempo_espera) * time.Second)
			espera_camion2 = true
		}else if(len(camion2) == 2 || espera_camion2){
			disponibilidad_camion2 = false
			if(len(camion2) == 1){
				UpdatateEstado(camion2[0][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion2)
				camion2 = entregarPaquetes(camion2)
				UpdatateEstado(camion2[0][1], camion2[0][5], camion2[0][4])
				//fmt.Println("El camion luego de volver es ", camion2)
				guardarRegistro(camion2[0][0],camion2[0][2],camion2[0][3],camion2[0][6],camion2[0][7], camion2[0][4], camion2[0][5], "camion2.csv")
				camion2 = pop(camion2)
			}else{
				UpdatateEstado(camion2[0][1], "En Camino", "0")
				UpdatateEstado(camion2[1][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion2)
				camion2 = entregarPaquetes(camion2)
				UpdatateEstado(camion2[0][1], camion2[0][5], camion2[0][4])
				UpdatateEstado(camion2[1][1], camion2[1][5], camion2[1][4])
				guardarRegistro(camion2[0][0],camion2[0][2],camion2[0][3],camion2[0][6],camion2[0][7], camion2[0][4], camion2[0][5], "camion2.csv")
				guardarRegistro(camion2[1][0],camion2[1][2],camion2[1][3],camion2[1][6],camion2[1][7], camion2[1][4], camion2[1][5], "camion2.csv")
				//fmt.Println("El camion luego de volver es ", camion2)
				camion2 = pop(camion2)
				camion2 = pop(camion2)
		}
			disponibilidad_camion2 = true
			espera_camion2 = false
			//yaHay1Prioritario = false
		}
	}
}

func Camion3() {
	time.Sleep(300 * time.Millisecond)
	//var esperando bool = true
	for{
		if(disponibilidad_camion3){
			if(len(cola_prioritaria) >= 1){
				camion3 = append(camion3, cola_prioritaria[0])
				cola_prioritaria = pop(cola_prioritaria)
				if(len(camion3) == 2){
					disponibilidad_camion3 = false
				}
			}else if(len(cola_normal) >=1){
				camion3 = append(camion3, cola_normal[0])
				cola_normal = pop(cola_normal)
				if(len(camion3) == 2){
					disponibilidad_camion3 = false
				}
			}
		}

		if(len(camion3) == 1 && len(cola_prioritaria) == 0 && len(cola_normal) == 0 && espera_camion3 == false){
			time.Sleep(time.Duration(tiempo_espera) * time.Second)
			espera_camion3 = true
		}else if(len(camion3) == 2 || espera_camion3){
			disponibilidad_camion3 = false
			if(len(camion3) == 1){
				UpdatateEstado(camion3[0][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion3)
				camion3 = entregarPaquetes(camion3)
				UpdatateEstado(camion3[0][1], camion3[0][5], camion3[0][4])
				//fmt.Println("El camion luego de volver es ", camion3)
				guardarRegistro(camion3[0][0],camion3[0][2],camion3[0][3],camion3[0][6],camion3[0][7], camion3[0][4], camion3[0][5], "camion3.csv")
				camion3 = pop(camion3)
			}else{
				UpdatateEstado(camion3[0][1], "En Camino", "0")
				UpdatateEstado(camion3[1][1], "En Camino", "0")
				//fmt.Println("El camion antes del viaje es ", camion3)
				camion3 = entregarPaquetes(camion3)
				UpdatateEstado(camion3[0][1], camion3[0][5], camion3[0][4])
				UpdatateEstado(camion3[1][1], camion3[1][5], camion3[1][4])
				//fmt.Println("El camion luego de volver es ", camion3)
				guardarRegistro(camion3[0][0],camion3[0][2],camion3[0][3],camion3[0][6],camion3[0][7], camion3[0][4], camion3[0][5], "camion3.csv")
				guardarRegistro(camion3[1][0],camion3[1][2],camion3[1][3],camion3[1][6],camion3[1][7], camion3[1][4], camion3[1][5], "camion3.csv")
				camion3 = pop(camion3)
				camion3 = pop(camion3)
		}
			disponibilidad_camion3 = true
			espera_camion3 = false
		}
	}
}

// Se reciben los paquetes del centro de logistica y se guardan en la colas respectivas
func (s *server) Camiones(ctx context.Context, in *pb.PaqueteCamion) (*pb.RespuestaPedido, error) {
  log.Printf("Se recibio el paquete %s %s %s %s %s %s", in.GetId(), in.GetNro_Seguimiento(), in.GetTipo(), in.GetValor(), in.GetIntentos(), in.GetEstado())
	aux := []string {in.GetId(), in.GetNro_Seguimiento(), in.GetTipo(), in.GetValor(), in.GetIntentos(), in.GetEstado(), in.GetTienda(), in.GetDestino()}
	if(in.GetTipo() == "normal"){
		cola_normal= append(cola_normal,aux)
	}
	if(in.GetTipo() == "prioritario"){
		cola_prioritaria = append(cola_prioritaria,aux)
	}
	if(in.GetTipo() == "retail"){
		cola_retail = append(cola_retail,aux)
	}
	//fmt.Println("El largo de la cola normal, prioritario y retail es respectivamente", len(cola_normal),len(cola_prioritaria),len(cola_retail))
  return &pb.RespuestaPedido{Seguimiento: "En Bodega"}, nil
}

func main() {
	fmt.Println("Indique el tiempo de espera del camion (Segundos)")
  _, _ = fmt.Scan(&tiempo_espera)
	fmt.Println("Indique el tiempo que demora el envio (Segundos)")
  _, _ = fmt.Scan(&tiempo_viaje)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Se llaman a los camiones con subrutinas para que funcionen de manera paralela
	go Camion1()
	go Camion2()
	go Camion3()
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
