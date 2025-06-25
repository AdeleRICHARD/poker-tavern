import { defineStore } from "pinia";
import { ref, computed } from "vue";

export interface Player {
  id: string;
  name: string;
  character: string;
  emoji: string;
  hasVoted: boolean;
  vote?: string;
  position: { x: number; y: number };
  isReady: boolean;
}

export interface Story {
  id: string;
  title: string;
  description: string;
  jiraKey?: string;
  estimatedPoints?: number;
}

export interface ChatMessage {
  id: string;
  author: string;
  text: string;
  timestamp: Date;
  type: "message" | "system";
}

export interface GameSession {
  id: string;
  name: string;
  createdAt: Date;
  isActive: boolean;
  currentStoryIndex: number;
  stories: Story[];
  revealVotes: boolean;
}

export const useGameStore = defineStore("game", () => {
  // Reactive state
  const currentPlayer = ref<Player | null>(null);
  const players = ref<Player[]>([]);
  const currentSession = ref<GameSession | null>(null);
  const chatMessages = ref<ChatMessage[]>([]);
  const selectedCard = ref<string | null>(null);
  const gamePhase = ref<"waiting" | "voting" | "discussion" | "revealed">(
    "waiting",
  );
  const isConnected = ref(false);

  // Mock data for testing
  const availableCharacters = ref([
    { id: "mage", name: "Mage", emoji: "🧙‍♂️", class: "mage" },
    { id: "paladin", name: "Paladin", emoji: "⚔️", class: "paladin" },
    { id: "rogue", name: "Rogue", emoji: "🗡️", class: "rogue" },
    { id: "priest", name: "Priest", emoji: "✨", class: "priest" },
    { id: "warrior", name: "Warrior", emoji: "🛡️", class: "warrior" },
    { id: "hunter", name: "Hunter", emoji: "🏹", class: "hunter" },
    { id: "warlock", name: "Warlock", emoji: "😈", class: "warlock" },
    { id: "druid", name: "Druid", emoji: "🌿", class: "druid" },
  ]);

  const pokerCards = ref([
    { value: "0", label: "0", description: "Nothing to do" },
    { value: "1", label: "1", description: "Very simple" },
    { value: "2", label: "2", description: "Simple" },
    { value: "3", label: "3", description: "Medium" },
    { value: "5", label: "5", description: "Complex" },
    { value: "8", label: "8", description: "Very complex" },
    { value: "13", label: "13", description: "Huge" },
    { value: "21", label: "21", description: "Too big" },
    { value: "?", label: "?", description: "No idea" },
    { value: "☕", label: "☕", description: "Coffee break" },
  ]);

  // Computed properties
  const currentStory = computed(() => {
    if (!currentSession.value || !currentSession.value.stories.length)
      return null;
    return (
      currentSession.value.stories[currentSession.value.currentStoryIndex] ||
      null
    );
  });

  const allPlayersVoted = computed(() => {
    return (
      players.value.length > 0 &&
      players.value.every((player) => player.hasVoted)
    );
  });

  const canReveal = computed(() => {
    return gamePhase.value === "voting" && allPlayersVoted.value;
  });

  const votingResults = computed(() => {
    if (gamePhase.value !== "revealed") return [];

    const votes = players.value
      .filter((player) => player.vote)
      .map((player) => ({
        player: player.name,
        vote: player.vote!,
        character: player.character,
      }));

    return votes;
  });

  const averageVote = computed(() => {
    const numericVotes = players.value
      .map((player) => player.vote)
      .filter((vote) => vote && !isNaN(Number(vote)))
      .map(Number);

    if (numericVotes.length === 0) return null;

    const sum = numericVotes.reduce((acc, vote) => acc + vote, 0);
    return Math.round((sum / numericVotes.length) * 10) / 10;
  });

  // Actions
  function initMockData() {
    // Test session
    currentSession.value = {
      id: "session-1",
      name: "Sprint 24 - Planning",
      createdAt: new Date(),
      isActive: true,
      currentStoryIndex: 0,
      stories: [
        {
          id: "story-1",
          title: "OAuth 2.0 Implementation",
          description: "Add OAuth authentication with Google and GitHub",
          jiraKey: "DEV-123",
        },
        {
          id: "story-2",
          title: "Push Notification API",
          description: "Create real-time notification system",
          jiraKey: "DEV-124",
        },
        {
          id: "story-3",
          title: "Analytics Dashboard",
          description: "User metrics visualization interface",
          jiraKey: "DEV-125",
        },
      ],
      revealVotes: false,
    };

    // Test players (already voted to enable reveal testing)
    players.value = [
      {
        id: "player-1",
        name: "Alice",
        character: "mage",
        emoji: "🧙‍♀️",
        hasVoted: true,
        vote: "5",
        position: { x: 400, y: 200 },
        isReady: true,
      },
      {
        id: "player-2",
        name: "Bob",
        character: "paladin",
        emoji: "⚔️",
        hasVoted: true,
        vote: "8",
        position: { x: 600, y: 300 },
        isReady: true,
      },
    ];

    // Test chat messages
    chatMessages.value = [
      {
        id: "msg-1",
        author: "System",
        text: "Planning session started",
        timestamp: new Date(),
        type: "system",
      },
      {
        id: "msg-2",
        author: "Alice",
        text: "Hello everyone!",
        timestamp: new Date(),
        type: "message",
      },
    ];

    gamePhase.value = "voting";
  }

  function selectCharacter(characterId: string) {
    const character = availableCharacters.value.find(
      (char) => char.id === characterId,
    );
    if (!character) return;

    currentPlayer.value = {
      id: `player-${Date.now()}`,
      name: "You", // To be replaced with real name
      character: character.id,
      emoji: character.emoji,
      hasVoted: false,
      position: { x: 400, y: 400 }, // Default position
      isReady: true,
    };
  }

  function selectCard(cardValue: string) {
    if (gamePhase.value !== "voting") return;
    selectedCard.value = cardValue;
  }

  function submitVote() {
    if (!selectedCard.value || !currentPlayer.value) return;

    currentPlayer.value.vote = selectedCard.value;
    currentPlayer.value.hasVoted = true;

    // Add/update player in the list
    const existingPlayerIndex = players.value.findIndex(
      (p) => p.id === currentPlayer.value!.id,
    );
    if (existingPlayerIndex >= 0) {
      players.value[existingPlayerIndex] = { ...currentPlayer.value };
    } else {
      players.value.push({ ...currentPlayer.value });
    }

    addChatMessage({
      author: "System",
      text: `${currentPlayer.value.name} voted`,
      type: "system",
    });
  }

  function revealVotes() {
    if (!canReveal.value) return;

    gamePhase.value = "revealed";
    if (currentSession.value) {
      currentSession.value.revealVotes = true;
    }

    addChatMessage({
      author: "System",
      text: "Votes revealed!",
      type: "system",
    });
  }

  function startNextStory() {
    if (!currentSession.value) return;

    // Reset votes
    players.value.forEach((player) => {
      player.hasVoted = false;
      player.vote = undefined;
    });
    selectedCard.value = null;

    // Next story
    currentSession.value.currentStoryIndex++;
    currentSession.value.revealVotes = false;
    gamePhase.value = "voting";

    addChatMessage({
      author: "System",
      text: `New story: ${currentStory.value?.title}`,
      type: "system",
    });
  }

  function addChatMessage(message: Omit<ChatMessage, "id" | "timestamp">) {
    chatMessages.value.push({
      id: `msg-${Date.now()}`,
      timestamp: new Date(),
      ...message,
    });
  }

  function sendChatMessage(text: string) {
    if (!currentPlayer.value || !text.trim()) return;

    addChatMessage({
      author: currentPlayer.value.name,
      text: text.trim(),
      type: "message",
    });
  }

  function resetGame() {
    players.value = [];
    currentSession.value = null;
    chatMessages.value = [];
    selectedCard.value = null;
    gamePhase.value = "waiting";
    currentPlayer.value = null;
  }

  // WebSocket connection (mocked for now)
  function connectToSession(sessionId: string) {
    isConnected.value = true;
    initMockData();

    addChatMessage({
      author: "System",
      text: "Connected to session",
      type: "system",
    });
  }

  function disconnectFromSession() {
    isConnected.value = false;
    resetGame();
  }

  return {
    // State
    currentPlayer,
    players,
    currentSession,
    chatMessages,
    selectedCard,
    gamePhase,
    isConnected,
    availableCharacters,
    pokerCards,

    // Computed
    currentStory,
    allPlayersVoted,
    canReveal,
    votingResults,
    averageVote,

    // Actions
    initMockData,
    selectCharacter,
    selectCard,
    submitVote,
    revealVotes,
    startNextStory,
    addChatMessage,
    sendChatMessage,
    resetGame,
    connectToSession,
    disconnectFromSession,
  };
});
