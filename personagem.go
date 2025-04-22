package main

type EventoTeclado struct {
	Tipo  string
	Tecla rune
}

func personagemMover(tecla rune, jogo *Jogo) {
	jogo.Mutex.Lock()
	dx, dy := 0, 0
	switch tecla {
	case 'w':
		dy = -1
	case 'a':
		dx = -1
	case 's':
		dy = 1
	case 'd':
		dx = 1
	}
	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogo.Mapa[jogo.PosY][jogo.PosX] = jogo.UltimoVisitado
		jogo.UltimoVisitado = jogo.Mapa[ny][nx]
		jogo.Mapa[ny][nx] = Personagem
		jogo.PosX, jogo.PosY = nx, ny
	}
	jogo.Mutex.Unlock()

	// notifica inimigos próximos (distância ≤2) para perseguir
	for i, ch := range detChans {
		inim := jogo.Inimigos[i]
		if abs(jogo.PosX-inim.PosX) <= 2 && abs(jogo.PosY-inim.PosY) <= 2 {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}
}

func personagemInteragir(jogo *Jogo) {
	// sinaliza item oculto
	select {
	case chPortal <- struct{}{}:
	default:
	}
	jogo.Mutex.Lock()
	jogo.StatusMsg = "Interagindo..."
	jogo.Mutex.Unlock()
}

func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		return false
	case "interagir":
		personagemInteragir(jogo)
	case "mover":
		personagemMover(ev.Tecla, jogo)
	}
	return true
}
