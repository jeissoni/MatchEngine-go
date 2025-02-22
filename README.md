# Motor de Emparejamiento de Órdenes

Este proyecto implementa un motor de emparejamiento de órdenes en Go, diseñado para procesar órdenes de compra y venta de manera eficiente y concurrente.

## Características

* **Emparejamiento de Órdenes:** Empareja órdenes de compra y venta según el precio y la cantidad.
* **Heaps para Órdenes:** Utiliza heaps (`BuyHeap` y `SellHeap`) para mantener las órdenes ordenadas por precio.
* **Concurrencia:** Implementa concurrencia mediante mutexes y canales para manejar múltiples órdenes simultáneamente.
* **API HTTP:** Expone un API HTTP con Fiber para recibir órdenes y consultar el estado del motor.
* **Pruebas de Carga:** Incluye pruebas de carga para verificar el rendimiento del motor en diferentes escenarios.


## Funcionamiento de los Heaps

Este motor de emparejamiento utiliza heaps (`BuyHeap` y `SellHeap`) para mantener las órdenes ordenadas por precio, lo que permite un emparejamiento eficiente.

### ¿Qué es un Heap?

Un heap es una estructura de datos de árbol binario especializada que cumple con la propiedad del heap:

* **Heap Máximo (`BuyHeap`):** En un heap máximo, el valor de cada nodo padre es mayor o igual que el valor de sus nodos hijos. Se utiliza para las órdenes de compra, donde los precios más altos tienen prioridad.
* **Heap Mínimo (`SellHeap`):** En un heap mínimo, el valor de cada nodo padre es menor o igual que el valor de sus nodos hijos. Se utiliza para las órdenes de venta, donde los precios más bajos tienen prioridad.

### Funcionamiento de los Heaps en este Proyecto

1.  **Inserción (Push):**
    * Cuando se agrega una nueva orden al heap, se coloca al final del árbol.
    * Luego, se compara con su nodo padre. Si viola la propiedad del heap, se intercambian.
    * Este proceso se repite hasta que la orden esté en la posición correcta.

2.  **Extracción (Pop):**
    * Cuando se extrae la orden con el precio más alto (en `BuyHeap`) o el precio más bajo (en `SellHeap`), se elimina el nodo raíz.
    * El último nodo del árbol se mueve a la raíz.
    * Luego, se compara con sus nodos hijos. Si viola la propiedad del heap, se intercambian.
    * Este proceso se repite hasta que el árbol vuelva a cumplir con la propiedad del heap.

3.  **Ordenamiento:**
    * La función `Less` determina el orden en que se organizará el heap.
    * La función `Swap` intercambia las posiciones de los elementos, y es aquí donde se deben actualizar los índices de los elementos.

### ¿Por qué usar Heaps?

* **Eficiencia:** Los heaps permiten acceder rápidamente a la mejor orden de compra o venta disponible.
* **Mantenimiento del Orden:** Los heaps mantienen automáticamente las órdenes ordenadas por precio.
* **Complejidad:** Las operaciones de inserción y extracción en un heap tienen una complejidad de tiempo de O(log n), donde n es el número de elementos en el heap.

En resumen, los heaps son una estructura de datos eficiente para mantener las órdenes ordenadas por precio y permitir un acceso rápido a las mejores órdenes disponibles.


## Concurrencia y Manejo de Bloqueos

Este motor de emparejamiento está diseñado para manejar múltiples órdenes de forma concurrente, lo que requiere una gestión cuidadosa de los recursos compartidos para evitar condiciones de carrera y garantizar la integridad de los datos.

### Canales (`orderChannel`)

* Se utiliza un canal (`orderChannel`) para recibir órdenes de forma asíncrona.
* Las órdenes se envían al canal desde el handler de la API (`AddOrderHandler`).
* Una goroutine separada (`processOrders`) consume las órdenes del canal y las agrega a los heaps correspondientes (`BuyHeap` o `SellHeap`).
* El uso de un canal permite desacoplar la recepción de órdenes del procesamiento, lo que mejora la concurrencia y la capacidad de respuesta del motor.

### Bloqueos (`sync.Mutex`)

* Se utilizan mutexes (`buyMutex` y `sellMutex`) para proteger el acceso a los heaps (`BuyHeap` y `SellHeap`).
* Los mutexes garantizan que solo una goroutine pueda acceder a un heap en un momento dado, lo que evita condiciones de carrera y asegura la integridad de los datos.
* Los bloqueos se aplican en las siguientes operaciones:
    * Inserción de órdenes en los heaps (`heap.Push`).
    * Extracción de órdenes de los heaps (`heap.Pop`).
    * Lectura y modificación de los heaps en la función `MatchOrders`.
* La granularidad de los bloqueos se ha optimizado para minimizar el tiempo de bloqueo y maximizar la concurrencia.
* Se ha tenido especial cuidado en desbloquear los mutexes en todos los casos, incluso en situaciones de error, para evitar bloqueos permanentes.

### Gestión de la Concurrencia en `MatchOrders`

* La función `MatchOrders` se ejecuta en una goroutine separada para realizar el emparejamiento de órdenes de forma continua.
* Se aplican bloqueos a los heaps durante la lectura y modificación de las órdenes para garantizar la integridad de los datos.
* Se utiliza un bucle infinito para procesar las órdenes hasta que no haya más órdenes disponibles o no se encuentren emparejamientos.
* Se ha implementado una lógica para manejar órdenes parciales (órdenes con cantidades restantes) y volver a insertarlas en los heaps.

En resumen, se han utilizado canales y mutexes para gestionar la concurrencia y los bloqueos de forma eficiente, lo que garantiza la integridad de los datos y el rendimiento del motor de emparejamiento.

## Requisitos

* Go 1.16 o superior
* Fiber v2
* Paquete `container/heap` de Go

## Instalación

1.  Clona el repositorio:

    ```bash
    git clone <URL_del_repositorio>
    ```

2.  Navega al directorio del proyecto:

    ```bash
    cd <nombre_del_directorio>
    ```

3.  Descarga las dependencias:

    ```bash
    go mod tidy
    ```

## Uso

1.  Ejecuta la aplicación:

    ```bash
    go run main.go
    ```

2.  Envía órdenes al endpoint `/orders` mediante peticiones POST con el siguiente formato JSON:

    ```json
    {
      "ID": 1,
      "Type": "BUY",
      "Price": 100.0,
      "Amount": 10
    }
    ```

3.  Consulta la orden de compra con el precio más alto en el endpoint `/highest-buy-order` mediante peticiones GET.

## Pruebas

Para ejecutar las pruebas de carga:

1.  Asegúrate de que la aplicación esté en ejecución.
2.  Mover a la carpeta `Test`
3.  Ejecuta el script de pruebas de carga:



    ```bash
    go run main.go
    ```
