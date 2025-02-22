package main

import (
	"container/heap"

	"fmt"

	"sync"
	"time"

	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type OrderType string

const (
	Buy  OrderType = "BUY"
	Sell OrderType = "SELL"
)

// Order representa una orden en el libro
type Order struct {
	ID     int
	Type   OrderType
	Price  float64
	Amount int
	Index  int // Necesario para la estructura heap
}

// ==========   Heap de compras  ========================
type BuyHeap []*Order

// Implementar la interfaz heap.Interface
// para que BuyHeap sea un heap de máximos
// (la orden con el precio más alto estará en la parte superior)
// https://golang.org/pkg/container/heap/

// Len, Less, Swap, Push y Pop son métodos necesarios para implementar la interfaz heap.Interface

// Len devuelve la longitud del heap
func (h BuyHeap) Len() int { return len(h) }

// Less determina el orden de clasificación
// en este caso, queremos un heap de máximos
// por lo que la orden con el precio más alto estará en la parte superior
func (h BuyHeap) Less(i, j int) bool { return h[i].Price > h[j].Price }

// Swap intercambia dos elementos en el heap
// y actualiza los índices de los elementos
// (necesario para la estructura heap)
func (h BuyHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

// Push agrega un elemento al heap
// y actualiza el índice del elemento
func (h *BuyHeap) Push(x interface{}) {
	order := x.(*Order)
	order.Index = len(*h)
	*h = append(*h, order)
}

// Pop elimina un elemento del heap
// y actualiza el índice del elemento
func (h *BuyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	order := old[n-1]
	order.Index = -1
	*h = old[0 : n-1]
	return order
}

// heap de ventas
type SellHeap []*Order

func (h SellHeap) Len() int { return len(h) }

// En este caso, queremos un heap de mínimos
// por lo que la orden con el precio más bajo estará en la parte superior
func (h SellHeap) Less(i, j int) bool { return h[i].Price < h[j].Price }
func (h SellHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *SellHeap) Push(x interface{}) {
	order := x.(*Order)
	order.Index = len(*h)
	*h = append(*h, order)
}

func (h *SellHeap) Pop() interface{} {
	old := *h
	n := len(old)
	order := old[n-1]
	order.Index = -1
	*h = old[0 : n-1]
	return order
}

// MatchingEngine es el motor de emparejamiento
// que empareja las órdenes de compra y venta
type MatchingEngine struct {
	// BuyOrders y SellOrders son heaps de órdenes
	BuyOrders  *BuyHeap
	SellOrders *SellHeap

	// Mutexes para proteger los heaps
	// de acceso concurrente
	/*
		(mutual exclusion) que garantizará que nuestro
		código no acceda a una variable hasta que
		nosotros le indiquemos, evitando que se den
		las condiciones de carrera o race conditions.
	*/
	buyMutex  sync.Mutex
	sellMutex sync.Mutex

	// orderChannel es un canal para enviar órdenes
	// al motor de emparejamiento
	orderChannel chan *Order
}

// NewMatchingEngine crea un nuevo MatchingEngine
// y comienza a procesar las órdenes
// en un gorutina separada
func NewMatchingEngine() *MatchingEngine {

	// Crear un nuevo MatchingEngine
	// con heaps de compras y ventas
	me := &MatchingEngine{
		BuyOrders:  &BuyHeap{},
		SellOrders: &SellHeap{},

		// Inicializar el canal de órdenes
		// para enviar órdenes al motor de emparejamiento
		// desde otras gorutinas
		orderChannel: make(chan *Order),
	}

	// Comenzar a procesar las órdenes
	// en un gorutina separada
	go me.processOrders()

	return me
}

// processOrders es un bucle infinito que procesa
// las órdenes enviadas al motor de emparejamiento
// a través del canal orderChannel
func (me *MatchingEngine) processOrders() {

	// Recorrer todas las órdenes enviadas al canal
	for order := range me.orderChannel {

		// Agregar la orden al heap correspondiente
		me.addOrderInternal(order)

		// Intentar emparejar las órdenes
		//me.MatchOrders()
	}
}

// AddOrder agrega una orden al canal de órdenes
// para que sea procesada por el motor de emparejamiento
func (me *MatchingEngine) AddOrder(order *Order) {
	me.orderChannel <- order
}

// addOrderInternal agrega una orden al heap correspondiente
// y actualiza el heap
func (me *MatchingEngine) addOrderInternal(order *Order) {

	if order.Type == Buy {
		// Proteger el heap de compras
		me.buyMutex.Lock()

		// Agregar la orden al heap de compras
		heap.Push(me.BuyOrders, order)

		// Desbloquear el heap de compras
		me.buyMutex.Unlock()
	} else {
		me.sellMutex.Lock()
		heap.Push(me.SellOrders, order)
		me.sellMutex.Unlock()
	}
}

func (me *MatchingEngine) MatchOrders() {
	startTime := time.Now()
	matchesFound := 0
	iterationCount := 0

	fmt.Printf("\n=== INICIO DEL MATCHING ===\n")
	fmt.Printf("Estado inicial: %d compras, %d ventas\n", me.BuyOrders.Len(), me.SellOrders.Len())
	fmt.Printf("Tiempo de inicio: %s\n", startTime)

	for {
		iterationCount++
		fmt.Printf("\n--- Iteración %d ---\n", iterationCount)

		// Obtener la mejor orden de compra
		me.buyMutex.Lock()
		if me.BuyOrders.Len() == 0 {
			fmt.Printf("No hay órdenes de compra disponibles\n")
			me.buyMutex.Unlock()
			break
		}
		buyLen := me.BuyOrders.Len()
		bestBuy := heap.Pop(me.BuyOrders).(*Order)
		fmt.Printf("Compras antes/después del Pop: %d/%d\n", buyLen, me.BuyOrders.Len())
		me.buyMutex.Unlock()

		// Obtener la mejor orden de venta
		me.sellMutex.Lock()
		if me.SellOrders.Len() == 0 {
			fmt.Printf("No hay órdenes de venta disponibles, devolviendo compra al heap\n")
			me.sellMutex.Unlock()
			me.buyMutex.Lock()
			prevLen := me.BuyOrders.Len()
			heap.Push(me.BuyOrders, bestBuy)
			fmt.Printf("Compras antes/después del Push: %d/%d\n", prevLen, me.BuyOrders.Len())
			me.buyMutex.Unlock()
			break
		}
		sellLen := me.SellOrders.Len()
		bestSell := heap.Pop(me.SellOrders).(*Order)
		fmt.Printf("Ventas antes/después del Pop: %d/%d\n", sellLen, me.SellOrders.Len())
		me.sellMutex.Unlock()

		fmt.Printf("Comparando: Compra ID=%d Precio=%.2f Cantidad=%d vs Venta ID=%d Precio=%.2f Cantidad=%d\n",
			bestBuy.ID, bestBuy.Price, bestBuy.Amount, bestSell.ID, bestSell.Price, bestSell.Amount)

		// Verificar si los precios coinciden
		if bestBuy.Price >= bestSell.Price {
			tradeAmount := min(bestBuy.Amount, bestSell.Amount)
			matchesFound++

			fmt.Printf("\nMATCH #%d: Compra %d @ %.2f vs Venta %d @ %.2f, Cantidad=%d\n",
				matchesFound, bestBuy.ID, bestBuy.Price, bestSell.ID, bestSell.Price, tradeAmount)

			fmt.Printf("Cantidades antes del trade - Compra: %d, Venta: %d\n",
				bestBuy.Amount, bestSell.Amount)

			// Actualizar cantidades
			bestBuy.Amount -= tradeAmount
			bestSell.Amount -= tradeAmount

			fmt.Printf("Cantidades después del trade - Compra: %d, Venta: %d\n",
				bestBuy.Amount, bestSell.Amount)

			// Solo reinsertar órdenes si tienen cantidad restante
			if bestBuy.Amount > 0 {
				me.buyMutex.Lock()
				prevLen := me.BuyOrders.Len()
				heap.Push(me.BuyOrders, bestBuy)
				fmt.Printf("Reinsertando compra %d - Heap antes/después: %d/%d\n",
					bestBuy.ID, prevLen, me.BuyOrders.Len())
				me.buyMutex.Unlock()
			} else {
				fmt.Printf("Orden de compra %d completada y eliminada\n", bestBuy.ID)
			}

			if bestSell.Amount > 0 {
				me.sellMutex.Lock()
				prevLen := me.SellOrders.Len()
				heap.Push(me.SellOrders, bestSell)
				fmt.Printf("Reinsertando venta %d - Heap antes/después: %d/%d\n",
					bestSell.ID, prevLen, me.SellOrders.Len())
				me.sellMutex.Unlock()
			} else {
				fmt.Printf("Orden de venta %d completada y eliminada\n", bestSell.ID)
			}
		} else {
			fmt.Printf("\nNo hay match: Precio compra %.2f < Precio venta %.2f\n",
				bestBuy.Price, bestSell.Price)

			me.buyMutex.Lock()
			prevBuyLen := me.BuyOrders.Len()
			heap.Push(me.BuyOrders, bestBuy)
			fmt.Printf("Devolviendo compra %d - Heap antes/después: %d/%d\n",
				bestBuy.ID, prevBuyLen, me.BuyOrders.Len())
			me.buyMutex.Unlock()

			me.sellMutex.Lock()
			prevSellLen := me.SellOrders.Len()
			heap.Push(me.SellOrders, bestSell)
			fmt.Printf("Devolviendo venta %d - Heap antes/después: %d/%d\n",
				bestSell.ID, prevSellLen, me.SellOrders.Len())
			me.sellMutex.Unlock()
			break
		}

		fmt.Printf("\nEstado actual: %d compras, %d ventas\n",
			me.BuyOrders.Len(), me.SellOrders.Len())

		// Verificar si quedan órdenes suficientes para continuar
		if me.BuyOrders.Len() == 0 || me.SellOrders.Len() == 0 {
			fmt.Printf("Terminando: No hay suficientes órdenes para continuar\n")
			break
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n=== FIN DEL MATCHING ===\n")
	fmt.Printf("Tiempo total: %s\n", duration)
	fmt.Printf("Matches realizados: %d\n", matchesFound)
	fmt.Printf("Iteraciones totales: %d\n", iterationCount)
	fmt.Printf("Órdenes restantes: %d compras, %d ventas\n",
		me.BuyOrders.Len(), me.SellOrders.Len())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (me *MatchingEngine) executeTrade(buyOrder *Order, sellOrder *Order, amount int, duration time.Duration) {
	//Lógica para ejecutar el trade.
	fmt.Printf("Trade ejecutado. Compra: %d, Venta: %d, Cantidad: %d, Precio: %.2f, Duración: %s\n", buyOrder.ID, sellOrder.ID, amount, sellOrder.Price, duration)
}

func AddOrderHandler(engine *MatchingEngine) fiber.Handler {
	return func(c *fiber.Ctx) error {
		order := new(Order)
		if err := json.Unmarshal(c.Body(), order); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order format"})
		}

		engine.AddOrder(order)
		return c.SendStatus(http.StatusCreated)
	}
}

func (me *MatchingEngine) StartMatching() {
	go func() {
		for {
			me.MatchOrders()
			time.Sleep(50 * time.Millisecond) // Pausa para no consumir todos los recursos
		}
	}()
}

func (me *MatchingEngine) GetHighestBuyOrder() *Order {
	me.buyMutex.Lock()
	defer me.buyMutex.Unlock()

	if me.BuyOrders.Len() > 0 {
		// Clonar la orden para evitar modificaciones externas
		order := *(*me.BuyOrders)[0]
		return &order
	}
	return nil
}

func (me *MatchingEngine) GetHighestSellOrder() *Order {
	me.sellMutex.Lock()
	defer me.sellMutex.Unlock()

	if me.SellOrders.Len() > 0 {
		// Clonar la orden para evitar modificaciones externas
		order := *(*me.SellOrders)[0]
		return &order
	}
	return nil
}

func GetHighestBuyOrderHandler(engine *MatchingEngine) fiber.Handler {
	return func(c *fiber.Ctx) error {
		order := engine.GetHighestBuyOrder()
		if order == nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No hay órdenes de compra disponibles"})
		}
		return c.JSON(order)
	}
}

func GetHighestSellOrderHandler(engine *MatchingEngine) fiber.Handler {
	return func(c *fiber.Ctx) error {
		order := engine.GetHighestSellOrder()
		if order == nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "No hay órdenes de venta disponibles"})
		}
		return c.JSON(order)
	}
}

func main() {

	app := fiber.New()
	engine := &MatchingEngine{
		BuyOrders:    &BuyHeap{},
		SellOrders:   &SellHeap{},
		orderChannel: make(chan *Order),
	}

	go engine.processOrders()
	engine.StartMatching()

	app.Post("/orders", AddOrderHandler(engine))

	app.Get("/orders", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"buys":  engine.BuyOrders,
			"sells": engine.SellOrders,
		})
	})

	//optener la orden de compra con el precio más alto
	app.Get("/highest-buy-order", GetHighestBuyOrderHandler(engine))

	app.Get("/highest-sell-order", GetHighestSellOrderHandler(engine))

	// Esperar un poco para que se procesen las órdenes
	time.Sleep(1 * time.Second)

	app.Listen(":3000")

}
