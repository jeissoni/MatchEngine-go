package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/exp/rand"
)

type OrderType string

const (
	Buy  OrderType = "BUY"
	Sell OrderType = "SELL"
)

type Order struct {
	ID     int       `json:"ID"`
	Type   OrderType `json:"Type"`
	Price  float64   `json:"Price"`
	Amount int       `json:"Amount"`
	Index  int       `json:"Index"`
}

// generateOrders genera n órdenes aleatorias
func generateOrders(n int) []Order {

	// Crear un slice de órdenes
	orders := make([]Order, n)

	// Crear un slice de tipos de orden
	orderTypes := []OrderType{Buy, Sell} // Crear el slice una vez
	for i := 0; i < n; i++ {

		// Generar un precio aleatorio entre 90 y 110
		price := 90.0 + rand.Float64()*20.0

		// Generar una cantidad aleatoria entre 1 y 10
		amount := rand.Intn(10) + 1

		// Seleccionar un tipo de orden aleatorio
		// 0 -> Buy, 1 -> Sell
		orderType := orderTypes[rand.Intn(2)]

		// Crear la orden
		order := Order{
			ID:     i,
			Price:  price,
			Amount: amount,
			Type:   orderType,
		}

		// Agregar la orden al slice
		orders[i] = order
	}
	return orders
}

func generateOrdersWithMatchingPrices(n int) []Order {
	orders := make([]Order, n)
	orderTypes := []OrderType{Buy, Sell}
	basePrice := 100.0 // Precio base para la coincidencia
	priceRange := 5.0  // Rango de precios cercanos

	for i := 0; i < n; i++ {
		orderType := orderTypes[rand.Intn(2)]
		var price float64
		if orderType == Buy {
			price = basePrice + rand.Float64()*priceRange
		} else {
			price = basePrice - rand.Float64()*priceRange
		}

		amount := rand.Intn(10) + 1

		order := Order{
			ID:     i,
			Price:  price,
			Amount: amount,
			Type:   orderType,
		}
		orders[i] = order
	}
	return orders
}

func sendOrders(orders []Order) {
	url := "http://127.0.0.1:3000/orders"
	client := &http.Client{}

	// Crear un WaitGroup para esperar a que todas las gorutinas terminen
	// de enviar las órdenes
	var wg sync.WaitGroup

	// Iniciar el cronómetro
	startTime := time.Now()
	errorCount := 0

	// Recorrer todas las órdenes
	for _, order := range orders {

		// Incrementar el contador de gorutinas
		wg.Add(1)

		// Crear una gorutina para enviar la orden
		go func(order Order) {

			// Reducir el contador de gorutinas al finalizar
			defer wg.Done()

			// Convertir la orden a JSON
			jsonOrder, err := json.Marshal(order)
			if err != nil {
				fmt.Println("Error marshaling order:", err)
				errorCount++
				return
			}

			// Crear una nueva solicitud POST
			// con la orden en formato JSON
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonOrder))
			if err != nil {
				fmt.Println("Error creating request:", err)
				errorCount++
				return
			}

			// Establecer el tipo de contenido
			// en la cabecera de la solicitud
			req.Header.Set("Content-Type", "application/json")

			// Enviar la solicitud
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				errorCount++
				return
			}

			// Cerrar el cuerpo de la respuesta
			defer resp.Body.Close()

			// Verificar el código de estado
			if resp.StatusCode != http.StatusCreated {
				fmt.Printf("Error: status code %d\n", resp.StatusCode)
				errorCount++
			}

		}(order)
	}

	// Esperar a que todas las gorutinas terminen
	wg.Wait()

	// Detener el cronómetro y mostrar los resultados
	elapsedTime := time.Since(startTime)
	fmt.Printf("Sent %d orders in %s. Errors: %d\n", len(orders), elapsedTime, errorCount)
}

func specialLoadTest(duration time.Duration) {
	startTime := time.Now()
	for time.Since(startTime) < duration {
		numOrders := rand.Intn(100) + 1 // Ajustar según sea necesario
		orders := generateOrdersWithMatchingPrices(numOrders)
		sendOrders(orders)
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond) // Ajustar según sea necesario
	}
}

func normalLoadTest(duration time.Duration) {
	startTime := time.Now()
	for time.Since(startTime) < duration {
		numOrders := rand.Intn(20) + 1 // Ajustar según sea necesario
		orders := generateOrdersWithMatchingPrices(numOrders)
		sendOrders(orders)
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // Ajustar según sea necesario
	}
}

func main() {

	specialLoadTest(30 * time.Second)

	//normalLoadTest()

	fmt.Println("Test finished.")
}
