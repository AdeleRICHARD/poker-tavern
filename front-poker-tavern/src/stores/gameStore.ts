import { defineStore } from "pinia";
import { ref, computed } from "vue";
import {
  saveSessionData,
  getSessionData,
  getMostRecentSession,
  getAvailableRooms as getAvailableRoomsUtil,
  deleteRoom as deleteRoomUtil,
  clearAllSessionData,
} from "@/utils/sessionStorage";

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

export enum GamePhase {
  VOTING = "voting",
  REVEALED = "revealed",
  END = "end",
  WAITING = "waiting",
  START = "start",
  IDLE = "idle",
  LOBBY = "lobby",
}

export const useGameStore = defineStore("game", () => {
  // Reactive state
  const currentPlayer = ref<Player | null>(null);
  const players = ref<Player[]>([]);
  const currentSession = ref<GameSession | null>(null);
  const chatMessages = ref<ChatMessage[]>([]);
  const selectedCard = ref<string | null>(null);
  const gamePhase = ref<GamePhase>(GamePhase.WAITING);
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
  function initializeStore() {
    // Initialize empty state - will be populated by backend
    currentSession.value = null;
    players.value = [];
    chatMessages.value = [];
    gamePhase.value = GamePhase.WAITING;
    localStoryIndex.value = 0;

    // Load any persisted local state
    loadPersistedState();
  }

  function selectCharacter(characterId: string) {
    const character = availableCharacters.value.find(
      (char) => char.id === characterId,
    );
    if (!character) return;

    // Generate or retrieve persistent player ID only if not already set
    if (!localPlayerId.value) {
      localPlayerId.value = `player-${Date.now()}`;
      localStorage.setItem("localPlayerId", localPlayerId.value);
    }

    // Update existing player or create new one
    if (currentPlayer.value) {
      // Update existing player's character
      currentPlayer.value.character = character.id;
      currentPlayer.value.emoji = character.emoji;
      currentPlayer.value.hasVoted = hasCurrentPlayerVoted();

      // Update the player in the players array
      const existingPlayerIndex = players.value.findIndex(
        (p) => p.id === currentPlayer.value!.id,
      );
      if (existingPlayerIndex >= 0) {
        players.value[existingPlayerIndex] = { ...currentPlayer.value };
      }
    } else {
      // Create new player
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

    // Save state after character selection
    savePersistedState();
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

    gamePhase.value = GamePhase.REVEALED;
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
    gamePhase.value = GamePhase.VOTING;

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
      gamePhase.value = GamePhase.REVEALED;
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
    gamePhase.value = GamePhase.WAITING;
    currentPlayer.value = null;
  }

  // Persistence functions
  function savePersistedState() {
    if (!currentSession.value) return;

    const persistedData = {
      sessionId: currentSession.value.id,
      sessionName: currentSession.value.name,
      persistentVotes: currentSession.value.persistentVotes,
      localPlayerId: localPlayerId.value,
      currentPlayer: currentPlayer.value,
      players: players.value,
      stories: currentSession.value.stories,
      requiredPlayers: currentSession.value.requiredPlayers,
      gamePhase: gamePhase.value,
      localStoryIndex: localStoryIndex.value,
    };

    // Use utility function to save session data
    saveSessionData(currentSession.value.id, persistedData);
  }

  function loadPersistedState(sessionId?: string) {
    let parsed = null;

    // If sessionId is provided, try to load session-specific data first
    if (sessionId) {
      parsed = getSessionData(sessionId);
    }

    // Fall back to most recent session if no specific session requested
    if (!parsed) {
      parsed = getMostRecentSession();
    }

    if (parsed && parsed.sessionId && parsed.sessionName) {
      // Restore session
      currentSession.value = {
        id: parsed.sessionId,
        name: parsed.sessionName,
        createdAt: new Date(),
        isActive: true,
        currentStoryIndex: 0,
        stories: parsed.stories || [],
        revealVotes: false,
        persistentVotes: parsed.persistentVotes || {},
        requiredPlayers: parsed.requiredPlayers || [],
      };

      // Restore local state
      localPlayerId.value = parsed.localPlayerId;
      currentPlayer.value = parsed.currentPlayer;
      players.value = parsed.players || [];
      gamePhase.value = parsed.gamePhase || GamePhase.WAITING;
      localStoryIndex.value = parsed.localStoryIndex || 0;
      isConnected.value = true;

      console.log(
        "Restored session state from localStorage:",
        parsed.sessionId,
      );
    }

    // Load local player ID if not already loaded
    if (!localPlayerId.value) {
      const savedPlayerId = localStorage.getItem("localPlayerId");
      if (savedPlayerId) {
        localPlayerId.value = savedPlayerId;
      }
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

  // WebSocket connection - to be implemented with Go backend
  function connectToSession(sessionId: string) {
    isConnected.value = true;

    // Only create mock session if no session exists yet OR if the session ID doesn't match
    // This should only happen when joining an existing session, not when creating a new one
    if (!currentSession.value || currentSession.value.id !== sessionId) {
      currentSession.value = {
        id: sessionId,
        name: `Session ${sessionId}`,
        createdAt: new Date(),
        isActive: true,
        currentStoryIndex: 0,
        stories: [],
        revealVotes: false,
        persistentVotes: {},
        requiredPlayers: [],
      };
    }

    gamePhase.value = GamePhase.WAITING;

    addChatMessage({
      author: "System",
      text: `Connected to session ${currentSession.value?.name || sessionId}`,
      type: "system",
    });

    // Save session state
    savePersistedState();

    // TODO: Implement WebSocket connection to Go backend
    console.log("TODO: Connect to backend session:", sessionId);
  }

  function disconnectFromSession() {
    isConnected.value = false;

    // Only clear local connection state, preserve room data for reconnection
    // The session data stays in localStorage so user can rejoin later

    addChatMessage({
      author: "System",
      text: "Disconnected from session (room data preserved)",
      type: "system",
    });

    // TODO: Close WebSocket connection
    console.log("TODO: Disconnect from backend");
  }

  // New functions for backend integration
  async function joinSession(sessionId: string, playerName: string) {
    console.log("Joining session", sessionId, "as", playerName);

    try {
      // First, join the session via the backend
      const joinResponse = await fetch("http://localhost:8080/session/join", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          sessionId: sessionId,
          playerName: playerName,
        }),
      });

      if (!joinResponse.ok) {
        throw new Error(`Failed to join session: ${joinResponse.status}`);
      }

      const sessionData = await joinResponse.json();
      console.log("Joined session data:", sessionData);

      // Create local session object with backend data
      currentSession.value = {
        id: sessionData.sessionId,
        name: sessionData.name,
        createdAt: new Date(sessionData.createdAt || Date.now()),
        isActive: true,
        currentStoryIndex: 0,
        stories: sessionData.stories || [],
        revealVotes: false,
        persistentVotes: {},
        requiredPlayers: sessionData.players || [],
      };

      // Convert backend player strings to frontend player objects
      if (sessionData.players) {
        players.value = sessionData.players.map((playerName: string) => ({
          id: playerName, // Using name as ID for now
          name: playerName,
          character: "mage",
          emoji: "🧙‍♂️",
          hasVoted: false,
          position: { x: 400, y: 400 },
          isReady: true,
        }));
      }

      return Promise.resolve();
    } catch (error) {
      console.error("Error joining session:", error);
      throw error;
    }
  }

  async function createSession(sessionName: string, stories: Story[]) {
    console.log("Creating session", sessionName, "with stories", stories);

    try {
      // Call the backend API to create session
      const response = await fetch("http://localhost:8080/session", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          name: sessionName,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to create session: ${response.status}`);
      }

      const sessionData = await response.json();
      console.log("Backend response:", sessionData);

      // Create local session object with backend data
      currentSession.value = {
        id: sessionData.sessionId,
        name: sessionData.name,
        createdAt: new Date(),
        isActive: true,
        currentStoryIndex: 0,
        stories: stories,
        revealVotes: false,
        persistentVotes: {},
        requiredPlayers: [],
      };

      // Save session state
      savePersistedState();

      return sessionData.sessionId;
    } catch (error) {
      console.error("Error creating session:", error);
      throw error;
    }
  }

  function authenticatePlayer(credentials: any) {
    // TODO: Authenticate with backend
    console.log("TODO: Authenticate player", credentials);
  }

  function logout() {
    disconnectFromSession();
    currentPlayer.value = null;
    localPlayerId.value = null;
    localStorage.removeItem("localPlayerId");
    localStorage.removeItem("planningPokerState");

    addChatMessage({
      author: "System",
      text: "Logged out successfully",
      type: "system",
    });
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
    initializeStore,
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
    joinSession,
    createSession,
    authenticatePlayer,
    logout,
    savePersistedState,
    loadPersistedState,
    hasCurrentPlayerVoted,
    updatePlayerVotedStatus,
    getAvailableRoomsUtil,
    deleteRoomUtil,
  };
});
