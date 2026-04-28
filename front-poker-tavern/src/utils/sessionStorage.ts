import type { GamePhase } from "@/stores/gameStore";

export interface SessionStorageData {
  sessionId: string;
  sessionName: string;
  persistentVotes: { [storyId: string]: { [playerId: string]: string } };
  localPlayerId: string | null;
  currentPlayer: any;
  players: any[];
  stories: any[];
  requiredPlayers: string[];
  gamePhase: GamePhase;
  localStoryIndex: number;
  playerCharacters?: Record<string, { character: string; emoji: string }>;
}

export interface RoomInfo {
  id: string;
  name: string;
  storiesCount: number;
  playersCount: number;
  lastAccessed: Date;
}

/**
 * Check if there's any active session in localStorage
 */
export function hasActiveSession(): boolean {
  try {
    // Check generic key first
    const storedData = localStorage.getItem("planningPokerState");
    if (storedData) {
      const parsedData = JSON.parse(storedData);
      if (parsedData.sessionId) {
        return true;
      }
    }

    // Check for session-specific keys
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith("planningPokerState_")) {
        try {
          const sessionData = JSON.parse(localStorage.getItem(key) || "{}");
          if (sessionData.sessionId) {
            return true;
          }
        } catch (err) {
          console.warn(`Failed to parse session data for key ${key}:`, err);
        }
      }
    }

    return false;
  } catch (error) {
    console.warn("Failed to check for active session:", error);
    return false;
  }
}

/**
 * Get session data for a specific session ID
 */
export function getSessionData(sessionId: string): SessionStorageData | null {
  try {
    const sessionKey = `planningPokerState_${sessionId}`;
    const savedData = localStorage.getItem(sessionKey);

    if (savedData) {
      return JSON.parse(savedData);
    }

    return null;
  } catch (error) {
    console.warn(`Failed to get session data for ${sessionId}:`, error);
    return null;
  }
}

/**
 * Save session data for a specific session ID
 */
export function saveSessionData(
  sessionId: string,
  data: SessionStorageData,
): void {
  try {
    const sessionKey = `planningPokerState_${sessionId}`;
    localStorage.setItem(sessionKey, JSON.stringify(data));
  } catch (error) {
    console.error(`Failed to save session data for ${sessionId}:`, error);
  }
}

/**
 * Get all available rooms from localStorage
 */
export function getAvailableRooms(): RoomInfo[] {
  const rooms: RoomInfo[] = [];

  try {
    // Iterate through all localStorage keys to find room data
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith("planningPokerState_")) {
        try {
          const roomData = JSON.parse(localStorage.getItem(key) || "{}");
          if (roomData.sessionId && roomData.sessionName) {
            rooms.push({
              id: roomData.sessionId,
              name: roomData.sessionName,
              storiesCount: roomData.stories ? roomData.stories.length : 0,
              playersCount: roomData.players ? roomData.players.length : 0,
              lastAccessed: new Date(), // Could be enhanced to track actual last access
            });
          }
        } catch (error) {
          console.warn(`Failed to parse room data for key ${key}:`, error);
        }
      }
    }
  } catch (error) {
    console.warn("Failed to get available rooms:", error);
  }

  return rooms;
}

/**
 * Delete a specific room from localStorage
 */
export function deleteRoom(sessionId: string): void {
  try {
    const storageKey = `planningPokerState_${sessionId}`;
    localStorage.removeItem(storageKey);
    console.log(`Room ${sessionId} deleted from localStorage`);
  } catch (error) {
    console.error(`Failed to delete room ${sessionId}:`, error);
  }
}

/**
 * Clear all session data from localStorage
 */
export function clearAllSessionData(): void {
  try {
    // Remove generic key
    localStorage.removeItem("planningPokerState");

    // Remove all session-specific keys
    const keysToRemove: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith("planningPokerState_")) {
        keysToRemove.push(key);
      }
    }

    keysToRemove.forEach((key) => localStorage.removeItem(key));

    console.log("All session data cleared from localStorage");
  } catch (error) {
    console.error("Failed to clear all session data:", error);
  }
}

/**
 * Get the most recent session data (for auto-login)
 */
export function getMostRecentSession(): SessionStorageData | null {
  try {
    // First check generic key
    const storedData = localStorage.getItem("planningPokerState");
    if (storedData) {
      const parsedData = JSON.parse(storedData);
      if (parsedData.sessionId) {
        return parsedData;
      }
    }

    // Then check session-specific keys (return first valid one found)
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key && key.startsWith("planningPokerState_")) {
        try {
          const sessionData = JSON.parse(localStorage.getItem(key) || "{}");
          if (sessionData.sessionId) {
            return sessionData;
          }
        } catch (err) {
          console.warn(`Failed to parse session data for key ${key}:`, err);
        }
      }
    }

    return null;
  } catch (error) {
    console.warn("Failed to get most recent session:", error);
    return null;
  }
}
