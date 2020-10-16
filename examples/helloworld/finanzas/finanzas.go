package main

import (
	"fmt"
	"strings"
	"strconv"
	"encoding/csv"
	"os"
	"log"
	"github.com/streadway/amqp"
)

//Variables que iran guardando las ganancias, perdidas y balance final totales
var ganancias_totales float64 = 0.0
var perdidas_totales float64 = 0.0
var balance_total float64 = 0.0

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}

//Se guarda toda la informacion del producto en un archivo csv con los campos id, nombre, valor, tipo, intentos, estado, balance
func guardarRegistro(id string, nombre string, valor string, tipo string, intentos string, estado string, balance string){
	file, err := os.OpenFile("registrofinanzas.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  checkError("Cannot create file", err)
  defer file.Close()

  writer := csv.NewWriter(file)
  defer writer.Flush()

  registro := []string{id, nombre, valor, tipo, intentos, estado, balance}
  err = writer.Write(registro)
  checkError("Cannot write to file", err)
}

//Funcion que se encarga de obtener las ganancias, perdidas y balance final de un producto
func obtenerBalance(datos string){
	aux := strings.Split(datos, ",")
	var balance float64 = 0.0
	var ganancia float64 = 0.0
	var perdida float64 = 0.0
	valor,_ := strconv.ParseFloat(aux[4], 64)
	intentos,_ := strconv.ParseFloat(aux[7], 64)
	perdida += intentos * 10
	if(aux[0] == "prioritario"){
		ganancia += valor*0.3
	}
	if(aux[1] == "Recibido"){
		ganancia += valor
	}else if(aux[0] == "retail"){
		ganancia += valor
	}
	//fmt.Printf("Las ganancias del paquete", aux[2], "son ", ganancia)
	//fmt.Printf("Las perdidas del paquete", aux[2], "son ", perdida)
	balance = ganancia - perdida
	ganancias_totales += ganancia
	perdidas_totales += perdida
	balance_total += balance
	//Se gudardan los datos en el registro (id_producto, nombre, valor, tipo, intentos, estado, balance)
	blnc := fmt.Sprintf("%f", balance)
	guardarRegistro(aux[2], aux[3], aux[4], aux[0], aux[7], aux[1], blnc)
}

func main() {
	//Se realiza la conexion con RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Println("Failed Initializing Broker Connection")
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"TestQueue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			fmt.Printf("Se recibio el mensaje: %s\n", d.Body)
			s := string(d.Body)
			obtenerBalance(s)
			fmt.Printf("Las ganacias actuales totales son: %f\n", ganancias_totales)
			fmt.Printf("Las perdidas actuales totales son: %f\n", perdidas_totales)
			fmt.Printf("El balance total actual es: %f\n", balance_total)
			fmt.Println()
		}
	}()

	fmt.Println("Successfully Connected to our RabbitMQ Instance")
	fmt.Println(" [*] - Waiting for messages")
	<-forever
}
