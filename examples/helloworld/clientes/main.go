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
	"os"
	"time"
	"encoding/csv"
	"io"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

//Lista que guardara los numeros de seguimiento recibidos por Logistica
var numeros_seguimiento []string


//Funcion que toma aleatoriamente un numero de seguimiento y consulta a Logistica por su estado
func seguimientoAleatorio(t int){
	//La funcion corre infinitamente y cada cierto tiempo solicita seguimiento
	for true{
		time.Sleep(time.Duration(t*5) * time.Second)
		if(len(numeros_seguimiento) >= 1){
			randomIndex := rand.Intn(len(numeros_seguimiento))
			n_seguimiento := numeros_seguimiento[randomIndex]
			conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			c := pb.NewGreeterClient(conn)

			ctx := context.Background()

			//Se consulta a Logistica por el estado del pedido con numero de seguimiento n_seguimiento
			res, err := c.SeguirPaquete(ctx, &pb.RespuestaPedido{Seguimiento:n_seguimiento})
			if err != nil {
						log.Fatalf("could not greet: %v", err)
			}
			log.Printf("El estado del paquete con numero de seguimiento %s es %s",n_seguimiento, res.GetSeguimiento())
		}
	}
}

func main() {
	var tipo_cliente int
	var tiempos_pedidos int
  fmt.Println("Seleccione Tipo de Cliente\n1-Retail\n2-Pymes")
  _, err := fmt.Scan(&tipo_cliente)
	fmt.Println("Indique el tiempo entre pedidos (Segundos)")
  _, err = fmt.Scan(&tiempos_pedidos)

	fmt.Println("El tiempo entre pedidos es ", tiempos_pedidos)

	//Se llama a la funcion para solicitar seguimientos periodicamente
	go seguimientoAleatorio(tiempos_pedidos)

	// Se setea la coneccion con el servidor
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx := context.Background()

	// Se ve si se trata de un cliente tipo Pyme o tipo Retail para leer el archivo csv respectivo
	if(tipo_cliente == 1){
		csvfile, err := os.Open("retail.csv")
		if err != nil {
			log.Fatalln("Couldn't open the csv file", err)
		}
		file := csv.NewReader(csvfile)
		// Se itera sobre los registros
		for {
			// Se lee cada registro del csv
			record, err := file.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if(record[0] != "id"){
				//Se envia el pedido Retail al servidor
				res, err := c.NuevoPedidoRetail(ctx, &pb.PedidoRetail{Id:record[0], Producto:record[1], Valor:record[2], Tienda:record[3],  Destino:record[4]})
				if err != nil {
							log.Fatalf("could not greet: %v", err)
				}
				//Se recibe y se gurda el numero de seguimiento
				log.Printf("El numero de seguimiento es: %s", res.GetSeguimiento())
				numeros_seguimiento = append(numeros_seguimiento, res.GetSeguimiento())
				//fmt.Printf("%s %s %s %s %s\n", record[0], record[1], record[2], record[3], record[4])
				time.Sleep(time.Duration(tiempos_pedidos) * time.Second)
			}
		}
	}else{
		csvfile, err := os.Open("pymes.csv")
		if err != nil {
			log.Fatalln("Couldn't open the csv file", err)
		}
		file := csv.NewReader(csvfile)
		for {
			record, err := file.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			if(record[0] != "id"){
				//Se envia el pedido Pyme al servidor
				res, err := c.NuevoPedidoPyme(ctx, &pb.PedidoPyme{Id:record[0], Producto:record[1], Valor:record[2], Tienda:record[3],  Destino:record[4], Prioritario: record[5]})
				if err != nil {
							log.Fatalf("could not greet: %v", err)
				}
				//Se recibe y se gurda el numero de seguimiento
				log.Printf("El numero de seguimiento para %s es: %s",record[0], res.GetSeguimiento())
				numeros_seguimiento = append(numeros_seguimiento, res.GetSeguimiento())
				time.Sleep(time.Duration(tiempos_pedidos) * time.Second)
			}
		}
	}
}
