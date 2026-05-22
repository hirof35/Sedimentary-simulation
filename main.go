package main

import (
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 320 // 画面の幅（ピクセル）
	screenHeight = 240 // 画面の高さ（ピクセル）
)

// 各ピクセルの状態（ID）
const (
	Air  = 0
	Rock = 1
	Mud  = 2
	Sand = 3
)

// シミュレーションの状態を管理する構造体
type Game struct {
	grid  [screenHeight][screenWidth]uint8
	count int
}

// 物質ごとのカラーマップ
var colors = map[uint8]color.RGBA{
	Air:  {200, 230, 255, 255}, // 水色の空
	Rock: {80, 80, 80, 255},     // 濃いグレーの基盤岩
	Mud:  {139, 69, 19, 255},    // 茶色の泥層
	Sand: {244, 164, 96, 255},   // 薄茶・砂色の砂層
}

func NewGame() *Game {
	g := &Game{}
	rand.Seed(time.Now().UnixNano())

	// 下部の20ピクセルを基盤岩（地盤）にする
	for y := screenHeight - 20; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			g.grid[y][x] = Rock
		}
	}
	return g
}

// Update: 1秒間に60回実行されるロジック処理
func (g *Game) Update() error {
	g.count++

	// 10フレームごとに物質のタイプを切り替える（綺麗な縞模様を作るため）
	var currentMaterial uint8 = Sand
	if (g.count/120)%2 == 0 {
		currentMaterial = Mud
	}

	// 毎フレーム、ランダムな横位置（X座標）に土砂を3個降らせる
	for i := 0; i < 3; i++ {
		dropX := rand.Intn(screenWidth)
		g.deposit(dropX, currentMaterial)
	}

	// 侵食・崩落処理（毎フレーム全体を1回チェック）
	g.erode()

	return nil
}

// deposit: 粒子を上から落とし、何かにぶつかったらその上に固定する
func (g *Game) deposit(x int, material uint8) {
	for y := 0; y < screenHeight; y++ {
		// 空気以外の物質にぶつかったら、その1ピクセル上に配置
		if g.grid[y][x] != Air {
			if y > 0 {
				g.grid[y-1][x] = material
			}
			return
		}
	}
}

// erode: 砂山が急になりすぎたら、重力で左右の低い方へサラサラと崩れる挙動（セル・オートマトン）
func (g *Game) erode() {
	// 下から上に向かってスキャン（崩落を自然に連鎖させるため）
	for y := screenHeight - 2; y > 0; y-- {
		for x := 0; x < screenWidth; x++ {
			// 現在のピクセルが「砂」か「泥」で、真下が「空気」ならそのまま落下
			if g.grid[y][x] != Air && g.grid[y][x] != Rock {
				if g.grid[y+1][x] == Air {
					g.grid[y+1][x] = g.grid[y][x]
					g.grid[y][x] = Air
					continue
				}

				// 真下が詰まっている場合、斜め下（左右）が空いているかチェック
				leftEmpty := x > 0 && g.grid[y+1][x-1] == Air
				rightEmpty := x < screenWidth-1 && g.grid[y+1][x+1] == Air

				if leftEmpty && rightEmpty {
					// 両方空いていたらランダムにどちらかへ崩れる
					targetX := x - 1
					if rand.Float32() < 0.5 {
						targetX = x + 1
					}
					g.grid[y+1][targetX] = g.grid[y][x]
					g.grid[y][x] = Air
				} else if leftEmpty {
					g.grid[y+1][x-1] = g.grid[y][x]
					g.grid[y][x] = Air
				} else if rightEmpty {
					g.grid[y+1][x+1] = g.grid[y][x]
					g.grid[y][x] = Air
				}
			}
		}
	}
}

// Draw: 画面の描画処理（1秒間に60回呼び出される）
func (g *Game) Draw(screen *ebiten.Image) {
	// グリッドの状態をそのまま画面のピクセルとして1つずつ塗る
	for y := 0; y < screenHeight; y++ {
		for x := 0; x < screenWidth; x++ {
			material := g.grid[y][x]
			screen.Set(x, y, colors[material])
		}
	}
}

// Layout: ゲーム画面の論理サイズを決定
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2) // 画面を少し大きく見せるために2倍サイズでウィンドウを表示
	ebiten.SetWindowTitle("地層堆積シミュレーション (Ebitengine)")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}