package main
!
import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text"

	"github.com/tinne26/mpegg"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Language string

const (
	English   Language = "en"
	Ukrainian Language = "ua"
)

type GameMode int

const (
	Competition GameMode = iota
	Story
)

type StoryChapter int

const (
	Chapter1 StoryChapter = iota
)

type StoryLevel int

const (
	Level1 StoryLevel = iota
	Level2
	Level3
)

type Player struct {
	x, y          float64
	speed         float64
	isShooting    bool
	shootCoolDown int
}

func (p Player) getX() float64 {
	return p.x
}

func (p Player) getY() float64 {
	return p.y
}

type PlayerBullet struct {
	x, y   float64
	speed  float64
	active bool
}

func (pb PlayerBullet) getX() float64 {
	return pb.x
}

func (pb PlayerBullet) getY() float64 {
	return pb.y
}

type Enemy struct {
	x, y    float64
	speedY  float64
	active  bool
	hasShot bool
}

func (e Enemy) getX() float64 {
	return e.x
}

func (e Enemy) getY() float64 {
	return e.y
}

type EnemyBullet struct {
	x, y   float64
	speedX float64
	speedY float64
	active bool
}

func (eb EnemyBullet) getX() float64 {
	return eb.x
}

func (eb EnemyBullet) getY() float64 {
	return eb.y
}

type Boss struct {
	x, y           float64
	speedX, speedY float64
	active         bool
	health         int
	shotCooldown   int
	shotCounter    int
}

func (b Boss) getX() float64 {
	return b.x
}

func (b Boss) getY() float64 {
	return b.y
}

type PowerUp struct {
	x, y   float64
	active bool
	speedY float64
}

func (p PowerUp) getX() float64 {
	return p.x
}

func (p PowerUp) getY() float64 {
	return p.y
}

type Game struct {
	gameMode                    GameMode
	player                      Player
	enemies                     []Enemy
	playerBullets               []PlayerBullet
	score                       int
	playerLives                 int
	isGameOver                  bool
	frameCount                  int // Keep track of frames for shooting timer
	bgOffsetY                   float64
	powerUp                     PowerUp
	powerUpCounter              int
	enemyBullets                []EnemyBullet
	isPaused                    bool
	pauseImage, resumeImage     *ebiten.Image
	clickedButton               bool
	startScreenActive           bool
	audioContext                *audio.Context
	musicPlayer                 *audio.Player
	shootingPlayer              *audio.Player
	startSound                  *audio.Player
	startSoundPlayer            *audio.Player
	storyChapter                StoryChapter // Track the current chapter in story mode
	storyLevel                  StoryLevel   // Track the current level in story mode
	showLevelScreen             bool
	showLevelCompleted          bool
	levelScreenCounter          int
	levelCompletedScreenCounter int
	levelScreenShown            bool
	gameCompleted               bool
	boss                        Boss
	isBossActive                bool
	currentChapter              StoryChapter
	showStartButton             bool
	startButtonImage            *ebiten.Image
	language                    Language
	languageScreenActive        bool
	videoPlayer                 *mpegg.Player
	videoScreenCounter          int
	showVideoScreen             bool
}

var (
	mplusNormalFont font.Face
)

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	mplusNormalFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    12,
		DPI:     dpi,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func musicStream(context *audio.Context) (*audio.Player, error) {
	file, err := ebitenutil.OpenFile("assets/gameplay.mp3")
	if err != nil {
		return nil, err
	}

	mp3Stream, err := mp3.Decode(context, file)
	if err != nil {
		return nil, err
	}

	player, err := audio.NewPlayer(context, mp3Stream)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func shootingSoundStream(context *audio.Context) (*audio.Player, error) {
	file, err := ebitenutil.OpenFile("assets/shooting.mp3")
	if err != nil {
		return nil, err
	}

	mp3Stream, err := mp3.Decode(context, file)
	if err != nil {
		return nil, err
	}

	player, err := audio.NewPlayer(context, mp3Stream)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func loadSound(context *audio.Context, path string) (*audio.Player, error) {
	file, err := ebitenutil.OpenFile(path)
	if err != nil {
		return nil, err
	}

	mp3Stream, err := mp3.Decode(context, file)
	if err != nil {
		return nil, err
	}

	player, err := audio.NewPlayer(context, mp3Stream)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func collision(a, b interface {
	getX() float64
	getY() float64
}) bool {
	return a.getX() < b.getX()+32 && a.getX()+32 > b.getX() && a.getY() < b.getY()+32 && a.getY()+32 > b.getY()
}

func (g *Game) initializeGame() {
	g.player.x = screenWidth / 2
	g.player.y = screenHeight - 50
	g.enemies = nil
	g.playerBullets = nil
	g.isGameOver = false
	g.playerLives = 3
	g.score = 0
	g.enemyBullets = nil
	g.bgOffsetY = 0
	g.powerUp = PowerUp{}
	g.powerUpCounter = 0
	g.isPaused = false
	g.clickedButton = false
	var err error
	g.musicPlayer, err = musicStream(g.audioContext)
	if err != nil {
		log.Fatal(err)
	}
	g.musicPlayer.Play()
	g.shootingPlayer, err = shootingSoundStream(g.audioContext)
	if err != nil {
		log.Fatal(err)
	}
	// Stop and close the start sound player
	if g.startSoundPlayer != nil {
		g.startSoundPlayer.Close()
	}
	switch g.language {
	case English:
		src, err := os.Open("assets/testdata_test.mpg")
		if err != nil {
			log.Fatal(err)
		}
		videoPlayer, err := mpegg.NewPlayer(src)
		if err != nil {
			log.Fatal(err)
		}
		g.videoPlayer = videoPlayer
	case Ukrainian:
		src, err := os.Open("assets/testdata_test.mpg")
		if err != nil {
			log.Fatal(err)
		}
		videoPlayer, err := mpegg.NewPlayer(src)
		if err != nil {
			log.Fatal(err)
		}
		g.videoPlayer = videoPlayer
	}
}

func (g *Game) initializeLevel(level StoryLevel) {
	switch level {
	case Level1:
		// Initialize settings for Level 1
		g.enemies = nil
		g.playerBullets = nil
		g.score = 0
	case Level2:
		// Initialize settings for Level 2
		g.enemies = nil
		g.playerBullets = nil
		g.score = 0
	case Level3:
		// Initialize settings for Level 3
		g.enemies = nil
		g.playerBullets = nil
		g.score = 0
	}
}

func (g *Game) Update() error {
	// Handle language selection input
	if g.languageScreenActive {
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			g.language = English
			g.languageScreenActive = false
			g.startScreenActive = true
		} else if ebiten.IsKeyPressed(ebiten.KeyU) {
			g.language = Ukrainian
			g.languageScreenActive = false
			g.startScreenActive = true
		}
		return nil
	}
	if g.startScreenActive {
		if ebiten.IsKeyPressed(ebiten.Key1) {
			g.gameMode = Competition
			g.startScreenActive = false
			g.initializeGame()
		} else if ebiten.IsKeyPressed(ebiten.Key2) {
			g.gameMode = Story
			g.startScreenActive = false
			g.storyChapter = Chapter1 // Start with the first chapter
			g.storyLevel = Level1     // Start with the first level
			g.initializeLevel(Level1) // Initialize Level 1
			g.initializeGame()
		}
		return nil
	}

	if g.gameMode == Story {
		LevelScreenDuration := 3 * 60
		VideoScreenDuration := 3 * 60
		if !g.videoPlayer.IsPlaying() {
			g.showVideoScreen = true
			g.videoScreenCounter = VideoScreenDuration
			g.videoPlayer.Play()
		} else if g.videoPlayer.IsPlaying() {
			// Countdown the level screen timer
			g.videoScreenCounter--
			if g.videoScreenCounter <= 0 {
				g.videoPlayer.Pause()
				g.showVideoScreen = false
			}
			return nil
		}
		switch g.storyChapter {
		case Chapter1:
			switch g.storyLevel {
			case Level1:
				if !g.showVideoScreen && !g.levelScreenShown {
					// Display "Level 1" screen
					g.showLevelScreen = true
					g.levelScreenCounter = LevelScreenDuration
					g.levelScreenShown = true
					return nil
				} else if g.showLevelScreen {
					// Countdown the level screen timer
					g.levelScreenCounter--
					if g.levelScreenCounter <= 0 {
						g.showLevelScreen = false // Hide "Level 1" screen
					}
					return nil
				} else if g.score >= 1 && !g.showLevelCompleted {
					// Display "Level 1 completed" screen
					g.showLevelCompleted = true
					g.levelCompletedScreenCounter = LevelScreenDuration
				} else if g.showLevelCompleted {
					// Countdown the level completed screen timer
					g.levelCompletedScreenCounter--
					if g.levelCompletedScreenCounter <= 0 {
						g.showLevelCompleted = false // Hide "Level 1 completed" screen
						g.storyLevel = Level2        // Start Level 2
						g.initializeLevel(Level2)
					}
					if g.storyLevel == Level2 {
						// Clear Level 1 enemies and bullets
						g.enemies = nil
						g.playerBullets = nil
						g.enemyBullets = nil
					}
				}
			case Level2:
				if !g.levelScreenShown {
					g.showLevelScreen = true
					g.levelScreenCounter = LevelScreenDuration
					g.levelScreenShown = true
					return nil
				} else if g.showLevelScreen {
					g.levelScreenCounter--
					if g.levelScreenCounter <= 0 {
						g.showLevelScreen = false
					}
					return nil
				} else if g.score >= 2 && !g.showLevelCompleted {
					g.showLevelCompleted = true
					g.levelCompletedScreenCounter = LevelScreenDuration
				} else if g.showLevelCompleted {
					g.levelCompletedScreenCounter--
					if g.levelCompletedScreenCounter <= 0 {
						g.showLevelCompleted = false
						g.storyLevel = Level3
						g.initializeLevel(Level3)
					}
					if g.storyLevel == Level3 {
						g.enemies = nil
						g.playerBullets = nil
						g.enemyBullets = nil
					}
				}
			case Level3:
				if !g.levelScreenShown {
					g.showLevelScreen = true
					g.levelScreenCounter = LevelScreenDuration
					g.levelScreenShown = true
					return nil
				} else if g.showLevelScreen {
					g.levelScreenCounter--
					if g.levelScreenCounter <= 0 {
						g.showLevelScreen = false
					}
					return nil
				}
				if g.score >= 2 && !g.gameCompleted && !g.isBossActive {
					g.enemies = nil
					g.enemyBullets = nil
					g.isBossActive = true
					g.boss = Boss{
						x:            screenWidth / 2,
						y:            50,
						speedX:       2,
						speedY:       2,
						active:       true,
						health:       10,
						shotCooldown: 60,
					}
					if !g.boss.active && !g.isBossActive {
						g.gameCompleted = true
					}
					if g.gameCompleted {
						// Show the "Game Completed" screen and stop the game
						return nil
					}
				}
				// We should add new chapters and levels here
			}
		}
	}

	// Handle pause/resume button click
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mouseX, mouseY := ebiten.CursorPosition()
		bX := screenWidth - 50
		bY := screenHeight - 470
		buttonWidth := 32
		buttonHeight := 32
		if mouseX >= bX && mouseX <= bX+buttonWidth && mouseY >= bY && mouseY <= bY+buttonHeight {
			// Click occurred within the button area
			if g.isPaused {
				g.isPaused = false
			} else {
				g.isPaused = true
			}
			g.clickedButton = true
		}
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		buttonX := screenWidth - 110
		buttonY := screenHeight - 470
		buttonWidth := 32
		buttonHeight := 32

		// Check if the left mouse click was within the button's bounds
		if mx >= buttonX && mx <= buttonX+buttonWidth && my >= buttonY && my <= buttonY+buttonHeight {
			g.startScreenActive = true
			g.clickedButton = true
		} else {
			g.clickedButton = false
		}
	}

	if g.isPaused {
		// Game is paused and game logic doesn't updates
		return nil
	}

	// Check for game over condition
	if g.playerLives <= 0 {
		g.isGameOver = true
	}

	if g.isGameOver {
		if ebiten.IsKeyPressed(ebiten.KeyEnter) {
			// Restart
			g.initializeGame()
		} else if ebiten.IsKeyPressed(ebiten.KeyEscape) {
			// Go to start screen
			g.startScreenActive = true
		}
		return nil
	}

	// Player controls
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.x -= g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.x += g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.y -= g.player.speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.y += g.player.speed
	}

	// Ensure player stays within screen bounds
	g.player.x = clamp(g.player.x, 0, screenWidth-32)
	g.player.y = clamp(g.player.y, 0, screenHeight-32)
	// Shooting behavior
	shootCooldown := 20
	var shootSpeed float64 = 5
	if g.frameCount%shootCooldown == 0 {
		playerCenterX := g.player.x + 17 - 2 // 17 is half of the player image width (34/2) and 2 is half of the bullet image width (4/2).
		g.playerBullets = append(g.playerBullets, PlayerBullet{x: playerCenterX, y: g.player.y, speed: shootSpeed, active: true})
	}

	// Update player bullets
	for i := range g.playerBullets {
		if g.playerBullets[i].active {
			g.playerBullets[i].y -= g.playerBullets[i].speed
		}
	}

	// Increment the frame count
	g.frameCount++

	// Update enemies and ensure they move on the y-axis from top to bottom
	for i := range g.enemies {
		if g.enemies[i].active {
			g.enemies[i].y += g.enemies[i].speedY
			if g.enemies[i].y > (screenHeight - 32) {
				g.enemies[i].active = false
			}
			// Check for collision with player bullets
			for j := range g.playerBullets {
				if g.playerBullets[j].active && collision(g.enemies[i], g.playerBullets[j]) {
					g.enemies[i].active = false
					g.playerBullets[j].active = false
					g.score++
					// Play the shooting sound effect
					if err := g.shootingPlayer.Rewind(); err != nil {
						log.Fatal(err)
					}
					g.shootingPlayer.Play()
				}
			}
			// Check for collision with player
			if collision(g.enemies[i], Enemy{x: g.player.x, y: g.player.y}) {
				g.playerLives--
				g.player.x = screenWidth / 2
				g.player.y = screenHeight - 50
			}
		}
	}

	// Spawn enemies
	if rand.Intn(100) == 2 {
		speedY := rand.Float64() + 3
		g.enemies = append(g.enemies, Enemy{x: rand.Float64() * screenWidth, y: 0, speedY: speedY, active: true})
	}

	// Update enemies and ensure they stay within screen bounds
	for i := range g.enemies {
		if g.enemies[i].active {
			g.enemies[i].x = clamp(g.enemies[i].x, 0, screenWidth-32)
			g.enemies[i].y = clamp(g.enemies[i].y, 0, screenHeight-32)
		}
	}

	// Update enemy bullets
	for i := range g.enemyBullets {
		if g.enemyBullets[i].active {
			g.enemyBullets[i].x += g.enemyBullets[i].speedX
			g.enemyBullets[i].y += g.enemyBullets[i].speedY

			// Check for collision with player
			if collision(g.player, g.enemyBullets[i]) {
				g.playerLives--
				g.enemyBullets[i].active = false
			}
		}
	}

	// Enemy shooting logic
	for i := range g.enemies {
		if g.enemies[i].active && !g.enemies[i].hasShot && rand.Intn(100) < 2 {
			// Calculate bullet direction towards the player
			dx := g.player.x - g.enemies[i].x
			dy := g.player.y - g.enemies[i].y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Create an enemy bullet with the direction towards the player
			if distance != 0 {
				speedX := (dx / distance) * 1
				speedY := (dy / distance) * 1
				enemyCenterX := g.enemies[i].x + 17 - 2 // 17 is half of the player image width (34/2) and 2 is half of the bullet image width (4/2).
				g.enemyBullets = append(g.enemyBullets, EnemyBullet{
					x:      enemyCenterX,
					y:      g.enemies[i].y,
					speedX: speedX,
					speedY: speedY,
					active: true,
				})

				// Mark the enemy as having shot a bullet
				g.enemies[i].hasShot = true
			}
		}
	}

	// Handle boss logic (only if the boss is active)
	if g.isBossActive {
		// Randomly change the boss's direction every few frames
		if g.frameCount%120 == 0 { // Change direction every 2 seconds
			g.boss.speedX = rand.Float64()*4 - 2 // Random speed between -2 and 2 for left-right movement
			g.boss.speedY = rand.Float64()*4 - 2 // Random speed between -2 and 2 for top-bottom movement
		}

		// Move the boss
		g.boss.x += g.boss.speedX
		g.boss.y += g.boss.speedY

		// Ensure boss stays within screen bounds
		g.boss.x = clamp(g.boss.x, 0, screenWidth-32)
		g.boss.y = clamp(g.boss.y, 0, screenHeight-32)

		// Boss shooting logic
		g.boss.shotCounter++
		if g.boss.shotCounter >= g.boss.shotCooldown {
			// Calculate bullet direction towards the player
			dx := g.player.x - g.boss.x
			dy := g.player.y - g.boss.y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Create boss bullets with the direction towards the player
			if distance != 0 {
				speedX := (dx / distance) * 2
				speedY := (dy / distance) * 2
				g.enemyBullets = append(g.enemyBullets, EnemyBullet{
					x:      g.boss.x,
					y:      g.boss.y,
					speedX: speedX,
					speedY: speedY,
					active: true,
				})

				// Reset the shot counter
				g.boss.shotCounter = 0
			}
		}

		// Check for collision with player bullets
		for j := range g.playerBullets {
			if g.playerBullets[j].active && collision(g.boss, g.playerBullets[j]) {
				g.boss.health--
				g.playerBullets[j].active = false
				if err := g.shootingPlayer.Rewind(); err != nil {
					log.Fatal(err)
				}
				g.shootingPlayer.Play()

				// Check if the boss has been defeated
				if g.boss.health <= 0 {
					g.boss.active = false
					g.isBossActive = false
					g.gameCompleted = true
					return nil
				}
			}
		}
	}

	// Update power-up
	powerUpRespawnTime := 30 * 60
	g.powerUpCounter++
	if g.powerUpCounter >= powerUpRespawnTime {
		// If the power-up has been inactive for the specified time, respawn it randomly on the map
		g.powerUp.x = rand.Float64() * screenWidth
		g.powerUp.y = rand.Float64() * screenHeight
		g.powerUp.active = true
		g.powerUpCounter = 0 // Reset the timer
	}

	// Update power-up position for movement
	g.powerUp.speedY = 1
	if g.powerUp.active {
		g.powerUp.y += g.powerUp.speedY

		// Check for collision with power-up
		if g.powerUp.active && collision(g.player, g.powerUp) && g.playerLives < 3 {
			g.playerLives++
			g.powerUp.active = false // Deactivate the power-up after collecting
		} else if g.powerUp.active && collision(g.player, g.powerUp) && g.playerLives == 3 {
			g.powerUp.active = false
		}
		if g.powerUp.y > (screenHeight - 32) {
			g.powerUp.active = false
		}
	}

	// Update the background scrolling
	g.bgOffsetY += 2

	// Check if the audio player has finished playing
	if g.musicPlayer.IsPlaying() == false {
		g.musicPlayer.Rewind()
		g.musicPlayer.Play()
	}

	return nil
}

func (g *Game) drawLanguageScreen(screen *ebiten.Image) {
	text.Draw(screen, "E for English", mplusNormalFont, 100, 180, color.White)
	text.Draw(screen, "U щоб обрати Українську", mplusNormalFont, 100, 200, color.White)
}

func (g *Game) drawStartScreen(screen *ebiten.Image) {
	// Only play the start sound if it's not playing
	if g.startSoundPlayer == nil || !g.startSoundPlayer.IsPlaying() {
		startSoundPlayer, err := loadSound(g.audioContext, "assets/start.mp3")
		if err != nil {
			log.Fatal(err)
		}
		g.startSoundPlayer = startSoundPlayer
		g.startSoundPlayer.Play()
	}
	// Draw the start screen
	switch g.language {
	case English:
		ebitenutil.DebugPrint(screen, "Choose Game Mode:")
		ebitenutil.DebugPrintAt(screen, "1. Competition", 100, 180)
		ebitenutil.DebugPrintAt(screen, "2. Story", 100, 200)
	case Ukrainian:
		text.Draw(screen, "Виберіть ігровий режим:", mplusNormalFont, 20, 80, color.White)
		text.Draw(screen, "1. Змагання", mplusNormalFont, 100, 180, color.White)
		text.Draw(screen, "2. Історія", mplusNormalFont, 100, 200, color.White)
	}
}

func (g *Game) drawGameCompleted(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Game Completed")
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.languageScreenActive {
		g.drawLanguageScreen(screen)
		return
	}
	if g.startScreenActive {
		g.drawStartScreen(screen)
		return
	}

	if g.gameMode == Story {
		if g.showLevelScreen {
			text := fmt.Sprintf("Level %d", g.storyLevel+1)
			ebitenutil.DebugPrint(screen, text)
			return
		} else if g.videoPlayer.IsPlaying() {
			// Draw the video
			mpegg.Draw(screen, g.videoPlayer.CurrentFrame())
			return
		} else if g.showLevelCompleted {
			text := fmt.Sprintf("Level %d completed\nStarting level %d", g.storyLevel+1, g.storyLevel+2)
			ebitenutil.DebugPrint(screen, text)
			return
		}
		if g.gameCompleted {
			g.drawGameCompleted(screen)
			return
		}
	}

	if g.gameMode == Competition {
		// Draw background image for competition mode
		effectiveY := int(g.bgOffsetY) % backgroundImage.Bounds().Dy()

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(effectiveY))
		screen.DrawImage(backgroundImage, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(effectiveY-backgroundImage.Bounds().Dy()))
		screen.DrawImage(backgroundImage, op)
	} else if g.gameMode == Story {
		// Draw background image for story mode
		var currentChapterBackground *ebiten.Image

		switch g.currentChapter {
		case Chapter1:
			currentChapterBackground = chapterBackgroundImage
		}

		// Calculate the effective position of the background image by taking the modulo
		effectiveY := int(g.bgOffsetY) % currentChapterBackground.Bounds().Dy()

		// Draw the background image at the current effective position
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(effectiveY))
		screen.DrawImage(currentChapterBackground, op)

		// Draw the background image again just below the first one to create the illusion of an infinite loop
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(effectiveY-currentChapterBackground.Bounds().Dy()))
		screen.DrawImage(currentChapterBackground, op)
	}

	// Draw pause/resume button
	pauseButtonOp := &ebiten.DrawImageOptions{}
	pauseButtonOp.GeoM.Translate(screenWidth-50, screenHeight-470)

	if !g.isPaused {
		screen.DrawImage(g.pauseImage, pauseButtonOp)
	} else {
		screen.DrawImage(g.resumeImage, pauseButtonOp)
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenWidth-110, screenHeight-470)
	screen.DrawImage(g.startButtonImage, opts)

	// Draw an indicator on the button if it was clicked
	if g.clickedButton {
		// Draw a circle or border around the button to indicate the click
		buttonIndicatorOp := &ebiten.DrawImageOptions{}
		buttonIndicatorOp.GeoM.Translate(screenWidth-50, screenHeight-470)
		buttonIndicatorOp.ColorM.Scale(1, 0, 0, 0.5)
		if !g.isPaused {
			screen.DrawImage(g.pauseImage, buttonIndicatorOp)
		} else {
			screen.DrawImage(g.resumeImage, buttonIndicatorOp)
		}
	}

	// Reset the clickedButton flag
	g.clickedButton = false

	if g.isGameOver {
		ebitenutil.DebugPrint(screen, "Game Over. Press Enter to Restart or Escape to Exit")
		return
	}

	// Draw player
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.player.x, g.player.y)
	screen.DrawImage(playerImage, op)

	// Draw player bullets
	for i := range g.playerBullets {
		if g.playerBullets[i].active {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.playerBullets[i].x, g.playerBullets[i].y)
			screen.DrawImage(bulletImage, op)
		}
	}

	// Draw enemies
	for i := range g.enemies {
		if g.enemies[i].active {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.enemies[i].x, g.enemies[i].y)
			screen.DrawImage(enemyImage, op)
		}
	}

	// Draw enemy bullets
	for i := range g.enemyBullets {
		if g.enemyBullets[i].active {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.enemyBullets[i].x, g.enemyBullets[i].y)
			screen.DrawImage(enemyBulletImage, op)
		}
	}

	// Draw the boss if it's active
	if g.isBossActive && g.boss.active {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.boss.x, g.boss.y)
		screen.DrawImage(bossImage, op) // You can use the enemy image for the boss
	}

	// Draw the power-up if it's active
	if g.powerUp.active {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.powerUp.x, g.powerUp.y)
		screen.DrawImage(powerUpImage, op)
	}

	// Draw score and lives
	if !g.isGameOver {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d   Lives: %d", g.score, g.playerLives))
	} else {
		ebitenutil.DebugPrint(screen, "Game Over. Press Enter to Restart or Escape to Exit")
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

var (
	backgroundImage, playerImage, bulletImage, enemyImage, powerUpImage, enemyBulletImage, bossImage, chapterBackgroundImage *ebiten.Image
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Load images
	var err error
	backgroundImage, _, err = ebitenutil.NewImageFromFile("assets/background.png")
	if err != nil {
		log.Fatal(err)
	}
	playerImage, _, err = ebitenutil.NewImageFromFile("assets/player.png")
	if err != nil {
		log.Fatal(err)
	}
	bulletImage, _, err = ebitenutil.NewImageFromFile("assets/bullet.png")
	if err != nil {
		log.Fatal(err)
	}
	enemyImage, _, err = ebitenutil.NewImageFromFile("assets/enemy.png")
	if err != nil {
		log.Fatal(err)
	}
	powerUpImage, _, err = ebitenutil.NewImageFromFile("assets/powerup.png")
	if err != nil {
		log.Fatal(err)
	}
	enemyBulletImage, _, err = ebitenutil.NewImageFromFile("assets/enemy_bullet.png")
	if err != nil {
		log.Fatal(err)
	}
	pauseImage, _, err := ebitenutil.NewImageFromFile("assets/pause.png")
	if err != nil {
		log.Fatal(err)
	}
	resumeImage, _, err := ebitenutil.NewImageFromFile("assets/resume.png")
	if err != nil {
		log.Fatal(err)
	}
	bossImage, _, err = ebitenutil.NewImageFromFile("assets/boss.png")
	if err != nil {
		log.Fatal(err)
	}
	chapterBackgroundImage, _, err = ebitenutil.NewImageFromFile("assets/chapter_background.png")
	if err != nil {
		log.Fatal(err)
	}
	startButtonImage, _, err := ebitenutil.NewImageFromFile("assets/start_button.png")
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the game
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ghost of Kyiv")

	// Create the audio context
	audioContext := audio.NewContext(44100)

	// Load the start sound effect
	startSound, err := loadSound(audioContext, "assets/start.mp3")
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		player:               Player{x: screenWidth / 2, y: screenHeight - 50, speed: 4},
		frameCount:           0,
		pauseImage:           pauseImage,
		resumeImage:          resumeImage,
		languageScreenActive: true, // Game start with the language screen
		audioContext:         audioContext,
		startSound:           startSound,
		startButtonImage:     startButtonImage,
	}

	// Start the game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
