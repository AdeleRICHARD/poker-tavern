import Phaser from "phaser";
import { TavernScene } from "./scenes/TavernScene";

export class GameManager {
  private game: Phaser.Game | null = null;
  private currentScene: TavernScene | null = null;

  constructor() {
    // Default configuration
  }

  init(containerId: string): Phaser.Game {
    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 800,
      height: 600,
      parent: containerId,
      backgroundColor: "#2c3e50",
      scene: [TavernScene],
      physics: {
        default: "arcade",
        arcade: {
          gravity: { x: 0, y: 0 }, // No gravity for top-down view
          debug: false,
        },
      },
      scale: {
        mode: Phaser.Scale.FIT,
        autoCenter: Phaser.Scale.CENTER_BOTH,
        min: {
          width: 400,
          height: 300,
        },
        max: {
          width: 1200,
          height: 900,
        },
      },
      render: {
        antialias: true,
        pixelArt: true, // For pixel art style
      },
    };

    this.game = new Phaser.Game(config);
    return this.game;
  }

  destroy(): void {
    if (this.game) {
      this.game.destroy(true);
      this.game = null;
      this.currentScene = null;
    }
  }

  getGame(): Phaser.Game | null {
    return this.game;
  }

  getCurrentScene(): TavernScene | null {
    if (this.game) {
      return this.game.scene.getScene("TavernScene") as TavernScene;
    }
    return null;
  }

  // Methods to communicate with Vue
  addPlayer(playerId: string, playerData: any): void {
    const scene = this.getCurrentScene();
    if (scene) {
      scene.addPlayer(playerId, playerData);
    }
  }

  syncInitialPlayers(players: any[]): void {
    const scene = this.getCurrentScene();
    if (scene) {
      players.forEach((player) => {
        scene.addPlayer(player.id, player);
        if (player.hasVoted) {
          // Delay to ensure sprite is created
          setTimeout(() => {
            scene.updatePlayerVote(player.id, true);
          }, 100);
        }
      });
    }
  }

  removePlayer(playerId: string): void {
    const scene = this.getCurrentScene();
    if (scene) {
      scene.removePlayer(playerId);
    }
  }

  updatePlayerVote(playerId: string, hasVoted: boolean): void {
    const scene = this.getCurrentScene();
    if (scene) {
      scene.updatePlayerVote(playerId, hasVoted);
    }
  }

  revealAllVotes(votes: { [playerId: string]: string }): void {
    const scene = this.getCurrentScene();
    if (scene) {
      scene.revealAllVotes(votes);
    }
  }

  resetTable(): void {
    const scene = this.getCurrentScene();
    if (scene) {
      scene.resetTable();
    }
  }

  // Utility methods
  resize(width: number, height: number): void {
    if (this.game) {
      this.game.scale.resize(width, height);
    }
  }

  pause(): void {
    if (this.game) {
      this.game.scene.pause("TavernScene");
    }
  }

  resume(): void {
    if (this.game) {
      this.game.scene.resume("TavernScene");
    }
  }
}
