package main

import (
	"fmt"
	"sync"
	"time"
)

// Contador global
var cont int = 1

type corrutine struct {
	ch chan struct{}
}

// Pausa la corrutina esperando una respuesta para continuar
func (c *corrutine) wait() {
	for {
		if _, ok := <-c.ch; ok {
			break
		}
	}
}

// Reanuda la corrutina mandando una señal al channel
func (c *corrutine) resume() {
	c.ch <- struct{}{}
}

var wg = sync.WaitGroup{}
var oddCh = corrutine{ch: make(chan struct{})}
var evenCh = corrutine{ch: make(chan struct{})}

func main() {

	// Espera a que se ejecuten 11 - 1 ciclos
	wg.Add(11)

	go controller()
	go oddPrinter(oddCh)
	go evenPrinter(evenCh)

	// Inicia la corrutina
	oddCh.resume()
	wg.Wait()
}

// Muestra por pantalla números pares
func oddPrinter(ch corrutine) {
	for {
		ch.wait()
		fmt.Println("odd: ", cont)
		cont++
		ch.resume()
	}
}

// Muestra por pantalla números impares
func evenPrinter(ch corrutine) {
	for {
		ch.wait()
		fmt.Println("even: ", cont)
		cont++
		ch.resume()
	}
}

// Controlador principal que permuta señales entre canales
func controller() {
	for {
		select {
		// Recibe señal por parte de los números pares
		case i := <-oddCh.ch:
			wg.Done()
			time.Sleep(time.Second * 1)

			// Envía señal a canal de números pares para indicar que es su turno
			evenCh.ch <- i

		// Recibe señal por parte de los números impares
		case i := <-evenCh.ch:
			wg.Done()
			time.Sleep(time.Second * 1)

			// Envía señal a canal de números impares para indicar que es su turno
			oddCh.ch <- i
		}
	}
}
