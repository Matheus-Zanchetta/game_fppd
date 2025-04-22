package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

type Jogo struct {
	Mapa           [][]Elemento
	PosX, PosY     int
	UltimoVisitado Elemento
	StatusMsg      string
	VidaPersonagem int
	Items          []Item
	Inimigos       []Inimigo
	Mutex          sync.Mutex
}

func jogoNovo() Jogo {
	return Jogo{
		Mapa:           [][]Elemento{},
		PosX:           0,
		PosY:           0,
		UltimoVisitado: Vazio,
		StatusMsg:      "Jogo Iniciado",
		VidaPersonagem: 5,
		Items:          []Item{},
		Inimigos:       []Inimigo{},
	}
}

type Item struct {
	Simbolo    rune
	Visivel    bool
	PosX, PosY int
}

type Inimigo struct {
	Elemento
	PosX, PosY  int
	Visivel     bool
	PatrulhaDir int
	Vida        int
}

var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
)

func novoElemento(sim rune, cor, corFundo Cor, tang bool) Elemento {
	return Elemento{simbolo: sim, cor: cor, corFundo: corFundo, tangivel: tang}
}

func adicionarInimigos(jogo *Jogo) {
	inim1 := Inimigo{Elemento: novoElemento('⚔', CorVermelho, CorPadrao, true), PosX: 5, PosY: 5, Visivel: true, PatrulhaDir: 1, Vida: 3}
	inim2 := Inimigo{Elemento: novoElemento('⚔', CorVermelho, CorPadrao, true), PosX: 10, PosY: 7, Visivel: true, PatrulhaDir: 1, Vida: 3}
	jogo.Inimigos = append(jogo.Inimigos, inim1, inim2)
}

func adicionarSegurancas(jogo *Jogo) {
	g1 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 15, PosY: 5, Visivel: true, PatrulhaDir: 1, Vida: 5}
	g2 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 20, PosY: 7, Visivel: true, PatrulhaDir: 1, Vida: 5}
	jogo.Inimigos = append(jogo.Inimigos, g1, g2)
}

// loopConcorrente agora verifica colisão com paredes tanto na patrulha quanto na perseguição
func (i *Inimigo) loopConcorrente(jogo *Jogo, chPat, chRea chan string, chDet chan struct{}) {
	pursuing := false
	for {
		select {
		case <-chDet:
			pursuing = true

		case <-chPat:
			jogo.Mutex.Lock()
			if pursuing {
				// perseguir jogador sem atravessar paredes
				dx := jogo.PosX - i.PosX
				dy := jogo.PosY - i.PosY
				// movimento horizontal
				if dx != 0 {
					nx := i.PosX + dx/abs(dx)
					if jogoPodeMoverPara(jogo, nx, i.PosY) {
						i.PosX = nx
					}
				}
				// movimento vertical
				if dy != 0 {
					ny := i.PosY + dy/abs(dy)
					if jogoPodeMoverPara(jogo, i.PosX, ny) {
						i.PosY = ny
					}
				}
			} else {
				// patrulha normal sem atravessar paredes
				nx := i.PosX + i.PatrulhaDir
				if !jogoPodeMoverPara(jogo, nx, i.PosY) {
					i.PatrulhaDir *= -1
					nx = i.PosX + i.PatrulhaDir
				}
				if jogoPodeMoverPara(jogo, nx, i.PosY) {
					i.PosX = nx
				}
			}
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)

		case <-chRea:
			jogo.Mutex.Lock()
			if jogo.PosX == i.PosX && jogo.PosY == i.PosY {
				jogo.VidaPersonagem--
				jogo.StatusMsg = fmt.Sprintf("Você foi atingido! Vida: %d", jogo.VidaPersonagem)
			}
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func jogoCarregarMapa(nome string, jogo *Jogo) error {
	f, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		row := make([]Elemento, 0, len(linha))
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Vegetacao.simbolo:
				e = Vegetacao
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y
			}
			row = append(row, e)
		}
		jogo.Mapa = append(jogo.Mapa, row)
		y++
	}
	return scanner.Err()
}

func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) || x < 0 || x >= len(jogo.Mapa[0]) {
		return false
	}
	return !jogo.Mapa[y][x].tangivel
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}
