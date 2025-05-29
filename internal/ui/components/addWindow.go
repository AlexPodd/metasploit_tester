package components

import (
	"image/color"
	"io"
	"log"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"github.com/AlexPodd/metasploit_tester/internal/domain"
)

type FileEditor struct {
	SubdirEditor   widget.Editor
	FilenameEditor widget.Editor
	SelectFileBtn  widget.Clickable
	SelectedFile   string
	FileContent    string
}

type AddFileWindow struct {
	window   *app.Window
	editor   FileEditor
	saveBtn  widget.Clickable
	closeBtn widget.Clickable
	theme    *material.Theme
	onSave   func(domain.ExploitFile)
	onClose  func()
	explorer *explorer.Explorer
}

func NewAddFileWindow(theme *material.Theme, onSave func(domain.ExploitFile), onClose func()) *AddFileWindow {
	win := new(app.Window)
	app.Title("Добавить эксплойт")
	app.Size(unit.Dp(500), unit.Dp(400))

	editor := FileEditor{
		SubdirEditor:   widget.Editor{SingleLine: true},
		FilenameEditor: widget.Editor{SingleLine: true},
		SelectFileBtn:  widget.Clickable{},
	}

	// Инициализация explorer
	exp := explorer.NewExplorer(win)

	return &AddFileWindow{
		window:   win,
		editor:   editor,
		theme:    theme,
		onSave:   onSave,
		onClose:  onClose,
		explorer: exp,
	}
}

func (w *AddFileWindow) Run() error {
	var ops op.Ops
	for {
		switch e := w.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			w.layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (w *AddFileWindow) layout(gtx layout.Context) layout.Dimensions {
	// Обработка выбора файла
	if w.editor.SelectFileBtn.Clicked(gtx) {
		go func() {
			// Открываем диалог выбора файла и получаем io.ReadCloser
			fileReader, err := w.explorer.ChooseFile()
			if err != nil {
				return
			}
			defer fileReader.Close()

			// Читаем содержимое файла напрямую из reader'а
			content, err := io.ReadAll(fileReader)
			if err != nil {
				return
			}

			w.editor.FileContent = string(content)
			w.editor.SelectedFile = "Выбранный файл"
			w.window.Invalidate()
		}()
	}

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		border := widget.Border{
			Color:        color.NRGBA{R: 240, G: 240, B: 240, A: 255},
			CornerRadius: unit.Dp(8),
			Width:        unit.Dp(1),
		}
		return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(16), Bottom: unit.Dp(16), Left: unit.Dp(24), Right: unit.Dp(24)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.H6(w.theme, "Добавление эксплойта").Layout),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(8)}.Layout(gtx,
							material.Editor(w.theme, &w.editor.SubdirEditor, "Подкаталог").Layout)
					}),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(8)}.Layout(gtx,
							material.Editor(w.theme, &w.editor.FilenameEditor, "Имя файла").Layout)
					}),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return material.Button(w.theme, &w.editor.SelectFileBtn, "📁 Выбрать файл").Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if w.editor.SelectedFile != "" {
										return material.Body1(w.theme, w.editor.SelectedFile).Layout(gtx)
									}
									return material.Body1(w.theme, "Файл не выбран").Layout(gtx)
								}),
							)
						})
					}),

					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if w.saveBtn.Clicked(gtx) {
										exploit := domain.ExploitFile{
											Subdir:  w.editor.SubdirEditor.Text(),
											Name:    w.editor.FilenameEditor.Text(),
											Content: w.editor.FileContent,
										}
										log.Print(exploit)
										w.onSave(exploit)
										w.window.Perform(system.ActionClose)
									}
									return material.Button(w.theme, &w.saveBtn, "💾 Сохранить").Layout(gtx)
								}),
								layout.Rigid(layout.Spacer{Width: unit.Dp(16)}.Layout),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									if w.closeBtn.Clicked(gtx) {
										w.onClose()
										w.window.Perform(system.ActionClose)
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
