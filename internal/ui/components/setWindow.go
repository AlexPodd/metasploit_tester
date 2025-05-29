package components

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/AlexPodd/metasploit_tester/internal/domain"
)

type ParamEditor struct {
	KeyEditor   widget.Editor
	ValueEditor widget.Editor
	DeleteBtn   widget.Clickable
}

type SetWindow struct {
	exploit  domain.Exploit
	editors  []*ParamEditor
	saveBtn  widget.Clickable
	closeBtn widget.Clickable
	addBtn   widget.Clickable
	theme    *material.Theme
	onSave   func(updated []domain.ExploitParam)
	onClose  func()
}

func NewSetWindow(theme *material.Theme, exploit domain.Exploit, onSave func([]domain.ExploitParam), onClose func()) *SetWindow {
	editors := make([]*ParamEditor, len(exploit.Params))
	for i, param := range exploit.Params {
		keyEditor := widget.Editor{SingleLine: true, Submit: true}
		valueEditor := widget.Editor{SingleLine: true, Submit: true}
		keyEditor.SetText(param.Key)
		valueEditor.SetText(param.Value)
		editors[i] = &ParamEditor{
			KeyEditor:   keyEditor,
			ValueEditor: valueEditor,
			DeleteBtn:   widget.Clickable{},
		}
	}

	return &SetWindow{
		exploit: exploit,
		editors: editors,
		theme:   theme,
		onSave:  onSave,
		onClose: onClose,
	}
}
func (w *SetWindow) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Обёртка, имитирующая material.Card
		border := widget.Border{
			Color:        color.NRGBA{R: 240, G: 240, B: 240, A: 255},
			CornerRadius: unit.Dp(8),
			Width:        unit.Dp(1),
		}

		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(16), Left: unit.Dp(24), Right: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					// Заголовок
					layout.Rigid(material.H6(w.theme, "Параметры эксплойта: "+w.exploit.Name).Layout),

					// Список параметров
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx, func() []layout.FlexChild {
							children := make([]layout.FlexChild, 0, len(w.editors))
							for i, editor := range w.editors {
								idx := i
								children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Right: unit.Dp(4)}.Layout(gtx,
												material.Editor(w.theme, &editor.KeyEditor, "KEY").Layout)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return layout.Inset{Right: unit.Dp(4)}.Layout(gtx,
												material.Editor(w.theme, &editor.ValueEditor, "VALUE").Layout)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											if editor.DeleteBtn.Clicked(gtx) {
												w.editors = append(w.editors[:idx], w.editors[idx+1:]...)
											}
											return material.Button(w.theme, &editor.DeleteBtn, "✕").Layout(gtx)
										}),
									)
								}))
							}
							return children
						}()...)
					}),

					// Кнопка добавить
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if w.addBtn.Clicked(gtx) {
							w.editors = append(w.editors, &ParamEditor{
								KeyEditor:   widget.Editor{SingleLine: true},
								ValueEditor: widget.Editor{SingleLine: true},
								DeleteBtn:   widget.Clickable{},
							})
						}
						return layout.Inset{Top: unit.Dp(12)}.Layout(gtx,
							material.Button(w.theme, &w.addBtn, "+ Добавить параметр").Layout)
					}),

					// Кнопки сохранить и закрыть
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if w.saveBtn.Clicked(gtx) {
										params := make([]domain.ExploitParam, 0, len(w.editors))
										for _, editor := range w.editors {
											k := editor.KeyEditor.Text()
											v := editor.ValueEditor.Text()
											if k != "" {
												params = append(params, domain.ExploitParam{Key: k, Value: v})
											}
										}
										w.onSave(params)
									}
									return material.Button(w.theme, &w.saveBtn, "💾 Сохранить").Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if w.closeBtn.Clicked(gtx) {
										w.onClose()
									}
									return material.Button(w.theme, &w.closeBtn, "Закрыть").Layout(gtx)
								}),
							)
						})
					}),
				)
			})
		})
	})
}
