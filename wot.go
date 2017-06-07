package main

import (
	"flag"
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

//ステータス
type state struct {
	End       bool
	Ball      point
	Vec       point
	Life      int
	Score     int
	HighScore int
}

// snake body
type body struct {
	x int
	y int
}

//えさ
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

//タイマーイベント
func timerLoop(tch chan bool) {
	for {
		tch <- true
		time.Sleep(time.Duration(_timeSpan) * time.Millisecond)
	}
}

//キーイベント
func keyEventLoop(kch chan termbox.Event) {
	for {
		kch <- termbox.PollEvent()
	}
}

//画面描画
func drawLoop(sch chan state) {
	for {
		st := <-sch
		mu.Lock()
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		drawLine(1, 0, "EXIT : ESC KEY")
		drawLine(_width-10, 0, fmt.Sprintf("Life : %02d", st.Life))

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
			// TODO 縦範囲外に出る
			feed = []Feed{
				{
					x: rand.Intn(fld.right - 1),
					y: rand.Intn(fld.bottom - 1),
				},
			}
		}
		drawLine(feed[0].x, feed[0].y, "+")

		if st.End == false {
			pre_x := 0
			pre_y := 0
			for k, v := range snake {
				if k == 0 {
					drawLine(st.Ball.X, st.Ball.Y, head)
					snake[k].x = st.Ball.X
					snake[k].y = st.Ball.Y
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

//行を描画
func drawLine(x, y int, str string) {
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(x+i, y, runes[i], termbox.ColorDefault, termbox.ColorDefault)
	}
}

//ゲームメイン処理
func controller(stateCh chan state, keyCh chan termbox.Event, timerCh chan bool) {
	st := initGame()
	for {
		if flg == 1 {
			st.End = true
			return
		}
		select {
		case key := <-keyCh: //キーイベント
			mu.Lock()
			switch {
			case key.Key == termbox.KeyEsc || key.Key == termbox.KeyCtrlC: //ゲーム終了
				st.End = true
				mu.Unlock()
				return
			case key.Key == termbox.KeyArrowLeft || key.Ch == 'h': //ひだり
				if vector != "right" {
					vector = "left"
					toX = -1
					toY = 0
					head = ">"
				}
				break
			case key.Key == termbox.KeyArrowRight || key.Ch == 'l': //みぎ
				if vector != "left" {
					vector = "right"
					toX = 1
					toY = 0
					head = "<"
				}
				break
			case key.Key == termbox.KeyArrowUp || key.Ch == 'k': //うえ
				if vector != "down" {
					vector = "up"
					toX = 0
					toY = -1
					head = "V"
				}
				break
			case key.Key == termbox.KeyArrowDown || key.Ch == 'j': //した
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
		case <-timerCh: //タイマーイベント
			mu.Lock()
			if st.End == false {

				st.Ball.X += toX
				st.Ball.Y += toY
				//st.Ball.X += st.Vec.X
				//st.Ball.Y += st.Vec.Y
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
	st.Ball.X, st.Ball.Y = _width/2, _height*2/3
	st.Vec.X, st.Vec.Y = 1, -1
	st.Life = 3
	snake = []body{}
	snake = append(snake, body{x: st.Ball.X, y: st.Ball.Y})
	snake = append(snake, body{x: st.Ball.X + 1, y: st.Ball.Y + 1})

	// field
	fld.top = 2
	fld.right = _width
	fld.bottom = _height
	fld.left = 0

	return st
}

//衝突判定
func checkCollision(st state) state {
	//左右の壁
	if st.Ball.X <= fld.left || st.Ball.X >= fld.right {
		hs := 0
		if st.HighScore < st.Score {
			hs = st.Score
		}
		st = initGame()
		st.HighScore = hs
	}

	//上下の壁
	if st.Ball.Y <= fld.top || st.Ball.Y >= fld.bottom {
		hs := 0
		if st.HighScore < st.Score {
			hs = st.Score
		}
		st = initGame()
		st.HighScore = hs
	}

	//えさとの衝突判定
	for i := range feed {
		if feed[i].y == st.Ball.Y {
			if feed[i].x <= st.Ball.X && feed[i].x >= st.Ball.X {
				st.Vec.Y *= -1
				st.Score++
				snake = append(snake, body{x: st.Ball.X + 1, y: st.Ball.Y + 1})
				eatFeed()
				break
			}
		}
	}

	//体衝突判定
	for i := range snake {
		if snake[i].y == st.Ball.Y {
			if snake[i].x <= st.Ball.X && snake[i].x >= st.Ball.X {
				hs := 0
				if st.HighScore < st.Score {
					hs = st.Score
				}
				st = initGame()
				st.HighScore = hs
				break
			}
		}
	}

	return st
}

func eatFeed() {
	feed = []Feed{}
}

//main
func main() {

	var s = flag.Int("s", 100, "speed")
	var h = flag.Int("h", 25, "stage height")
	var w = flag.Int("w", 80, "stage width")
	var c = flag.String("c", "snake", "select cassette. [snake, ...]")
	flag.Parse()
	fmt.Println(*s)
	fmt.Println(*h)
	fmt.Println(*w)
	fmt.Println(*c)
	_timeSpan = *s
	_height = *h
	_width = *w

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	stateCh := make(chan state)
	keyCh := make(chan termbox.Event)
	timerCh := make(chan bool)

	go drawLoop(stateCh)
	go keyEventLoop(keyCh)
	go timerLoop(timerCh)

	if !terminal.IsTerminal(0) {
		go func() {
			out, _ := ioutil.ReadAll(os.Stdin)
			fmt.Println(string(out))

			time.Sleep(1000 * time.Millisecond)
			flg = 1
		}()
	}
	termbox.SetCursor(0, 0)
	controller(stateCh, keyCh, timerCh)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	defer termbox.Close()
}
