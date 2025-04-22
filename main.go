package main

import (
	"os"
	"time"
)

// Move o seguidor em direção ao jogador
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

	mapa := "mapa.txt"
	if len(os.Args) > 1 {
		mapa = os.Args[1]
	}

	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapa, &jogo); err != nil {
		panic(err)
	}

	adicionarInimigos(&jogo)
	adicionarSegurancas(&jogo)

	for i := range jogo.Inimigos {
		chPat, chRea := make(chan string), make(chan string)
		go jogo.Inimigos[i].loopConcorrente(&jogo, chPat, chRea)
		go func(c chan string) {
			for {
				time.Sleep(3 * time.Second)
				c <- "patrulhar"
			}
		}(chPat)
		go func(c chan string, idx int) {
			for {
				if jogo.PosX == jogo.Inimigos[idx].PosX && jogo.PosY == jogo.Inimigos[idx].PosY {
					c <- "atacar"
				}
				time.Sleep(1 * time.Second)
			}
		}(chRea, i)
	}

	segIdx := len(jogo.Inimigos)
	jogo.Inimigos = append(jogo.Inimigos,
		Inimigo{Elemento: novoElemento('☻', CorVerde, CorPadrao, true), PosX: 1, PosY: 1, Visivel: true, PatrulhaDir: 1, Vida: 3},
	)
	go func() {
		for {
			jogo.Mutex.Lock()
			moveEmDirecao(&jogo, &jogo.Inimigos[segIdx].PosX, &jogo.Inimigos[segIdx].PosY)
			interfaceDesenharJogo(&jogo)
			jogo.Mutex.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	jogo.Items = append(jogo.Items, Item{Simbolo: '♦', Visivel: false, PosX: 10, PosY: 5})
	go func() {
		time.Sleep(5 * time.Second)
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = true
		interfaceDesenharJogo(&jogo)
		jogo.Mutex.Unlock()

		time.Sleep(10 * time.Second)
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = false
		interfaceDesenharJogo(&jogo)
		jogo.Mutex.Unlock()
	}()
	go func() {
		<-time.After(10 * time.Second)
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = false
		interfaceDesenharJogo(&jogo)
		jogo.Mutex.Unlock()
	}()

	interfaceDesenharJogo(&jogo)
	for {
		ev := interfaceLerEventoTeclado()
		if !personagemExecutarAcao(ev, &jogo) {
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}
