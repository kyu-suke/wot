package gms

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
	"golang.org/x/crypto/ssh/terminal"
)

type field struct {
	top    int
	right  int
	bottom int
	left   int
}

type point struct {
	X int
	Y int
}

type state struct {
	End  bool
	Head point
}

//snake body
type body struct {
	x int
	y int
}

//feed
type Feed struct {
	x int
	y int
}

var _timeSpan int
var _height int
var _width int

var mu sync.Mutex

var head = "<"
var toX = 1
var toY = 0

var snake = []body{}
var vector = "right"

var fld = field{}
var feed = []Feed{}
var flg = 0
var out = ""

func timerLoop(tch chan bool) {
	for {
		tch <- true
		time.Sleep(time.Duration(_timeSpan) * time.Millisecond)
	}
}

func keyEventLoop(kch chan termbox.Event) {
	for {
		kch <- termbox.PollEvent()
	}
}

func drawLoop(sch chan state) {
	for {
		st := <-sch
		mu.Lock()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		drawLine(1, 0, "EXIT : ESC KEY")

		// field
		for i := fld.top; i < fld.bottom; i++ {
			drawLine(fld.right, i, "|")
			drawLine(fld.left, i, "|")
		}
		for i := 0; i < fld.right; i++ {
			drawLine(i, fld.top, "-")
			drawLine(i, fld.bottom, "-")
		}

		// esa
		if len(feed) < 1 {
			rand.Seed(time.Now().UnixNano())
			// XXX maybe fixed
			feed = []Feed{
				{
					x: rand.Intn(fld.right-1) + 1,
					y: rand.Intn(fld.bottom-4) + 3,
				},
			}
		}
		drawLine(feed[0].x, feed[0].y, "+")

		if st.End == false {
			pre_x := 0
			pre_y := 0
			for k, v := range snake {
				if k == 0 {
					drawLine(st.Head.X, st.Head.Y, head)
					snake[k].x = st.Head.X
					snake[k].y = st.Head.Y
				} else {
					drawLine(pre_x, pre_y, "*")
					snake[k].x = pre_x
					snake[k].y = pre_y
				}
				pre_x = v.x
				pre_y = v.y
			}
		} else {
			drawLine(0, 0, "PUSH ANY KEY")
		}
		termbox.Flush()
		mu.Unlock()
	}
}

func drawLine(x, y int, str string) {
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(x+i, y, runes[i], termbox.ColorDefault, termbox.ColorDefault)
	}
}

func controller(stateCh chan state, keyCh chan termbox.Event, timerCh chan bool) {
	st := initGame()
	for {
		if flg == 1 {
			st.End = true
			return
		}
		select {
		case key := <-keyCh:
			mu.Lock()
			switch {
			case key.Key == termbox.KeyEsc || key.Key == termbox.KeyCtrlC: //exit
				st.End = true
				mu.Unlock()
				return
			case key.Key == termbox.KeyArrowLeft || key.Ch == 'h': //left
				if vector != "right" {
					vector = "left"
					toX = -1
					toY = 0
					head = ">"
				}
				break
			case key.Key == termbox.KeyArrowRight || key.Ch == 'l': //right
				if vector != "left" {
					vector = "right"
					toX = 1
					toY = 0
					head = "<"
				}
				break
			case key.Key == termbox.KeyArrowUp || key.Ch == 'k': //up
				if vector != "down" {
					vector = "up"
					toX = 0
					toY = -1
					head = "V"
				}
				break
			case key.Key == termbox.KeyArrowDown || key.Ch == 'j': //down
				if vector != "up" {
					vector = "down"
					toX = 0
					toY = 1
					head = "A"
				}
				break
			default:
				st.End = false
				break
			}
			mu.Unlock()
			stateCh <- st
			break
		case <-timerCh:
			mu.Lock()
			if st.End == false {

				st.Head.X += toX
				st.Head.Y += toY
				st = checkCollision(st)
			}
			mu.Unlock()
			stateCh <- st
			break
		default:
			break
		}
	}
}

func initGame() state {
	st := state{End: true}
	st.Head.X, st.Head.Y = _width/2, _height*2/3
	snake = []body{}
	snake = append(snake, body{x: st.Head.X, y: st.Head.Y})
	snake = append(snake, body{x: st.Head.X + 1, y: st.Head.Y + 1})

	//field
	fld.top = 2
	fld.right = _width
	fld.bottom = _height
	fld.left = 0

	return st
}

//intersect
func checkCollision(st state) state {
	//LR wall
	if st.Head.X <= fld.left || st.Head.X >= fld.right {
		st = initGame()
	}

	//UD wall
	if st.Head.Y <= fld.top || st.Head.Y >= fld.bottom {
		st = initGame()
	}

	//feed
	for i := range feed {
		if feed[i].y == st.Head.Y {
			if feed[i].x <= st.Head.X && feed[i].x >= st.Head.X {
				snake = append(snake, body{x: st.Head.X + 1, y: st.Head.Y + 1})
				eatFeed()
				break
			}
		}
	}

	//body
	for i := range snake {
		if snake[i].y == st.Head.Y {
			if snake[i].x <= st.Head.X && snake[i].x >= st.Head.X {
				st = initGame()
				break
			}
		}
	}

	return st
}

func eatFeed() {
	feed = []Feed{}
}

func gameclose() {
	termbox.Close()
	fmt.Println(out)
}

func StartSnakeGame(speed, height, width int) {

	defer gameclose()

	_timeSpan = speed
	_height = height
	_width = width

	stateCh := make(chan state)
	keyCh := make(chan termbox.Event)
	timerCh := make(chan bool)

	go drawLoop(stateCh)
	go keyEventLoop(keyCh)
	go timerLoop(timerCh)

	if !terminal.IsTerminal(0) {
		go func() {
			stdin, _ := ioutil.ReadAll(os.Stdin)
			out = string(stdin)

			time.Sleep(1000 * time.Millisecond)
			flg = 1
		}()
	}

	// FIXME ignore output by other process excepting stdout e.g. git clone
	termbox.SetCursor(0, 0)

	controller(stateCh, keyCh, timerCh)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}
