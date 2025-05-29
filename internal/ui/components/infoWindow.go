package components

import (
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/AlexPodd/metasploit_tester/internal/domain"
)

type InfoWindow struct {
	window *app.Window
	theme  *material.Theme
	exp    domain.Exploit
}

func NewInfoWindow(exp domain.Exploit) *InfoWindow {
	return &InfoWindow{
		window: new(app.Window),
		theme:  material.NewTheme(),
		exp:    exp,
	}
}

func (iw *InfoWindow) Run() error {
	var ops op.Ops
	for {
		switch e := iw.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			iw.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (iw *InfoWindow) layout(gtx layout.Context) layout.Dimensions {
	return layout.Inset{
		Top:    unit.Dp(20),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
		Bottom: unit.Dp(20),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.H6(iw.theme, iw.exp.Name).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Spacer{Height: unit.Dp(10)}.Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Описание:\n"+iw.exp.Description).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Платформа: "+iw.exp.Platform).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Авторы: "+strings.Join(iw.exp.Authors, ", ")).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Цели: "+strings.Join(iw.exp.Targets, ", ")).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Теги: "+strings.Join(iw.exp.Tags, ", ")).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(iw.theme, "Дата раскрытия: "+iw.exp.Disclosure).Layout(gtx)
			}),
		)
	})
}
