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

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"encoding/csv"
	"os"
	"time"
	"strings"
	"fmt"
	"github.com/streadway/amqp"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
	address     = "10.6.40.225:50052"
	defaultName = "world"
)

// Varible que se utilizara para ir generando numeros de seguimiento
var seg int = 0

//Similar a diccionario de pyton en donde el la llave sera el numero de seguimiento y el valor sera una lista con todos los datos del producto con ese seguimiento
var pedidos = make(map[string][]string)


type server struct {
	pb.UnimplementedGreeterServer
}


//Funcion que se encarga de establecer conexion con los camiones y enviar los paquetes. Como parametro se reciben todos los datos necesarios para formar un paquete.
func conCaminon(id string, seguimiento string, tipo string, valor string, intentos string, estado string, tienda string, destino string){
	//Se realiza la conexion con el centro de camiones
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx := context.Background()
	//Se envia el paquete
	res, err := c.Camiones(ctx, &pb.PaqueteCamion{Id:id, Nro_Seguimiento:seguimiento, Tipo:tipo, Valor:valor, Intentos: intentos, Estado:estado, Tienda:tienda, Destino:destino})
	if err != nil {
				log.Fatalf("could not greet: %v", err)
	}
	//log.Printf("El estado actual del paquete es: %s", res.GetSeguimiento())

}

//Funcion que guarda en el archivo registropaquetes.csv un registro de todos los paquetes en el centro de logistica
func guardarRegistro(id string, tipo string, nombre string, valor string, origen string, destino string, seguimiento string){
	file, err := os.OpenFile("registropaquetes.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    checkError("Cannot create file", err)
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

		tiempo := time.Now()
    registro := []string{tiempo.Format("2006-01-02 15:04:05"),id, tipo, nombre, valor, origen, destino, seguimiento}
    err = writer.Write(registro)
    checkError("Cannot write to file", err)
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

// Se encarga de recibir los pedidos de las Pymes, guardarlos en el diccionario pedidos y de enviar estos paquetes al centro de camiones.
func (s *server) NuevoPedidoPyme(ctx context.Context, in *pb.PedidoPyme) (*pb.RespuestaPedido, error) {
	//Se actualiza el numero de seguimiento
	seg = seg+1
	log.Printf("Se recibio el producto %s %s %s %s %s %s\n", in.GetId(), in.GetProducto(), in.GetValor(), in.GetTienda(), in.GetDestino(), in.GetPrioritario())
	//Se ve si el paquete es "normal" o "prioritario", se agrega al diccionario y se envia al centro de camiones
	if(in.GetPrioritario() == "0"){
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "normal")
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "En Bodega")
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetId())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetProducto())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetValor())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetTienda())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetDestino())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "0")
		guardarRegistro(in.GetId(), "normal", in.GetProducto(), in.GetValor(), in.GetTienda(), in.GetDestino(),"ABCDE" + strconv.Itoa(seg))
		conCaminon(in.GetId(),"ABCDE" + strconv.Itoa(seg), "normal", in.GetValor(), "0", "En Bodega", in.GetTienda(), in.GetDestino())
	}else{
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "prioritario")
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "En Bodega")
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetId())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetProducto())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetValor())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetTienda())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetDestino())
		pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "0")
		guardarRegistro(in.GetId(), "prioritario", in.GetProducto(), in.GetValor(), in.GetTienda(), in.GetDestino(),"ABCDE" + strconv.Itoa(seg))
		conCaminon(in.GetId(),"ABCDE" + strconv.Itoa(seg), "prioritario", in.GetValor(), "0", "En Bodega", in.GetTienda(), in.GetDestino())
	}
  return &pb.RespuestaPedido{Seguimiento: "ABCDE" + strconv.Itoa(seg)}, nil
}

// Se encarga de recibir los pedidos de Retail, guardarlos en el diccionario pedidos y de enviar estos paquetes al centro de camiones.
func (s *server) NuevoPedidoRetail(ctx context.Context, in *pb.PedidoRetail) (*pb.RespuestaPedido, error) {
	seg = seg+1
	log.Printf("Se recibio el producto %s %s %s %s %s\n", in.GetId(), in.GetProducto(), in.GetValor(), in.GetTienda(), in.GetDestino())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "retail")
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "En Bodega")
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetId())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetProducto())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetValor())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetTienda())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], in.GetDestino())
	pedidos["ABCDE" + strconv.Itoa(seg)] = append(pedidos["ABCDE" + strconv.Itoa(seg)], "0")

	conCaminon(in.GetId(),"ABCDE" + strconv.Itoa(seg), "retail", in.GetValor(), "0", "En Bodega", in.GetTienda(), in.GetDestino())

	guardarRegistro(in.GetId(), "retail", in.GetProducto(), in.GetValor(), in.GetTienda(), in.GetDestino(),"ABCDE" + strconv.Itoa(seg))
  return &pb.RespuestaPedido{Seguimiento: "ABCDE" + strconv.Itoa(seg)}, nil
}

// Recibe una solicitud de seguimiento del cliente con el numero de seguimiento y envia como respuesta el estado del paquete con ese numero de seguimiento
func (s *server) SeguirPaquete(ctx context.Context, in *pb.RespuestaPedido) (*pb.RespuestaPedido, error) {
	log.Printf("Pedido %s solicita seguimiento\n", in.GetSeguimiento())
	return &pb.RespuestaPedido{Seguimiento: pedidos[in.GetSeguimiento()][1]}, nil
}

//Funcion que recibe un string con todos los datos de un paquete y lo envia a finanzas para que realice el calculo del balance final
func comunicarFinanzas(data string){
	conn, err := amqp.Dial("amqp://guest:guest@10.6.40.227:5672/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		panic(err)
	}

    // Se inicia la conexion con RabbitMQ
	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

    //Se declaran las colas a utilizar, en este caso TestQueue
	_, err = ch.QueueDeclare(
		"TestQueue",
		false,
		false,
		false,
		false,
		nil,
	)
    // Se manejan errores en caso de haber
	if err != nil {
		fmt.Println(err)
	}

    // Se publica el mensaje a la cola
	err = ch.Publish(
		"",
		"TestQueue",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(data),
		},
	)

	if err != nil {
		fmt.Println(err)
	}
    //log.Printf("Pedido %s enviado a finanzas\n", data[2])
}

//Actualiza el estado y el # de intentos del pedido con el numero de seguimiento recibido por caminones
func (s *server) ActualizarEstado(ctx context.Context, in *pb.NuevoEstado) (*pb.RespuestaPedido, error) {
	log.Printf("El pedido %s actualizo su estado a %s\n", in.GetId(), in.GetEstado())
	if(len(pedidos[in.GetId()]) > 1){
		pedidos[in.GetId()][1] = in.GetEstado()
		pedidos[in.GetId()][7] = in.GetIntentos()
	}
	if(in.GetEstado() == "Recibido" || in.GetEstado() == "No Recibido"){
		var datos string
		datos = strings.Join(pedidos[in.GetId()], ",")
		//fmt.Printf(datos)
		go comunicarFinanzas(datos)
	}
	return &pb.RespuestaPedido{Seguimiento: ""}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
