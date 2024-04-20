# Ghost of Kyiv - README
-x
## Overview

"Ghost of Kyiv" is a simple 2D shooter game developed in the Go programming language using the Ebiten game library. The game features different modes, including a competition mode and a story mode. The player controls a character to shoot enemies, collect power-ups, and navigate through levels. The game supports multiple languages, including English and Ukrainian.


## Technologies Used

The "Ghost of Kyiv" game is created using several technologies, libraries, and programming techniques:

1. **Go Programming Language:** The game is primarily written in the Go programming language, known for its simplicity and performance.

2. **Ebiten Game Library:** Ebiten is a 2D game library for the Go programming language, used for handling game graphics, input, and audio.

3. **Audio Handling:** The game utilizes Ebiten's audio package to manage sound effects and music.

4. **Video Playback:** Video playback is incorporated into the game using the `tinne26/mpegg` library to display videos during gameplay.

5. **Docker:** The game was containerized using Docker for easy development and deployment.

6. **Image Assets:** Various image assets, such as background images, player, bullets, enemies, power-ups, and buttons, are loaded and displayed using the `ebitenutil.NewImageFromFile` function.

7. **Text Rendering:** Text is rendered on the game screen using the `text.Draw` function, and custom fonts are used for text rendering.

8. **Randomization:** The game uses Go's `rand` package for randomization, allowing for randomized enemy spawning and power-up placement.

9. **Object-Oriented Design:** The game is structured using object-oriented programming principles. It includes classes like `Game`, `Player`, `Enemy`, and more to manage game entities and their behavior.

## Key Functions and Techniques

Here are some key functions and techniques used in creating the "Ghost of Kyiv" game:

1. **Game Loop:** The game utilizes the Ebiten game loop, where the `Update` method is called to update the game logic, and the `Draw` method is called to render the game.

2. **Multiple Game Modes:** The game offers two distinct modes: Competition and Story. The player can choose between these modes at the start of the game.

3. **Language Selection:** Players can choose between English and Ukrainian at the start of the game, and the game is localized accordingly.

4. **Player Controls:** Player movement is controlled using arrow keys, and the player can shoot bullets in response to key presses.

5. **Enemy Behavior:** Enemies move downward on the screen, and they shoot bullets randomly.

6. **Boss Battles:** The game features a boss battle with a unique boss character that has specific behaviors, health, and shooting patterns.

7. **Power-Ups:** Power-ups are collected by the player to gain extra lives.

8. **Audio and Video:** Sound and video effects are incorporated into the game, creating a more immersive experience.

## How to Play

1. Choose your preferred language at the start: "E" for English or "U" for Ukrainian.

2. In Competition mode, press "1" to play, or press "2" for Story mode, which includes multiple levels and a boss battle.

3. Navigate the player character using the arrow keys.

4. Player automatically shoot bullets.

5. Enemies tries to destroy Player aircraft using their auto-aim bullets and their own aircrafts.

6. Collect power-ups to gain extra lives and increase your chances of success.

7. Defeat enemies and bosses to increase your score and advance through the game.

8. If you lose all lives, the game is over. Press "Enter" to restart or "Escape" to return to the start screen.

9. You can pause and resume the game when needed.

Enjoy playing "Ghost of Kyiv"!

## Credits

This game was developed by maxg95. The Ebiten game library, fonts, and other assets used in the game are credited to their respective creators.
