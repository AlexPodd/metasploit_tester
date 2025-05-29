package ui

import (
	"image"
	"image/color"
	"log"
	"os"
	"sort"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/AlexPodd/metasploit_tester/internal/domain"
	"github.com/AlexPodd/metasploit_tester/internal/metasploit"
	"github.com/AlexPodd/metasploit_tester/internal/reportGenerate"
	"github.com/AlexPodd/metasploit_tester/internal/ui/components"
)

type MainWindow struct {
	window     *app.Window
	theme      *material.Theme
	exploits   []domain.Exploit
	selected   map[string]bool
	filterTags *map[string]struct{}
	tagButtons map[string]*widget.Clickable
	checkboxes map[string]*widget.Bool
	startBtn   widget.Clickable
	client     *metasploit.Client

	isRunning bool
	progress  float32
	total     float32

	infoButtons map[string]*widget.Clickable

	infoWindow     *components.InfoWindow
	infoWindowOpen bool
}

func NewMainWindow(exploits []domain.Exploit, tags *map[string]struct{}, client *metasploit.Client) *MainWindow {
	mw := &MainWindow{
		window:      new(app.Window),
		theme:       material.NewTheme(),
		exploits:    exploits,
		selected:    make(map[string]bool),
		filterTags:  tags,
		tagButtons:  make(map[string]*widget.Clickable),
		infoButtons: make(map[string]*widget.Clickable),
		checkboxes:  make(map[string]*widget.Bool),
		client:      client,
	}

	// Init checkboxes and tag buttons
	for _, exp := range exploits {
		mw.checkboxes[exp.Name] = new(widget.Bool)
		mw.infoButtons[exp.Name] = new(widget.Clickable)
		for _, tag := range exp.Tags {
			if _, exists := mw.tagButtons[tag]; !exists {
				mw.tagButtons[tag] = new(widget.Clickable)
			}
		}
	}
	return mw
}

func (mw *MainWindow) Start() error {
	go func() {
		if err := mw.run(); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
	return nil
}

func (mw *MainWindow) run() error {
	var ops op.Ops
	for {

		switch e := mw.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			mw.layoutUI(gtx)
			e.Frame(gtx.Ops)
		}
	}
}

func (mw *MainWindow) layoutUI(gtx layout.Context) layout.Dimensions {
	dims := mw.layoutMainContent(gtx)

	if mw.infoWindowOpen {
		macroOp := op.Record(gtx.Ops)

		clip.Rect(image.Rectangle{Max: dims.Size}).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 100})

		call := macroOp.Stop()
		op.Defer(gtx.Ops, call)
	}

	return dims
}

func (mw *MainWindow) layoutMainContent(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mw.infoWindowOpen {
				return layout.Dimensions{}
			}
			return mw.layoutTagFilters(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mw.infoWindowOpen {
				return layout.Dimensions{}
			}
			return mw.layoutExploits(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if mw.infoWindowOpen {
				return layout.Dimensions{}
			}
			return mw.layoutStartButton(gtx)
		}),
	)
}

func (mw *MainWindow) layoutTagFilters(gtx layout.Context) layout.Dimensions {
	tags := make([]string, 0, len(mw.tagButtons))
	for tag := range mw.tagButtons {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return layout.Flex{
		Axis: layout.Horizontal,
	}.Layout(gtx, func() []layout.FlexChild {
		children := make([]layout.FlexChild, 0, len(tags))
		for _, tag := range tags {
			btn := mw.tagButtons[tag]
			if btn.Clicked(gtx) {
				if _, exists := (*mw.filterTags)[tag]; exists {
					delete(*mw.filterTags, tag)
				} else {
					(*mw.filterTags)[tag] = struct{}{}
				}
			}

			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Выбран ли тег?
				selected := false
				if _, exists := (*mw.filterTags)[tag]; exists {
					selected = true
				}

				// Стиль кнопки
				style := material.Button(mw.theme, btn, tag)
				if selected {
					style.Background = color.NRGBA{R: 100, G: 149, B: 237, A: 255} // Cornflower Blue
					style.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}      // Белый текст
				} else {
					style.Background = color.NRGBA{R: 230, G: 230, B: 230, A: 255} // Светло-серый фон
					style.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}            // Чёрный текст
				}

				return style.Layout(gtx)
			}))
		}
		return children
	}()...)
}

func (mw *MainWindow) layoutExploits(gtx layout.Context) layout.Dimensions {
	filteredExploits := make([]domain.Exploit, 0, len(mw.exploits))
	filtering := len(*mw.filterTags) > 0
	for _, exp := range mw.exploits {
		if !filtering {
			filteredExploits = append(filteredExploits, exp)
			continue
		}
		for _, tag := range exp.Tags {
			if _, ok := (*mw.filterTags)[tag]; ok {
				filteredExploits = append(filteredExploits, exp)
				break
			}
		}
	}

	if len(filteredExploits) == 0 {
		lbl := material.Body1(mw.theme, "Нет подходящих эксплойтов")
		return lbl.Layout(gtx)
	}

	for _, exp := range filteredExploits {
		btn := mw.infoButtons[exp.Name]
		if btn != nil && btn.Clicked(gtx) && !mw.infoWindowOpen {
			mw.openInfoWindow(exp)
			break
		}
	}

	list := &layout.List{Axis: layout.Vertical}
	return list.Layout(gtx, len(filteredExploits), func(gtx layout.Context, i int) layout.Dimensions {
		exp := filteredExploits[i]
		checkbox := mw.checkboxes[exp.Name]
		infoBtn := mw.infoButtons[exp.Name]

		return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.CheckBox(mw.theme, checkbox, exp.Name).Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.Button(mw.theme, infoBtn, "ℹ").Layout(gtx)
				}),
			)
		})
	})
}

func (mw *MainWindow) openInfoWindow(exp domain.Exploit) {
	mw.infoWindow = components.NewInfoWindow(exp)
	mw.infoWindowOpen = true

	go func() {
		if err := mw.infoWindow.Run(); err != nil {
			log.Println("InfoWindow error:", err)
		}
		mw.infoWindow = nil
		mw.infoWindowOpen = false
		mw.window.Invalidate()
	}()
}

func (mw *MainWindow) layoutStartButton(gtx layout.Context) layout.Dimensions {
	if mw.isRunning {
		return mw.drawProgressBar(gtx, mw.progress, mw.total)
	}

	if mw.startBtn.Clicked(gtx) {
		log.Println("Running selected exploits:")

		var run []domain.Exploit
		for name, box := range mw.checkboxes {
			if box.Value {
				for _, element := range mw.exploits {
					if element.Name == name {
						run = append(run, element)
					}
				}
			}
		}

		mw.total = float32(len(run))
		mw.progress = 0
		mw.isRunning = true

		progressChan := make(chan float32)

		go func() {
			for p := range progressChan {
				mw.progress = p
				mw.window.Invalidate()
			}
			mw.isRunning = false
			mw.window.Invalidate()
		}()

		go func() {
			report, err := mw.client.Execute(run, progressChan)

			if err != nil {
				log.Println("Execution error:", err)
			} else {
				reportGenerate.GenerateReport(report)
			}
		}()
	}

	return material.Button(mw.theme, &mw.startBtn, "Запустить").Layout(gtx)
}

func (mw *MainWindow) drawProgressBar(gtx layout.Context, current, total float32) layout.Dimensions {
	if total == 0 {
		total = 1 // избегаем деления на 0
	}
	progress := float32(current) / float32(total)
	barHeight := gtx.Dp(6)
	barWidth := gtx.Constraints.Max.X

	// Обёртка для всей полоски
	return layout.Stack{}.Layout(gtx,
		// Фон (серый)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			dims := layout.Dimensions{
				Size: image.Pt(barWidth, barHeight),
			}
			paint.FillShape(gtx.Ops, color.NRGBA{R: 200, G: 200, B: 200, A: 255},
				clip.Rect{Max: dims.Size}.Op())
			return dims
		}),
		// Прогресс (синий)
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			progressWidth := int(float32(barWidth) * progress)
			dims := layout.Dimensions{
				Size: image.Pt(progressWidth, barHeight),
			}
			paint.FillShape(gtx.Ops, color.NRGBA{R: 33, G: 150, B: 243, A: 255},
				clip.Rect{Max: dims.Size}.Op())
			return dims
		}),
	)
}

func join(arr []string) string {
	if len(arr) == 0 {
		return "-"
	}
	return "• " + arr[0] + func() string {
		if len(arr) == 1 {
			return ""
		}
		s := ""
		for _, str := range arr[1:] {
			s += ", " + str
		}
		return s
	}()
}
