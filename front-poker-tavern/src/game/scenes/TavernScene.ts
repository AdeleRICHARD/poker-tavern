import Phaser from "phaser";

interface PlayerSprite {
  id: string;
  sprite: Phaser.GameObjects.Sprite | Phaser.GameObjects.Container;
  nameText: Phaser.GameObjects.Text;
  voteCards: Phaser.GameObjects.Sprite[]; // Multiple cards for multiple stories
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
  }

  private createTableTexture() {
    const graphics = this.add.graphics();

    // Table shadow
    graphics.fillStyle(0x0a0705, 0.5);
    graphics.fillCircle(155, 155, 145);

    // Main table body - rich dark wood
    graphics.fillStyle(0x5a3d2b);
    graphics.fillCircle(150, 150, 140);

    // Inner wood ring - lighter
    graphics.fillStyle(0x6b4a35);
    graphics.fillCircle(150, 150, 120);

    // Center - even lighter
    graphics.fillStyle(0x7a5540);
    graphics.fillCircle(150, 150, 80);

    // Wood grain rings
    graphics.lineStyle(1, 0x4a3020, 0.4);
    graphics.strokeCircle(150, 150, 100);
    graphics.strokeCircle(150, 150, 60);
    graphics.strokeCircle(150, 150, 30);

    // Radial wood grain lines
    graphics.lineStyle(2, 0x3a2515, 0.3);
    for (let i = 0; i < 12; i++) {
      const angle = (i * Math.PI * 2) / 12;
      const x1 = 150 + Math.cos(angle) * 40;
      const y1 = 150 + Math.sin(angle) * 40;
      const x2 = 150 + Math.cos(angle) * 130;
      const y2 = 150 + Math.sin(angle) * 130;
      graphics.lineBetween(x1, y1, x2, y2);
    }

    // Thick golden border with bevel effect
    graphics.lineStyle(8, 0x8b6914);
    graphics.strokeCircle(150, 150, 140);
    graphics.lineStyle(4, 0xd4a756);
    graphics.strokeCircle(150, 150, 142);
    graphics.lineStyle(2, 0xffd700);
    graphics.strokeCircle(150, 150, 144);

    graphics.generateTexture("round-table", 300, 300);
    graphics.destroy();
  }

  private createBackgroundTexture() {
    const graphics = this.add.graphics();

    // Dark warm background - tavern wood/stone
    graphics.fillStyle(0x1a0f0a); // Dark brown base
    graphics.fillRect(0, 0, 800, 600);

    // Wooden floor planks
    graphics.fillStyle(0x2d1f15);
    for (let y = 0; y < 600; y += 40) {
      graphics.fillRect(0, y, 800, 38);
      graphics.lineStyle(2, 0x1a0f0a);
      graphics.lineBetween(0, y + 38, 800, y + 38);
    }

    // Wood grain lines
    graphics.lineStyle(1, 0x3d2a1a, 0.3);
    for (let x = 0; x < 800; x += 80) {
      for (let y = 0; y < 600; y += 40) {
        graphics.lineBetween(x + 10, y + 5, x + 70, y + 5);
        graphics.lineBetween(x + 15, y + 15, x + 60, y + 15);
        graphics.lineBetween(x + 5, y + 25, x + 75, y + 25);
      }
    }

    // Warm ambient light from center (where torches would be)
    const centerGradient = graphics.createGeometryMask();
    graphics.fillStyle(0x3d2a1a, 0.4);
    graphics.fillCircle(400, 300, 280);
    graphics.fillStyle(0x4a3525, 0.3);
    graphics.fillCircle(400, 300, 200);
    graphics.fillStyle(0x5a4530, 0.2);
    graphics.fillCircle(400, 300, 120);

    // Torch light spots on walls
    graphics.fillStyle(0x8b4513, 0.15);
    graphics.fillCircle(50, 300, 100); // Left torch
    graphics.fillCircle(750, 300, 100); // Right torch

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
      // Create a larger texture for better quality
      const size = 128;
      const center = size / 2;
      const radius = size / 2 - 4; // Leave room for shadow/glow

      const graphics = this.add.graphics();

      // 1. Drop shadow (simulated with semi-transparent black circle at offset)
      graphics.fillStyle(0x000000, 0.5);
      graphics.fillCircle(center + 4, center + 4, radius);

      // 2. Outer Metallic Rim (Gold/Silver/Bronze look)
      graphics.fillStyle(0xc0c0c0); // Silver base
      graphics.fillCircle(center, center, radius);
      
      // Gradient effect for rim (simulated with line styles)
      graphics.lineStyle(4, 0xffffff, 0.8); // Highlight top-left
      graphics.strokeCircle(center, center, radius - 2);
      
      // 3. Inner Class Color Ring
      graphics.fillStyle(playerClass.color);
      graphics.fillCircle(center, center, radius - 8);

      // 4. Dark Background for Symbol
      graphics.fillStyle(0x222222);
      graphics.fillCircle(center, center, radius - 14);

      // Generate the base token texture
      graphics.generateTexture(`player-base-${playerClass.name}`, size, size);
      graphics.destroy();
    });
  }

  private createCardTextures() {
    console.log("Creating card textures...");

    // Hidden card - brown back with gold border
    const hiddenGraphics = this.add.graphics();
    hiddenGraphics.fillStyle(0x8b4513);
    hiddenGraphics.fillRoundedRect(0, 0, 40, 55, 5);
    hiddenGraphics.lineStyle(3, 0xdaa520);
    hiddenGraphics.strokeRoundedRect(0, 0, 40, 55, 5);
    // Add decorative pattern
    hiddenGraphics.lineStyle(1, 0xdaa520, 0.5);
    hiddenGraphics.strokeRoundedRect(5, 5, 30, 45, 3);
    hiddenGraphics.generateTexture("card-hidden", 40, 55);
    hiddenGraphics.destroy();
    console.log("Created card-hidden texture");

    // Revealed cards for each value - we'll draw text separately during reveal
    // Just create a white card background texture
    const cardBgGraphics = this.add.graphics();
    cardBgGraphics.fillStyle(0xffffff);
    cardBgGraphics.fillRoundedRect(0, 0, 40, 55, 5);
    cardBgGraphics.lineStyle(2, 0x333333);
    cardBgGraphics.strokeRoundedRect(0, 0, 40, 55, 5);
    cardBgGraphics.generateTexture("card-revealed-bg", 40, 55);
    cardBgGraphics.destroy();
    console.log("Created card-revealed-bg texture");
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
    // Candles on the table center
    const candle1 = this.add.graphics();
    candle1.fillStyle(0xfff8dc);
    candle1.fillRect(395, 290, 4, 12);
    candle1.fillStyle(0xff6600);
    candle1.fillCircle(397, 288, 4);

    const candle2 = this.add.graphics();
    candle2.fillStyle(0xfff8dc);
    candle2.fillRect(405, 295, 4, 10);
    candle2.fillStyle(0xff6600);
    candle2.fillCircle(407, 293, 3);

    // Torch holders on walls (graphical)
    this.createTorch(50, 280);
    this.createTorch(750, 280);

    // Ambient glow from torches
    const leftGlow = this.add.graphics();
    leftGlow.fillStyle(0xff6600, 0.08);
    leftGlow.fillCircle(50, 300, 80);
    
    const rightGlow = this.add.graphics();
    rightGlow.fillStyle(0xff6600, 0.08);
    rightGlow.fillCircle(750, 300, 80);

    // Title banner
    this.add
      .text(400, 35, "⚔️ PLANNING POKER TAVERN ⚔️", {
        fontSize: "22px",
        color: "#d4a756",
        fontFamily: "serif",
        fontStyle: "bold",
        stroke: "#1a0f0a",
        strokeThickness: 4,
      })
      .setOrigin(0.5);
  }

  private createTorch(x: number, y: number) {
    // Torch holder
    const holder = this.add.graphics();
    holder.fillStyle(0x3a2515);
    holder.fillRect(x - 4, y, 8, 30);
    holder.fillStyle(0x2a1810);
    holder.fillRect(x - 6, y + 25, 12, 8);
    
    // Flame (animated separately in update)
    const flame = this.add
      .text(x, y - 5, "🔥", { fontSize: "24px" })
      .setOrigin(0.5)
      .setName(`torch-${x}`);
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

    // If player already exists, preserve their vote cards
    let existingVoteCards: Phaser.GameObjects.Sprite[] = [];
    let existingVote: string | undefined = undefined;

    if (this.players.has(playerId)) {
      console.log("Player already exists, preserving vote state:", playerId);
      const existingPlayer = this.players.get(playerId);
      if (existingPlayer) {
        existingVoteCards = existingPlayer.voteCards;
        existingVote = existingPlayer.data.vote;
        // Only remove visual elements without the cards
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

    // Create player container (Token + Symbol)
    const playerSprite = this.add.container(freePosition.x, freePosition.y);
    
    // Base Token Sprite
    const baseSprite = this.add.sprite(0, 0, `player-base-${playerData.character}`);
    baseSprite.setDisplaySize(70, 70);
    
    // Class Symbol (Emoji)
    const classes = [
      { name: "mage", symbol: "🧙" },
      { name: "paladin", symbol: "⚔️" },
      { name: "rogue", symbol: "🗡️" },
      { name: "priest", symbol: "✨" },
      { name: "warrior", symbol: "🛡️" },
      { name: "hunter", symbol: "🏹" },
      { name: "warlock", symbol: "😈" },
      { name: "druid", symbol: "🌿" },
    ];
    const playerClass = classes.find(c => c.name === playerData.character);
    const symbol = playerClass ? playerClass.symbol : "?";

    const symbolText = this.add.text(0, 0, symbol, { 
      fontSize: "36px",
      align: "center"
    }).setOrigin(0.5);

    // Add components to container
    playerSprite.add([baseSprite, symbolText]);

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
      voteCards: existingVoteCards, // Preserve existing cards
      statusIcon,
      position: freePosition,
      data: playerData,
    };

    // If we had existing cards, reposition them
    if (existingVoteCards.length > 0) {
      existingVoteCards.forEach((card, index) => {
        const centerX = 400;
        const centerY = 300;
        const cardDistance = 0.55;
        const offset = index * 12; // Stack offset
        const cardX = freePosition.x + (centerX - freePosition.x) * cardDistance + offset;
        const cardY = freePosition.y + (centerY - freePosition.y) * cardDistance - (index * 5);
        card.setPosition(cardX, cardY);
      });
      console.log("Repositioned existing vote cards for player:", playerId);
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
        ...player.voteCards,
      ],
      scale: 0,
      duration: 300,
      ease: "Back.easeIn",
      onComplete: () => {
        player.sprite.destroy();
        player.nameText.destroy();
        player.statusIcon?.destroy();
        player.voteCards.forEach(card => card.destroy());
      },
    });

    this.players.delete(playerId);
  }

  updatePlayerVote(playerId: string, hasVoted: boolean, voteCount: number = 1) {
    console.log("TavernScene.updatePlayerVote called:", playerId, hasVoted, "voteCount:", voteCount);
    const player = this.players.get(playerId);
    if (!player) {
      console.warn("Player not found for vote update:", playerId);
      return;
    }

    // Update status icon
    if (player.statusIcon) {
      player.statusIcon.setTexture(hasVoted ? "voted-icon" : "waiting-icon");
    }

    // Only add cards if we need more than we have
    if (hasVoted && player.voteCards.length < voteCount) {
      const cardsToAdd = voteCount - player.voteCards.length;
      console.log("Adding", cardsToAdd, "vote card(s) for player:", playerId);
      
      for (let i = 0; i < cardsToAdd; i++) {
        const cardIndex = player.voteCards.length;
        const centerX = 400;
        const centerY = 300;
        const cardDistance = 0.55;
        
        // Calculate base position towards center
        const baseCardX = player.position.x + (centerX - player.position.x) * cardDistance;
        const baseCardY = player.position.y + (centerY - player.position.y) * cardDistance;
        
        // Stack cards with slight offset
        const offsetX = cardIndex * 10;
        const offsetY = -cardIndex * 4;
        const cardX = baseCardX + offsetX;
        const cardY = baseCardY + offsetY;
        
        // Create new card
        const newCard = this.add.sprite(cardX, cardY, "card-hidden");
        newCard.setDisplaySize(40, 55);
        newCard.setDepth(10 + cardIndex);
        
        // Animation - slide from player to table
        newCard.setPosition(player.position.x, player.position.y);
        newCard.setScale(0);
        this.tweens.add({
          targets: newCard,
          x: cardX,
          y: cardY,
          scale: 1,
          duration: 400,
          ease: "Back.easeOut",
        });
        
        player.voteCards.push(newCard);
      }
      console.log("Vote cards updated, total cards:", player.voteCards.length);
    }
  }

  revealAllVotes(persistentVotes: { [storyId: string]: { [playerId: string]: string } }) {
    console.log("TavernScene.revealAllVotes called:", persistentVotes);
    this.isVotesRevealed = true;

    // Extract all vote values per player (list of votes for each story they voted on)
    const playerVotesMap = new Map<string, string[]>();
    Object.values(persistentVotes).forEach(storyVotes => {
      Object.entries(storyVotes).forEach(([playerId, vote]) => {
        if (!playerVotesMap.has(playerId)) {
          playerVotesMap.set(playerId, []);
        }
        playerVotesMap.get(playerId)!.push(vote);
      });
    });

    // Create a dark overlay for dramatic effect
    const overlay = this.add.graphics();
    overlay.fillStyle(0x000000, 0.7);
    overlay.fillRect(0, 0, 800, 600);
    overlay.setDepth(50);
    overlay.setAlpha(0);

    // Fade in overlay
    this.tweens.add({
      targets: overlay,
      alpha: 1,
      duration: 300,
      onComplete: () => {
        // Position for reveal display
        let yPosition = 80;
        const cardSpacing = 50;
        const playerSpacing = 90;

        // Convert players map to array and reveal each
        const playersArray = Array.from(this.players.entries());
        
        playersArray.forEach(([playerId, player], pIndex) => {
          // Get all votes for this player (one per story)
          const playerVotes = playerVotesMap.get(playerId) || [];
          
          // Create player name header
          const nameY = yPosition + pIndex * playerSpacing;
          const playerHeader = this.add
            .text(100, nameY, `${player.data.name}:`, {
              fontSize: "20px",
              color: "#FFD700",
              fontFamily: "serif",
              fontStyle: "bold",
              stroke: "#000000",
              strokeThickness: 3,
            })
            .setOrigin(0, 0.5)
            .setDepth(100)
            .setAlpha(0);

          // Animate player name in
          this.tweens.add({
            targets: playerHeader,
            alpha: 1,
            x: 120,
            duration: 300,
            delay: pIndex * 200,
          });

          // Create cards for this player (all their vote cards)
          player.voteCards.forEach((card, cardIndex) => {
            const cardX = 250 + cardIndex * cardSpacing;
            const cardY = nameY;
            
            // Get the vote value for this specific card (by index)
            const voteValue = playerVotes[cardIndex] || "?";
            
            // Move card to reveal position with animation
            this.tweens.add({
              targets: card,
              x: cardX,
              y: cardY,
              scale: 1.2,
              depth: 100 + cardIndex,
              duration: 400,
              delay: pIndex * 200 + cardIndex * 100,
              ease: "Back.easeOut",
              onComplete: () => {
                // Flip card
                this.tweens.add({
                  targets: card,
                  scaleX: 0,
                  duration: 100,
                  onComplete: () => {
                    // Change to revealed texture
                    card.setTexture("card-revealed-bg");
                    
                    // Add vote value text on the card (using voteValue from parent scope)
                    const voteLabel = this.add
                      .text(cardX, cardY, voteValue, {
                        fontSize: "22px",
                        color: "#333333",
                        fontFamily: "Arial",
                        fontStyle: "bold",
                      })
                      .setOrigin(0.5)
                      .setDepth(101 + cardIndex);

                    // Flip back
                    this.tweens.add({
                      targets: card,
                      scaleX: 1.2,
                      duration: 100,
                    });
                  },
                });
              },
            });
          });
        });

        // Add "Tap to close" text
        const closeText = this.add
          .text(400, 550, "🎉 Votes Revealed! Click anywhere to continue...", {
            fontSize: "18px",
            color: "#FFFFFF",
            fontFamily: "serif",
          })
          .setOrigin(0.5)
          .setDepth(100)
          .setAlpha(0);

        this.tweens.add({
          targets: closeText,
          alpha: 1,
          duration: 500,
          delay: playersArray.length * 200 + 500,
        });

        // Allow click to dismiss
        this.input.once("pointerdown", () => {
          // Fade out overlay and clean up
          this.tweens.add({
            targets: [overlay, closeText],
            alpha: 0,
            duration: 300,
            onComplete: () => {
              overlay.destroy();
              closeText.destroy();
              // Return cards to original positions
              this.players.forEach((player) => {
                player.voteCards.forEach((card, idx) => {
                  const centerX = 400;
                  const centerY = 300;
                  const cardDistance = 0.55;
                  const baseCardX = player.position.x + (centerX - player.position.x) * cardDistance;
                  const baseCardY = player.position.y + (centerY - player.position.y) * cardDistance;
                  const offsetX = idx * 10;
                  const offsetY = -idx * 4;
                  
                  this.tweens.add({
                    targets: card,
                    x: baseCardX + offsetX,
                    y: baseCardY + offsetY,
                    scale: 1,
                    depth: 10 + idx,
                    duration: 300,
                  });
                });
              });
            },
          });
        });
      },
    });
  }

  resetTable() {
    this.isVotesRevealed = false;

    this.players.forEach((player) => {
      // Remove all cards and set back to waiting
      player.voteCards.forEach(card => card.destroy());
      player.voteCards = [];

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
    const time = this.time.now;

    // Make torches flicker with more dynamic effect
    this.children.list
      .filter(
        (child) =>
          child instanceof Phaser.GameObjects.Text &&
          (child as Phaser.GameObjects.Text).text === "🔥",
      )
      .forEach((torch, index) => {
        const offset = index * 500;
        const alpha = 0.75 + Math.sin((time + offset) * 0.008) * 0.25;
        const scale = 1 + Math.sin((time + offset) * 0.01) * 0.1;
        const textObj = torch as Phaser.GameObjects.Text;
        textObj.setAlpha(alpha);
        textObj.setScale(scale);
      });
  }
}
