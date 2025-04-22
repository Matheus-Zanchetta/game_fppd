package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

// Cor encapsula as cores do termbox
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

func interfaceDesenharElemento(x, y int, elem Elemento) {
	termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
}

func interfaceDesenharJogo(jogo *Jogo) {
	interfaceLimparTela()
	// Desenha mapa
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			interfaceDesenharElemento(x, y, elem)
		}
	}
	// Desenha itens
	for _, item := range jogo.Items {
		if item.Visivel {
			interfaceDesenharElemento(item.PosX, item.PosY,
				novoElemento(item.Simbolo, CorVerde, CorPadrao, false))
		}
	}
	// Desenha inimigos
	for _, inimigo := range jogo.Inimigos {
		if inimigo.Visivel {
			interfaceDesenharElemento(inimigo.PosX, inimigo.PosY, inimigo.Elemento)
		}
	}
	// Desenha personagem
	interfaceDesenharElemento(jogo.PosX, jogo.PosY, Personagem)
	// Barra de status
	status := fmt.Sprintf("Vida: %d    %s", jogo.VidaPersonagem, jogo.StatusMsg)
	for i, c := range status {
		termbox.SetCell(i, len(jogo.Mapa)+1, c, CorTexto, CorPadrao)
	}
	interfaceAtualizarTela()
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
