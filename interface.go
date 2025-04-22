// interface.go
package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

type Cor = termbox.Attribute

const (
	CorPadrao      Cor = termbox.ColorDefault
	CorCinzaEscuro     = termbox.ColorDarkGray
	CorVermelho        = termbox.ColorRed
	CorVerde           = termbox.ColorGreen
	CorParede          = termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim
	CorFundoParede     = termbox.ColorDarkGray
	CorTexto           = termbox.ColorDarkGray
)

func interfaceIniciar() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
}

func interfaceFinalizar() {
	termbox.Close()
}

func interfaceLimparTela() {
	termbox.Clear(CorPadrao, CorPadrao)
}

func interfaceAtualizarTela() {
	termbox.Flush()
}

// Protegido por mutex para evitar race conditions durante o redraw
func interfaceDesenharJogo(jogo *Jogo) {
	jogo.Mutex.Lock()
	defer jogo.Mutex.Unlock()

	interfaceLimparTela()
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			interfaceDesenharElemento(x, y, elem)
		}
	}
	for _, item := range jogo.Items {
		if item.Visivel {
			interfaceDesenharElemento(item.PosX, item.PosY,
				novoElemento(item.Simbolo, CorVerde, CorPadrao, false))
		}
	}
	for _, inimigo := range jogo.Inimigos {
		if inimigo.Visivel {
			interfaceDesenharElemento(inimigo.PosX, inimigo.PosY, inimigo.Elemento)
		}
	}
	interfaceDesenharElemento(jogo.PosX, jogo.PosY, Personagem)

	status := fmt.Sprintf("Vida: %d    %s", jogo.VidaPersonagem, jogo.StatusMsg)
	for i, c := range status {
		termbox.SetCell(i, len(jogo.Mapa)+1, c, CorTexto, CorPadrao)
	}

	interfaceAtualizarTela()
}

func interfaceDesenharElemento(x, y int, elem Elemento) {
	termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
}

func interfaceLerEventoTeclado() EventoTeclado {
	ev := termbox.PollEvent()
	if ev.Type != termbox.EventKey {
		return EventoTeclado{}
	}
	if ev.Key == termbox.KeyEsc {
		return EventoTeclado{Tipo: "sair"}
	}
	if ev.Ch == 'e' {
		return EventoTeclado{Tipo: "interagir"}
	}
	return EventoTeclado{Tipo: "mover", Tecla: ev.Ch}
}
