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
	PatrulhaDir int // +1 ou -1
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
	jogo.Inimigos = append(jogo.Inimigos,
		Inimigo{Elemento: Parede, PosX: 5, PosY: 5, Visivel: true, PatrulhaDir: 1, Vida: 3},
		Inimigo{Elemento: Vegetacao, PosX: 10, PosY: 7, Visivel: true, PatrulhaDir: 1, Vida: 3},
	)
}

func adicionarSegurancas(jogo *Jogo) {
	g1 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 5, PosY: 5, Visivel: true, PatrulhaDir: 1, Vida: 5}
	g2 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 10, PosY: 7, Visivel: true, PatrulhaDir: 1, Vida: 5}
	jogo.Inimigos = append(jogo.Inimigos, g1, g2)
}

func (i *Inimigo) loopConcorrente(jogo *Jogo, chPat, chRea chan string) {
	for {
		select {
		case <-chPat:
			jogo.Mutex.Lock()
			nx := i.PosX + i.PatrulhaDir
			if !jogoPodeMoverPara(jogo, nx, i.PosY) {
				i.PatrulhaDir *= -1
				nx = i.PosX + i.PatrulhaDir
			}
			if jogoPodeMoverPara(jogo, nx, i.PosY) {
				i.PosX = nx
			}
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)

		case <-chRea:
			jogo.Mutex.Lock()
			if jogo.PosX == i.PosX && jogo.PosY == i.PosY {
				jogo.VidaPersonagem--
				jogo.StatusMsg = fmt.Sprintf("Voc\u00ea foi atingido por um guarda! Vida: %d", jogo.VidaPersonagem)
			}
			jogo.Mutex.Unlock()
			interfaceDesenharJogo(jogo)
		}
		time.Sleep(100 * time.Millisecond)
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
