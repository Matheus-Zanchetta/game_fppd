package main

import (
	"os"
	"time"
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

	// Carrega o mapa
	mapa := "mapa.txt"
	if len(os.Args) > 1 {
		mapa = os.Args[1]
	}
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapa, &jogo); err != nil {
		panic(err)
	}

	// Adiciona inimigos de patrulha e seguranças
	adicionarInimigos(&jogo)
	adicionarSegurancas(&jogo)
	// Índice do futuro seguidor
	segIdx := len(jogo.Inimigos)
	// Adiciona o seguidor
	jogo.Inimigos = append(jogo.Inimigos,
		Inimigo{Elemento: novoElemento('☻', CorVerde, CorPadrao, true), PosX: 1, PosY: 1, Visivel: true, PatrulhaDir: 1, Vida: 3},
	)

	// Inicia geloConcorrente para todos exceto o seguidor
	totalPatrol := segIdx
	for i := 0; i < totalPatrol; i++ {
		chPat := make(chan string)
		chRea := make(chan string)
		go jogo.Inimigos[i].loopConcorrente(&jogo, chPat, chRea)
		// Comando patrulha
		go func(c chan string) {
			for {
				time.Sleep(2 * time.Second)
				c <- "patrulhar"
			}
		}(chPat)
		// Comando ataque
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

	// Inicia goroutine do seguidor
	go func() {
		for {
			jogo.Mutex.Lock()
			moveEmDirecao(&jogo, &jogo.Inimigos[segIdx].PosX, &jogo.Inimigos[segIdx].PosY)
			interfaceDesenharJogo(&jogo)
			jogo.Mutex.Unlock()
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Item autônomo
	jogo.Items = append(jogo.Items, Item{Simbolo: '♦', Visivel: false, PosX: 10, PosY: 5})
	go func() {
		<-time.After(5 * time.Second)
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = true
		interfaceDesenharJogo(&jogo)
		jogo.Mutex.Unlock()

		<-time.After(10 * time.Second)
		jogo.Mutex.Lock()
		jogo.Items[0].Visivel = false
		interfaceDesenharJogo(&jogo)
		jogo.Mutex.Unlock()
	}()

	// Loop principal
	interfaceDesenharJogo(&jogo)
	for {
		ev := interfaceLerEventoTeclado()
		if !personagemExecutarAcao(ev, &jogo) {
			break
		}
		interfaceDesenharJogo(&jogo)
	}
}
