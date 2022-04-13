package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	SCREEN_WIDTH  = int32(800)
	SCREEN_HEIGHT = int32(450)
	GRAVITY       = int32(3)
	TARGET_FPS    = 60
)

type RenderScreen int

const (
	MainMenu RenderScreen = iota
	GamePlay
	GameEnding
	GameOver
)

type GameEntity struct {
	xPos    int32
	yPos    int32
	width   int32
	height  int32
	texture rl.Texture2D
}

func (g *GameEntity) update(xPos int32, yPos int32, texture rl.Texture2D) {
	g.xPos = xPos
	g.yPos = yPos
	g.texture = texture
}

type GamePlayStartValues struct {
	score                float64
	gopher               GameEntity
	asteroids            []GameEntity
	renderScreen         RenderScreen
	gameEndingIterations int
}

func getRandomYPos() int32 {
	min := 50
	max := int(SCREEN_HEIGHT) - min
	return int32(rand.Intn(max-min) + min)
}

func setupNewGame(gopherTexture rl.Texture2D, asteroidTexture rl.Texture2D) GamePlayStartValues {
	rand.Seed(time.Now().UnixNano())
	gopher := GameEntity{
		xPos:    (SCREEN_WIDTH / 2) - (gopherTexture.Width / 2),
		yPos:    (SCREEN_HEIGHT / 2) - (gopherTexture.Height / 2),
		width:   gopherTexture.Width,
		height:  gopherTexture.Height,
		texture: gopherTexture,
	}

	asteroids := []GameEntity{}
	initialAsteroid := GameEntity{
		xPos:    SCREEN_WIDTH,
		yPos:    getRandomYPos(),
		width:   asteroidTexture.Width,
		height:  asteroidTexture.Height,
		texture: asteroidTexture,
	}
	asteroids = append(asteroids, initialAsteroid)

	return GamePlayStartValues{
		score:                0.0,
		gopher:               gopher,
		asteroids:            asteroids,
		renderScreen:         GamePlay,
		gameEndingIterations: 0,
	}
}

func main() {
	rl.InitWindow(SCREEN_WIDTH, SCREEN_HEIGHT, "Flappy Gopher")
	rl.SetTargetFPS(TARGET_FPS)

	backgroundImage := rl.LoadImage("assets/images/background.png")
	backgroundTexture := rl.LoadTextureFromImage(backgroundImage)

	explosionImage := rl.LoadImage("assets/images/explosion.png")
	explosionTexture := rl.LoadTextureFromImage(explosionImage)

	gopherUpImage := rl.LoadImage("assets/images/gopher-up.png")
	gopherUpTexture := rl.LoadTextureFromImage(gopherUpImage)

	gopherDownImage := rl.LoadImage("assets/images/gopher-down.png")
	gopherDownTexture := rl.LoadTextureFromImage(gopherDownImage)

	asteroidImage := rl.LoadImage("assets/images/asteroid.png")
	asteroidTexture := rl.LoadTextureFromImage(asteroidImage)

	rl.InitAudioDevice()
	explodeSound := rl.LoadSound("assets/sounds/explode.wav")
	gameplayMusic := rl.LoadMusicStream("assets/sounds/gameplay.mp3")

	highestScore := 0
	renderScreen := MainMenu

	gamePlayStartValues := setupNewGame(gopherDownTexture, asteroidTexture)
	score := gamePlayStartValues.score
	gopher := gamePlayStartValues.gopher
	asteroids := gamePlayStartValues.asteroids
	gameEndingIterations := gamePlayStartValues.gameEndingIterations

	//The main render loop. Targets `TARGET_FPS` iterations per second
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)
		rl.DrawTexture(backgroundTexture, 0, 0, rl.White)

		if renderScreen == MainMenu {
			txt := `
			Welcome to Flappy Gopher
			Avoid the asteroids as long as possible
			Hold the Spacebar to go up
			Press Esc to quit at any time
			Press Enter to start`
			txtMeasurements := rl.MeasureText(txt, 20)
			rl.DrawText(txt, (SCREEN_WIDTH/2)-(txtMeasurements/2), SCREEN_HEIGHT/2-(txtMeasurements/4), 20, rl.White)

			if rl.IsKeyPressed(rl.KeyEnter) {
				renderScreen = GamePlay
			}
		} else if highestScore > 0 {
			txt := fmt.Sprintf("Highest Score: %d", highestScore)
			txtMeasurements := rl.MeasureText(txt, 20)
			rl.DrawText(txt, SCREEN_WIDTH-txtMeasurements-10, 10, 20, rl.LightGray)
		}

		//Update gopher and asteroids for next frame
		if renderScreen == GamePlay {
			if !rl.IsMusicStreamPlaying(gameplayMusic) {
				rl.PlayMusicStream(gameplayMusic)
			} else {
				rl.UpdateMusicStream(gameplayMusic)
			}

			score += 0.25

			var newGopherY int32
			var newGopherTexture rl.Texture2D

			if rl.IsKeyDown(rl.KeySpace) {
				newGopherY = gopher.yPos - GRAVITY
				newGopherTexture = gopherUpTexture
			} else {
				// newGopherY = int32(math.Min(float64(SCREEN_HEIGHT-gopher.height), float64()))
				newGopherY = gopher.yPos + GRAVITY
				newGopherTexture = gopherDownTexture
			}

			if newGopherY < -gopher.height {
				//gopher has gone past top of window, flip to the bottom
				newGopherY = SCREEN_HEIGHT + gopher.height
			} else if newGopherY > SCREEN_HEIGHT+gopher.height {
				//opposite, gopher is below the bottom, flip to the top
				newGopherY = -gopher.height
			}

			gopher.update(gopher.xPos, newGopherY, newGopherTexture)

			if score > 1 && (math.Mod(score, 100.0) == 0) && len(asteroids) <= 15 {
				asteroids = append(asteroids, GameEntity{
					xPos:    SCREEN_WIDTH,
					yPos:    getRandomYPos(),
					width:   asteroidTexture.Width,
					height:  asteroidTexture.Height,
					texture: asteroidTexture,
				})
			}

			for idx, asteroid := range asteroids {
				newAsteroidY := asteroid.yPos
				newAsteroidX := asteroid.xPos - GRAVITY

				if newAsteroidX < -asteroid.width {
					//Off screen, move back to right hand side
					newAsteroidX = SCREEN_WIDTH + asteroid.width
					newAsteroidY = getRandomYPos()
				}

				asteroids[idx].update(newAsteroidX, newAsteroidY, asteroids[idx].texture)

				if rl.CheckCollisionRecs(
					rl.NewRectangle(float32(gopher.xPos), float32(gopher.yPos), float32(gopher.width), float32(gopher.height)),
					rl.NewRectangle(float32(newAsteroidX), float32(newAsteroidY), float32(asteroid.width), float32(asteroid.height)),
				) {
					renderScreen = GameEnding
					gopher.update(gopher.xPos, gopher.yPos, explosionTexture)
					rl.PlaySoundMulti(explodeSound)
					break
				}
			}
		} else {
			rl.StopMusicStream(gameplayMusic)
		}

		//Draw the main game play textures
		if renderScreen == GamePlay || renderScreen == GameEnding {
			rl.DrawText(fmt.Sprintf("Score: %d", int(score)), 10, 10, 20, rl.White)
			rl.DrawTexture(gopher.texture, gopher.xPos, gopher.yPos, rl.White)
			for _, asteroid := range asteroids {
				rl.DrawTexture(asteroid.texture, asteroid.xPos, asteroid.yPos, rl.White)
			}

			if renderScreen == GameEnding {
				gameEndingIterations++

				if gameEndingIterations > TARGET_FPS*2 {
					renderScreen = GameOver
				}
			}
		}

		if renderScreen == GameOver {
			if int(score) > highestScore {
				highestScore = int(score)
			}

			asteroids = nil
			txt := fmt.Sprintf(`
			Game Over
			Final Score: %d
			Press Enter to play again
			Press Esc to quit`, int(score))
			txtMeasurements := rl.MeasureText(txt, 20)
			rl.DrawText(txt, (SCREEN_WIDTH/2)-(txtMeasurements/2), SCREEN_HEIGHT/2-(txtMeasurements/4), 20, rl.Red)

			if rl.IsKeyPressed(rl.KeyEnter) {
				//Reset our gameplay variables
				gamePlayStartValues := setupNewGame(gopherDownTexture, asteroidTexture)
				score = gamePlayStartValues.score
				gopher = gamePlayStartValues.gopher
				asteroids = gamePlayStartValues.asteroids
				renderScreen = gamePlayStartValues.renderScreen
				gameEndingIterations = gamePlayStartValues.gameEndingIterations
			}
		}

		rl.EndDrawing()
	}

	rl.StopSoundMulti()
	rl.UnloadSound(explodeSound)
	rl.UnloadMusicStream(gameplayMusic)
	rl.UnloadTexture(gopherUpTexture)
	rl.UnloadTexture(gopherDownTexture)
	rl.UnloadTexture(asteroidTexture)
	rl.UnloadTexture(explosionTexture)
	rl.CloseWindow()
}
