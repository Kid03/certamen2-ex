package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

/*
	Se tomo el tiempo por parte del cliente, ya que es el quien
	trae el tramite y dara el tiempo de atencion, el cajero
	solo se limitara servirlo en sus tramites.
*/
type cliente struct {
	id       int
	tTramite int //Tiempo de tramite
}

type cajero struct {
	id         int
	atendiendo *cliente
	last       time.Time
	s          corrutine
}

// Estructura que modela las Corrutinas
type corrutine struct {
	signal chan struct{}
}

func (s *corrutine) contextChange() {
	s.signal <- struct{}{}
}
func (s *corrutine) resumenExecution() {
	<-s.signal
}

// ------------------------------------
type manager struct {
	cajeros   []cajero  // Lista de cajas habilitadas
	clientes  []cliente // Lista de clientes llegando
	s         corrutine // Wakeup and wait
	inter     corrutine // internal
	last      time.Time // hora del ultimo en arribar
	idCliente int       // Varible que registra el numero de clientes atendidos
}

func (m *manager) cashierOpener(n int) {
	m.s.resumenExecution()
	rand.Seed(time.Now().Unix()) // Genera una semilla para el tiempo
	// Crea a los cajeros y su linea de comunicacion
	for i := 0; i < n; i++ {
		m.cajeros = append(m.cajeros, cajero{id: i, atendiendo: nil, s: corrutine{signal: make(chan struct{})}})
		fmt.Println("Se abre la caja ", i)
	}
	// ---------------------------------------------
	m.s.contextChange()
}

func (m *manager) reception() {
	for {
		m.s.resumenExecution()
		//Llegada de clientes
		m.inter.contextChange()
		m.inter.resumenExecution()
		//-------------------
		// Recorre todas las cajas buscando una habilitada para la atencion mientras que llegan clientes
		for i := 0; i < len(m.cajeros); i++ {
			// Llegada de clientes
			m.inter.contextChange()
			m.inter.resumenExecution()
			// -------------------
			if m.cajeros[i].atendiendo == nil { // Pregunta si el cajero esta atendiendo clientes o no
				if len(m.clientes) > 0 { // Pregunta si hay clientes
					fmt.Println("El cliente ", m.clientes[0].id, " sera atendido por el cajero ", m.cajeros[i].id, " y tomara ", m.clientes[0].tTramite, "sec")
					go m.cajeros[i].serve(&m.clientes[0])
					m.clientes = m.clientes[1:]
					m.cajeros[i].s.contextChange()
					m.cajeros[i].s.resumenExecution()

				}
			} else {
				// Pregunta si dejo de atender al cliente
				m.cajeros[i].s.contextChange()
				m.cajeros[i].s.resumenExecution()
				// --------------------------------------
			}

		}
		// ----------------------------------------------------------------------------------------------
		m.s.contextChange()
	}
}

func (m *manager) upComingClient(waitTime int) {
	for {
		m.inter.resumenExecution()
		if m.last.Before(time.Now()) {
			fmt.Println("Llego el cliente ticket ", m.idCliente)
			m.clientes = append(m.clientes, cliente{id: m.idCliente, tTramite: rand.Intn(9) + 1})
			m.idCliente++
			m.last = time.Now().Add(time.Second * time.Duration(waitTime))
		} else {
			//fmt.Println("Aun no llega cliente")
		}
		m.inter.contextChange()
	}
}

func (c *cajero) serve(a *cliente) {
	c.s.resumenExecution() //Corrutina in
	//Cambia el tiempo de termino para la atencion
	c.last = time.Now().Add((time.Second * time.Duration(a.tTramite)))
	//--------------------------------------------
	//toma al cliente que debe atender
	c.atendiendo = a
	//--------------------------------
	c.s.contextChange()    //Corrutina out
	c.s.resumenExecution() //Corrutina in
	for {
		//Pregunta si el tiempo presente es pasado el tiempo limite de atencion
		if c.last.Before(time.Now()) {
			// Procede a dejar la atencion del cliente
			fmt.Println("El cajero ", c.id, " termino su atencion de ", a.id)
			c.atendiendo = nil
			break
			// ---------------------------------------
		}
		c.s.contextChange()    //Corrutina out
		c.s.resumenExecution() //Corrutina in
	}
	defer c.s.contextChange() //Corrutina out
}

func main() {
	// Se reciben parametros desde cmd
	numCajeros, _ := strconv.Atoi(os.Args[1])
	tiempoEspera, _ := strconv.Atoi(os.Args[2])
	//--------------------------------
	var m manager = manager{s: corrutine{signal: make(chan struct{})}, inter: corrutine{signal: make(chan struct{})}, idCliente: 0}
	// Abre las cajas
	go m.cashierOpener(numCajeros)
	m.s.contextChange()
	m.s.resumenExecution()
	// ---------------
	// Se inicia el proceso para recivir clientes
	fmt.Println("Se abre la maquina de tickets de atencion a los clientes")
	go m.upComingClient(tiempoEspera)
	// ------------------------------------------
	// Se inicia el procesado de clientes
	fmt.Println("Se abren las cajas")
	go m.reception()
	for {
		m.s.contextChange()
		m.s.resumenExecution()
	}
	// ----------------------------------
}
