// main.go
package main

import (
	"os"
	"time"
)

var (
	chPortal chan struct{}
	detChans []chan struct{}
)

func moveEmDirecao(jogo *Jogo, x, y *int) {
	if jogo.PosX > *x && jogoPodeMoverPara(jogo, *x+1, *y) {
		*x++
	} else if jogo.PosX < *x && jogoPodeMoverPara(jogo, *x-1, *y) {
		*x--
	}
	if jogo.PosY > *y && jogoPodeMoverPara(jogo, *x, *y+1) {
		*y++
	} else if jogo.PosY < *y && jogoPodeMoverPara(jogo, *x, *y-1) {
		*y--
	}
}

func main() {
	interfaceIniciar()
	defer interfaceFinalizar()

	// canais globais
	chPortal = make(chan struct{})
	// Carrega o mapa
	mapa := "mapa.txt"
	if len(os.Args) > 1 {
		mapa = os.Args[1]
	}
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapa, &jogo); err != nil {
		panic(err)
	}

	// adiciona inimigos de patrulha e seguranças
	adicionarInimigos(&jogo)
	adicionarSegurancas(&jogo)

	// preparamos canais de detecção para os dois inimigos originais
	totalPatrol := len(jogo.Inimigos)
	detChans = make([]chan struct{}, totalPatrol)
	for i := 0; i < totalPatrol; i++ {
		detChans[i] = make(chan struct{})
	}

	// adiciona seguidor
	segIdx := len(jogo.Inimigos)
	jogo.Inimigos = append(jogo.Inimigos,
		Inimigo{Elemento: novoElemento('☻', CorVerde, CorPadrao, true), PosX: 1, PosY: 1, Visivel: true, PatrulhaDir: 1, Vida: 3},
	)

	// inicia loops de inimigos
	for i := 0; i < totalPatrol; i++ {
		chPat := make(chan string)
		chRea := make(chan string)
		go jogo.Inimigos[i].loopConcorrente(&jogo, chPat, chRea, detChans[i])

		// patrulha periódica
		go func(c chan string) {
			for {
				time.Sleep(2 * time.Second)
				c <- "patrulhar"
			}
		}(chPat)

		// ataque ao colidir
		go func(c chan string, inim *Inimigo) {
			for {
				jogo.Mutex.Lock()
				px, py := jogo.PosX, jogo.PosY
				gx, gy := inim.PosX, inim.PosY
				jogo.Mutex.Unlock()
				if px == gx && py == gy {
					c <- "atacar"
				}
				time.Sleep(200 * time.Millisecond)
			}
		}(chRea, &jogo.Inimigos[i])
	}

	// seguidor
	go func() {
		for {
			jogo.Mutex.Lock()
			moveEmDirecao(&jogo, &jogo.Inimigos[segIdx].PosX, &jogo.Inimigos[segIdx].PosY)
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(&jogo)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// item autônomo com canal + timeout
	jogo.Items = append(jogo.Items, Item{Simbolo: '♦', Visivel: false, PosX: 10, PosY: 5})
	go func() {
		select {
		case <-chPortal: // aparece ao interagir
		case <-time.After(5 * time.Second): // ou por timeout
		}
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = true
		jogo.Mutex.Unlock()
		interfaceDesenharJogo(&jogo)

		<-time.After(10 * time.Second) // depois some
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = false
		jogo.Mutex.Unlock()
		interfaceDesenharJogo(&jogo)
	}()

	// loop principal
	interfaceDesenharJogo(&jogo)
	for {
		ev := interfaceLerEventoTeclado()
		if !personagemExecutarAcao(ev, &jogo) {
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}
