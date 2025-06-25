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
  persistentVotes: { [storyId: string]: { [playerId: string]: string } };
  requiredPlayers: string[]; // List of player IDs who need to vote
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
  const localPlayerId = ref<string | null>(null);
  const localStoryIndex = ref(0); // Independent story navigation

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
    return currentSession.value.stories[localStoryIndex.value] || null;
  });

  const allPlayersVotedCurrentStory = computed(() => {
    if (!currentSession.value || !currentStory.value) return false;

    const currentStoryVotes =
      currentSession.value.persistentVotes[currentStory.value.id] || {};
    const requiredPlayers = currentSession.value.requiredPlayers;

    return (
      requiredPlayers.length > 0 &&
      requiredPlayers.every((playerId) => currentStoryVotes[playerId])
    );
  });

  const allStoriesVotedByEveryone = computed(() => {
    if (!currentSession.value) return false;

    const stories = currentSession.value.stories;
    const requiredPlayers = currentSession.value.requiredPlayers;

    return stories.every((story) => {
      const storyVotes = currentSession.value!.persistentVotes[story.id] || {};
      return requiredPlayers.every((playerId) => storyVotes[playerId]);
    });
  });

  const isCurrentStoryRevealed = computed(() => {
    if (!currentStory.value || !currentSession.value) return false;
    return currentSession.value.revealVotes;
  });

  const canNavigateNext = computed(() => {
    if (!currentSession.value) return false;
    return localStoryIndex.value < currentSession.value.stories.length - 1;
  });

  const canNavigatePrev = computed(() => {
    return localStoryIndex.value > 0;
  });

  const currentStoryProgress = computed(() => {
    if (!currentSession.value) return { current: 0, total: 0 };
    return {
      current: localStoryIndex.value + 1,
      total: currentSession.value.stories.length,
    };
  });

  const canReveal = computed(() => {
    return gamePhase.value === "voting" && allStoriesVotedByEveryone.value;
  });

  const votingResults = computed(() => {
    if (
      gamePhase.value !== "revealed" ||
      !currentStory.value ||
      !currentSession.value
    )
      return [];

    const currentStoryVotes =
      currentSession.value.persistentVotes[currentStory.value.id] || {};
    const playerMap = new Map(players.value.map((p) => [p.id, p]));

    return Object.entries(currentStoryVotes).map(([playerId, vote]) => {
      const player = playerMap.get(playerId);
      return {
        player: player?.name || playerId,
        vote: vote,
        character: player?.character || "unknown",
      };
    });
  });

  const averageVote = computed(() => {
    if (!currentStory.value || !currentSession.value) return null;

    const currentStoryVotes =
      currentSession.value.persistentVotes[currentStory.value.id] || {};
    const numericVotes = Object.values(currentStoryVotes)
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
      persistentVotes: {},
      requiredPlayers: ["player-1", "player-2", "player-3"], // Simulated team
    };

    // Load persistent state from localStorage
    loadPersistedState();

    // Test players (no votes initially, will be loaded from persistence)
    players.value = [
      {
        id: "player-1",
        name: "Alice",
        character: "mage",
        emoji: "🧙‍♀️",
        hasVoted: false,
        position: { x: 400, y: 200 },
        isReady: true,
      },
      {
        id: "player-2",
        name: "Bob",
        character: "paladin",
        emoji: "⚔️",
        hasVoted: false,
        position: { x: 600, y: 300 },
        isReady: true,
      },
      {
        id: "player-3",
        name: "Charlie",
        character: "rogue",
        emoji: "🗡️",
        hasVoted: false,
        position: { x: 200, y: 400 },
        isReady: true,
      },
    ];

    // Update player voted status based on persistent votes
    updatePlayerVotedStatus();

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

    // Set initial story to first one
    localStoryIndex.value = 0;

    // Check if all stories are complete and should be revealed
    if (allStoriesVotedByEveryone.value && currentSession.value.revealVotes) {
      gamePhase.value = "revealed";
    }
  }

  function selectCharacter(characterId: string) {
    const character = availableCharacters.value.find(
      (char) => char.id === characterId,
    );
    if (!character) return;

    // Generate or retrieve persistent player ID
    if (!localPlayerId.value) {
      localPlayerId.value = `player-${Date.now()}`;
      localStorage.setItem("localPlayerId", localPlayerId.value);
    }

    currentPlayer.value = {
      id: localPlayerId.value,
      name: "You", // To be replaced with real name
      character: character.id,
      emoji: character.emoji,
      hasVoted: hasCurrentPlayerVoted(),
      position: { x: 400, y: 400 }, // Default position
      isReady: true,
    };
  }

  function selectCard(cardValue: string) {
    if (gamePhase.value !== "voting") return;
    selectedCard.value = cardValue;
  }

  function submitVote() {
    if (
      !selectedCard.value ||
      !currentPlayer.value ||
      !currentStory.value ||
      !currentSession.value
    )
      return;

    // Store vote persistently
    if (!currentSession.value.persistentVotes[currentStory.value.id]) {
      currentSession.value.persistentVotes[currentStory.value.id] = {};
    }

    currentSession.value.persistentVotes[currentStory.value.id][
      currentPlayer.value.id
    ] = selectedCard.value;

    // Update local state
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

    // Persist to localStorage
    savePersistedState();

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
      text: "All votes revealed for all stories!",
      type: "system",
    });
  }

  function navigateToStory(storyIndex: number) {
    if (
      !currentSession.value ||
      storyIndex < 0 ||
      storyIndex >= currentSession.value.stories.length
    ) {
      return;
    }

    localStoryIndex.value = storyIndex;

    // Reset local UI state
    selectedCard.value = null;
    gamePhase.value = "voting";

    // Load existing vote for this story if any
    if (currentPlayer.value && hasCurrentPlayerVoted()) {
      currentPlayer.value.hasVoted = true;
      const currentStoryVotes =
        currentSession.value.persistentVotes[currentStory.value?.id || ""] ||
        {};
      currentPlayer.value.vote = currentStoryVotes[currentPlayer.value.id];
      selectedCard.value = currentPlayer.value.vote || null;
    } else if (currentPlayer.value) {
      currentPlayer.value.hasVoted = false;
      currentPlayer.value.vote = undefined;
    }

    // Update all players' voted status for this story
    updatePlayerVotedStatus();

    // Check if all stories are complete and should be revealed
    if (allStoriesVotedByEveryone.value && currentSession.value.revealVotes) {
      gamePhase.value = "revealed";
    }

    addChatMessage({
      author: "System",
      text: `Navigated to: ${currentStory.value?.title}`,
      type: "system",
    });
  }

  function nextStory() {
    if (canNavigateNext.value) {
      navigateToStory(localStoryIndex.value + 1);
    }
  }

  function previousStory() {
    if (canNavigatePrev.value) {
      navigateToStory(localStoryIndex.value - 1);
    }
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

  // Persistence functions
  function savePersistedState() {
    if (!currentSession.value) return;

    const persistedData = {
      sessionId: currentSession.value.id,
      persistentVotes: currentSession.value.persistentVotes,
      localPlayerId: localPlayerId.value,
    };

    localStorage.setItem("planningPokerState", JSON.stringify(persistedData));
  }

  function loadPersistedState() {
    const savedData = localStorage.getItem("planningPokerState");
    if (savedData && currentSession.value) {
      try {
        const parsed = JSON.parse(savedData);
        if (parsed.sessionId === currentSession.value.id) {
          currentSession.value.persistentVotes = parsed.persistentVotes || {};
          localPlayerId.value = parsed.localPlayerId;
        }
      } catch (error) {
        console.warn("Failed to load persisted state:", error);
      }
    }

    // Load local player ID
    const savedPlayerId = localStorage.getItem("localPlayerId");
    if (savedPlayerId) {
      localPlayerId.value = savedPlayerId;
    }
  }

  function hasCurrentPlayerVoted(): boolean {
    if (!currentPlayer.value || !currentStory.value || !currentSession.value)
      return false;

    const currentStoryVotes =
      currentSession.value.persistentVotes[currentStory.value.id] || {};
    return !!currentStoryVotes[currentPlayer.value.id];
  }

  function updatePlayerVotedStatus() {
    if (!currentStory.value || !currentSession.value) return;

    const currentStoryVotes =
      currentSession.value.persistentVotes[currentStory.value.id] || {};

    players.value.forEach((player) => {
      player.hasVoted = !!currentStoryVotes[player.id];
      player.vote = currentStoryVotes[player.id];
    });
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
    localStoryIndex,

    // Computed
    currentStory,
    allPlayersVotedCurrentStory,
    allStoriesVotedByEveryone,
    canReveal,
    votingResults,
    averageVote,
    isCurrentStoryRevealed,
    canNavigateNext,
    canNavigatePrev,
    currentStoryProgress,

    // Actions
    initMockData,
    selectCharacter,
    selectCard,
    submitVote,
    revealVotes,
    navigateToStory,
    nextStory,
    previousStory,
    addChatMessage,
    sendChatMessage,
    resetGame,
    connectToSession,
    disconnectFromSession,
    savePersistedState,
    loadPersistedState,
    hasCurrentPlayerVoted,
    updatePlayerVotedStatus,
  };
});
