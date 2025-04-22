package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

// Elemento representa célula, personagem ou inimigo
type Elemento struct {
	simbolo  rune
	cor      Cor
	corFundo Cor
	tangivel bool
}

// Jogo contém todo o estado
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

// Item autônomo
type Item struct {
	Simbolo    rune
	Visivel    bool
	PosX, PosY int
}

// Inimigo autônomo com patrulha e vida
type Inimigo struct {
	Elemento
	PosX, PosY  int
	Visivel     bool
	PatrulhaDir int // +1 ou -1 para patrulha
	Vida        int
}

var (
	Personagem  = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Parede      = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao   = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio       = Elemento{' ', CorPadrao, CorPadrao, false}
	PatrulhaDir = Elemento{'☠', CorVermelho, CorPadrao, true}
)

// Construtor de Elemento
func novoElemento(sim rune, cor, corFundo Cor, tang bool) Elemento {
	return Elemento{simbolo: sim, cor: cor, corFundo: corFundo, tangivel: tang}
}

// Adiciona inimigos base
func adicionarInimigos(jogo *Jogo) {
	jogo.Inimigos = append(jogo.Inimigos,
		Inimigo{Elemento: Parede, PosX: 5, PosY: 5, Visivel: true, PatrulhaDir: 1, Vida: 3},
		Inimigo{Elemento: Vegetacao, PosX: 10, PosY: 7, Visivel: true, PatrulhaDir: 1, Vida: 3},
	)
}

// Adiciona dois seguranças patrulhando áreas específicas
func adicionarSegurancas(jogo *Jogo) {
	// Guarda 1: x de 2 a 8 na linha y=3
	g1 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 15, PosY: 3, Visivel: true, PatrulhaDir: 1, Vida: 5}
	// Guarda 2: x de 15 a 22 na linha y=6
	g2 := Inimigo{Elemento: novoElemento('☠', CorVermelho, CorPadrao, true), PosX: 25, PosY: 6, Visivel: true, PatrulhaDir: 1, Vida: 5}
	jogo.Inimigos = append(jogo.Inimigos, g1, g2)
}

// Patrulha e reação via canais
func (i *Inimigo) loopConcorrente(jogo *Jogo, chPat, chRea chan string) {
	for {
		select {
		case <-chPat:
			jogo.Mutex.Lock()
			nx := i.PosX + i.PatrulhaDir
			// inverte se bate em obstáculo
			if !jogoPodeMoverPara(jogo, nx, i.PosY) {
				i.PatrulhaDir *= +1
				nx = i.PosX + i.PatrulhaDir
			}
			if jogoPodeMoverPara(jogo, nx, i.PosY) {
				i.PosX = nx
			}
			jogo.Mutex.Unlock()

		case <-chRea:
			jogo.Mutex.Lock()
			if jogo.PosX == i.PosX && jogo.PosY == i.PosY {
				// desconta 2 pontos de vida em vez de atribuir valor fixo
				jogo.VidaPersonagem -= 2
				jogo.StatusMsg = fmt.Sprintf("Você foi atingido! Vida restante: %d", jogo.VidaPersonagem)
			}
			jogo.Mutex.Unlock()
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

	s := bufio.NewScanner(f)
	y := 0
	for s.Scan() {
		linha := s.Text()
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
	return s.Err()
}

func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	if y < 0 || y >= len(jogo.Mapa) || x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}
	return !jogo.Mapa[y][x].tangivel
}
