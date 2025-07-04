import Phaser from "phaser";

interface PlayerSprite {
  id: string;
  sprite: Phaser.GameObjects.Sprite;
  nameText: Phaser.GameObjects.Text;
  voteCard: Phaser.GameObjects.Sprite | null;
  statusIcon: Phaser.GameObjects.Sprite | null;
  position: { x: number; y: number };
  data: any;
}

export class TavernScene extends Phaser.Scene {
  private players: Map<string, PlayerSprite> = new Map();
  private tableSprite!: Phaser.GameObjects.Sprite;
  private backgroundSprite!: Phaser.GameObjects.Sprite;
  private tablePositions: { x: number; y: number }[] = [];
  private isVotesRevealed = false;

  constructor() {
    super({ key: "TavernScene" });
  }

  preload() {
    // Create procedural textures for WoW/pixel art style
    this.createTableTexture();
    this.createBackgroundTexture();
    this.createPlayerTextures();
    this.createCardTextures();
    this.createUITextures();
  }

  create() {
    // Tavern background
    this.backgroundSprite = this.add.sprite(400, 300, "tavern-bg");
    this.backgroundSprite.setDisplaySize(800, 600);

    // Round table in center
    this.tableSprite = this.add.sprite(400, 300, "round-table");
    this.tableSprite.setDisplaySize(300, 300);

    // Define positions around table (8 seats maximum)
    this.setupTablePositions();

    // Add decorative elements
    this.addTavernDecorations();

    // Table click events
    this.setupTableInteractions();

    // Information text
    this.add
      .text(400, 50, "🏰 Planning Poker Tavern", {
        fontSize: "24px",
        color: "#FFD700",
        fontFamily: "serif",
      })
      .setOrigin(0.5);
  }

  private createTableTexture() {
    const graphics = this.add.graphics();

    // Round wooden table
    graphics.fillStyle(0x8b4513); // Brown
    graphics.fillCircle(150, 150, 140);

    // Golden border
    graphics.lineStyle(6, 0xdaa520); // Gold
    graphics.strokeCircle(150, 150, 140);

    // Wood texture (lines)
    graphics.lineStyle(2, 0x654321);
    for (let i = 0; i < 8; i++) {
      const angle = (i * Math.PI * 2) / 8;
      const x1 = 150 + Math.cos(angle) * 100;
      const y1 = 150 + Math.sin(angle) * 100;
      const x2 = 150 + Math.cos(angle) * 130;
      const y2 = 150 + Math.sin(angle) * 130;
      graphics.lineBetween(x1, y1, x2, y2);
    }

    graphics.generateTexture("round-table", 300, 300);
    graphics.destroy();
  }

  private createBackgroundTexture() {
    const graphics = this.add.graphics();

    // Stone background
    graphics.fillStyle(0x2c3e50);
    graphics.fillRect(0, 0, 800, 600);

    // Stone pattern
    graphics.lineStyle(2, 0x34495e);
    for (let x = 0; x < 800; x += 100) {
      for (let y = 0; y < 600; y += 100) {
        graphics.strokeRect(x, y, 100, 100);
      }
    }

    // Ambient lighting (simulated gradients)
    graphics.fillStyle(0x3a4a5c);
    graphics.fillCircle(400, 300, 200);

    graphics.generateTexture("tavern-bg", 800, 600);
    graphics.destroy();
  }

  private createPlayerTextures() {
    // Create textures for each character class
    const classes = [
      { name: "mage", color: 0x4a90e2, symbol: "🧙" },
      { name: "paladin", color: 0xf5a623, symbol: "⚔️" },
      { name: "rogue", color: 0x7ed321, symbol: "🗡️" },
      { name: "priest", color: 0xd0021b, symbol: "✨" },
      { name: "warrior", color: 0xb68d00, symbol: "🛡️" },
      { name: "hunter", color: 0x50e3c2, symbol: "🏹" },
      { name: "warlock", color: 0x9013fe, symbol: "😈" },
      { name: "druid", color: 0x4a4a4a, symbol: "🌿" },
    ];

    classes.forEach((playerClass) => {
      const graphics = this.add.graphics();

      // Character body (colored circle)
      graphics.fillStyle(playerClass.color);
      graphics.fillCircle(25, 25, 20);

      // Border
      graphics.lineStyle(3, 0xffffff);
      graphics.strokeCircle(25, 25, 20);

      graphics.generateTexture(`player-${playerClass.name}`, 50, 50);
      graphics.destroy();
    });
  }

  private createCardTextures() {
    console.log("Creating card textures...");

    // Hidden card
    const hiddenGraphics = this.add.graphics();
    hiddenGraphics.fillStyle(0x8b4513); // Brown
    hiddenGraphics.fillRoundedRect(0, 0, 30, 40, 5);
    hiddenGraphics.lineStyle(2, 0xdaa520); // Gold border
    hiddenGraphics.strokeRoundedRect(0, 0, 30, 40, 5);
    hiddenGraphics.generateTexture("card-hidden", 30, 40);
    hiddenGraphics.destroy();
    console.log("Created card-hidden texture");

    // Revealed cards for each value
    const cardValues = ["0", "1", "2", "3", "5", "8", "13", "21", "?", "☕"];
    cardValues.forEach((value) => {
      const cardGraphics = this.add.graphics();
      cardGraphics.fillStyle(0xf4f4f4); // White
      cardGraphics.fillRoundedRect(0, 0, 30, 40, 5);
      cardGraphics.lineStyle(2, 0x333333);
      cardGraphics.strokeRoundedRect(0, 0, 30, 40, 5);

      // Add value text on the card
      const text = this.add
        .text(15, 20, value, {
          fontSize: "12px",
          color: "#333333",
          fontFamily: "Arial",
          fontStyle: "bold",
        })
        .setOrigin(0.5);

      cardGraphics.generateTexture(`card-${value}`, 30, 40);
      cardGraphics.destroy();
      text.destroy();
      console.log(`Created card-${value} texture`);
    });
  }

  private createUITextures() {
    // Voted icon
    const votedGraphics = this.add.graphics();
    votedGraphics.fillStyle(0x7ed321); // Green
    votedGraphics.fillCircle(10, 10, 8);
    votedGraphics.lineStyle(2, 0xffffff);
    votedGraphics.strokeCircle(10, 10, 8);
    votedGraphics.generateTexture("voted-icon", 20, 20);
    votedGraphics.destroy();

    // Waiting icon
    const waitingGraphics = this.add.graphics();
    waitingGraphics.fillStyle(0xf5a623); // Orange
    waitingGraphics.fillCircle(10, 10, 8);
    waitingGraphics.lineStyle(2, 0xffffff);
    waitingGraphics.strokeCircle(10, 10, 8);
    waitingGraphics.generateTexture("waiting-icon", 20, 20);
    waitingGraphics.destroy();
  }

  private setupTablePositions() {
    // 8 positions around the round table
    const centerX = 400;
    const centerY = 300;
    const radius = 180;

    for (let i = 0; i < 8; i++) {
      const angle = (i * Math.PI * 2) / 8 - Math.PI / 2; // Start at top
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      this.tablePositions.push({ x, y });
    }
  }

  private addTavernDecorations() {
    // Chandelier in center of table
    const chandelier = this.add
      .text(400, 300, "🕯️", {
        fontSize: "24px",
      })
      .setOrigin(0.5);

    // Some decorative elements
    this.add.text(100, 100, "🍺", { fontSize: "32px" });
    this.add.text(700, 500, "🗡️", { fontSize: "32px" });
    this.add.text(50, 550, "📜", { fontSize: "32px" });
    this.add.text(750, 80, "🏰", { fontSize: "32px" });

    // Torches on walls
    this.add.text(50, 300, "🔥", { fontSize: "28px" });
    this.add.text(750, 300, "🔥", { fontSize: "28px" });
  }

  private setupTableInteractions() {
    this.tableSprite.setInteractive();
    this.tableSprite.on("pointerdown", () => {
      // For now, just a visual effect
      this.tableSprite.setTint(0xffffaa);
      this.time.delayedCall(200, () => {
        this.tableSprite.clearTint();
      });
    });
  }

  // Public methods called from GameManager
  addPlayer(playerId: string, playerData: any) {
    console.log("TavernScene.addPlayer called:", playerId, playerData);

    // If player already exists, preserve their vote card
    let existingVoteCard: Phaser.GameObjects.Sprite | null = null;
    let existingVote: string | undefined = undefined;

    if (this.players.has(playerId)) {
      console.log("Player already exists, preserving vote state:", playerId);
      const existingPlayer = this.players.get(playerId);
      if (existingPlayer) {
        existingVoteCard = existingPlayer.voteCard;
        existingVote = existingPlayer.data.vote;
        // Only remove visual elements without the card
        existingPlayer.sprite.destroy();
        existingPlayer.nameText.destroy();
        existingPlayer.statusIcon?.destroy();
        this.players.delete(playerId);
      }
    }

    // Find a free position
    const occupiedPositions = Array.from(this.players.values()).map(
      (p) => p.position,
    );
    const freePosition = this.tablePositions.find(
      (pos) =>
        !occupiedPositions.some(
          (occupied) =>
            Math.abs(occupied.x - pos.x) < 10 &&
            Math.abs(occupied.y - pos.y) < 10,
        ),
    );

    if (!freePosition) {
      console.warn("No free position at the table");
      return;
    }

    console.log("Adding player at position:", freePosition);

    // Create player sprite
    const playerSprite = this.add.sprite(
      freePosition.x,
      freePosition.y,
      `player-${playerData.character}`,
    );
    playerSprite.setDisplaySize(60, 60);

    // Player name
    const nameText = this.add
      .text(freePosition.x, freePosition.y + 40, playerData.name, {
        fontSize: "14px",
        color: "#FFD700",
        fontFamily: "serif",
      })
      .setOrigin(0.5);

    // Status icon
    const statusIcon = this.add.sprite(
      freePosition.x + 25,
      freePosition.y - 25,
      "waiting-icon",
    );

    const playerObj: PlayerSprite = {
      id: playerId,
      sprite: playerSprite,
      nameText,
      voteCard: existingVoteCard, // Réutiliser la carte existante
      statusIcon,
      position: freePosition,
      data: playerData,
    };

    // If we had an existing card, reposition it
    if (existingVoteCard) {
      existingVoteCard.setPosition(freePosition.x, freePosition.y - 60);
      console.log("Repositioned existing vote card for player:", playerId);
    }

    this.players.set(playerId, playerObj);
    console.log(
      "Player added to scene:",
      playerId,
      "Total players:",
      this.players.size,
    );

    // Appearance animation
    playerSprite.setScale(0);
    this.tweens.add({
      targets: playerSprite,
      scale: 1,
      duration: 500,
      ease: "Back.easeOut",
    });
  }

  removePlayer(playerId: string) {
    const player = this.players.get(playerId);
    if (!player) return;

    // Disappearance animation
    this.tweens.add({
      targets: [
        player.sprite,
        player.nameText,
        player.statusIcon,
        player.voteCard,
      ],
      scale: 0,
      duration: 300,
      ease: "Back.easeIn",
      onComplete: () => {
        player.sprite.destroy();
        player.nameText.destroy();
        player.statusIcon?.destroy();
        player.voteCard?.destroy();
      },
    });

    this.players.delete(playerId);
  }

  updatePlayerVote(playerId: string, hasVoted: boolean) {
    console.log("TavernScene.updatePlayerVote called:", playerId, hasVoted);
    const player = this.players.get(playerId);
    if (!player) {
      console.warn("Player not found for vote update:", playerId);
      return;
    }

    // Update status icon
    if (player.statusIcon) {
      player.statusIcon.setTexture(hasVoted ? "voted-icon" : "waiting-icon");
    }

    if (hasVoted && !player.voteCard) {
      console.log("Creating vote card for player:", playerId);
      // Add hidden card
      player.voteCard = this.add.sprite(
        player.position.x,
        player.position.y - 60,
        "card-hidden",
      );

      // Ensure card is visible
      player.voteCard.setDisplaySize(30, 40);
      player.voteCard.setDepth(10); // Au-dessus des autres éléments

      console.log(
        "Vote card created:",
        player.voteCard,
        "texture exists:",
        this.textures.exists("card-hidden"),
      );

      // Card animation
      player.voteCard.setScale(0);
      this.tweens.add({
        targets: player.voteCard,
        scale: 1,
        duration: 300,
        ease: "Back.easeOut",
      });
    } else if (!hasVoted && player.voteCard) {
      // Remove card
      this.tweens.add({
        targets: player.voteCard,
        scale: 0,
        duration: 200,
        ease: "Back.easeIn",
        onComplete: () => {
          player.voteCard?.destroy();
          player.voteCard = null;
        },
      });
    }
  }

  revealAllVotes(votes: { [playerId: string]: string }) {
    console.log("TavernScene.revealAllVotes called:", votes);
    this.isVotesRevealed = true;

    this.players.forEach((player, playerId) => {
      const vote = votes[playerId];
      console.log(
        "Processing vote reveal for player:",
        playerId,
        "vote:",
        vote,
        "hasCard:",
        !!player.voteCard,
      );
      if (vote && player.voteCard) {
        // Card flip animation
        this.tweens.add({
          targets: player.voteCard,
          scaleX: 0,
          duration: 150,
          ease: "Power2",
          onComplete: () => {
            // Change texture to show the vote
            const textureKey = `card-${vote}`;
            console.log(
              "Setting card texture to:",
              textureKey,
              "exists:",
              this.textures.exists(textureKey),
            );
            player.voteCard?.setTexture(textureKey);

            // Add vote text above the card
            const voteText = this.add
              .text(player.position.x, player.position.y - 90, vote, {
                fontSize: "18px",
                color: "#FFD700",
                fontFamily: "serif",
                fontStyle: "bold",
                stroke: "#000000",
                strokeThickness: 2,
              })
              .setOrigin(0.5)
              .setDepth(11);

            // Card flip back
            this.tweens.add({
              targets: player.voteCard,
              scaleX: 1,
              duration: 150,
              ease: "Power2",
            });
          },
        });
      }
    });

    // Global reveal visual effect
    const revealEffect = this.add.graphics();
    revealEffect.fillStyle(0xffd700, 0.3);
    revealEffect.fillCircle(400, 300, 400);

    this.tweens.add({
      targets: revealEffect,
      alpha: 0,
      duration: 1000,
      onComplete: () => revealEffect.destroy(),
    });
  }

  resetTable() {
    this.isVotesRevealed = false;

    this.players.forEach((player) => {
      // Remove cards and set back to waiting
      if (player.voteCard) {
        player.voteCard.destroy();
        player.voteCard = null;
      }

      if (player.statusIcon) {
        player.statusIcon.setTexture("waiting-icon");
      }

      // Remove vote texts
      // (Vote texts are not stored, they will be automatically destroyed on next resetTable)
    });

    // Clean up all remaining vote texts
    this.children.list
      .filter(
        (child) =>
          child instanceof Phaser.GameObjects.Text &&
          (child as Phaser.GameObjects.Text).text.match(/^[0-9?☕]+$/),
      )
      .forEach((child) => child.destroy());
  }

  update() {
    // Subtle animations to bring the scene to life

    // Make torches flicker
    this.children.list
      .filter(
        (child) =>
          child instanceof Phaser.GameObjects.Text &&
          (child as Phaser.GameObjects.Text).text === "🔥",
      )
      .forEach((torch, index) => {
        const time = this.time.now + index * 1000;
        const alpha = 0.7 + Math.sin(time * 0.005) * 0.3;
        (torch as Phaser.GameObjects.Text).setAlpha(alpha);
      });
  }
}
