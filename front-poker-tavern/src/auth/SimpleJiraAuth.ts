import { API_CONFIG } from '@/config/api';

class SimpleJiraAuth {
  private baseUrl: string;

  constructor(baseUrl?: string) {
    this.baseUrl = baseUrl || API_CONFIG.BASE_URL;
  }

  /**
   * Authenticates user with JIRA using popup OAuth flow
   * @param sessionId - The session ID to authenticate for
   * @returns Promise that resolves with user data
   */
  async authenticate(sessionId: string): Promise<any> {
    try {
      // Step 1: Get OAuth URL from backend
      const response = await fetch(`${this.baseUrl}/api/jira/auth-url?sessionId=${sessionId}`);
      
      if (!response.ok) {
        throw new Error(`Failed to get OAuth URL: ${response.statusText}`);
      }
      
      const { authUrl } = await response.json();
      
      // Step 2: Open popup with OAuth URL
      const popup = window.open(
        authUrl,
        'jira-oauth',
        'width=500,height=600,scrollbars=yes,resizable=yes'
      );
      
      if (!popup) {
        throw new Error('Failed to open authentication popup');
      }
      
      // Step 3: Listen for authentication success
      return new Promise((resolve, reject) => {
        const messageHandler = (event: MessageEvent) => {
          // For development, be more flexible with origin validation
          // Accept messages from the backend or if using wildcard
          const isValidOrigin = event.origin === this.baseUrl || 
                               event.origin.includes('onrender.com') ||
                               this.baseUrl.includes('localhost');
          
          if (!isValidOrigin) {
            console.log('Ignoring message from origin:', event.origin, 'Expected:', this.baseUrl);
            return;
          }
          
          console.log('Received valid postMessage from:', event.origin, 'Data:', event.data);
          
          if (event.data.type === 'JIRA_AUTH_SUCCESS') {
            // Clean up
            window.removeEventListener('message', messageHandler);
            popup.close();
            
            // Resolve with user data
            resolve(event.data.user);
          } else if (event.data.type === 'JIRA_AUTH_ERROR') {
            // Clean up
            window.removeEventListener('message', messageHandler);
            popup.close();
            
            // Reject with error
            reject(new Error(event.data.error || 'Authentication failed'));
          }
        };
        
        // Listen for messages from popup
        window.addEventListener('message', messageHandler);
        
        // Handle popup being closed manually
        const checkClosed = setInterval(() => {
          if (popup.closed) {
            clearInterval(checkClosed);
            window.removeEventListener('message', messageHandler);
            reject(new Error('Authentication popup was closed'));
          }
        }, 1000);
        
        // Timeout after 5 minutes
        setTimeout(() => {
          clearInterval(checkClosed);
          window.removeEventListener('message', messageHandler);
          if (!popup.closed) {
            popup.close();
          }
          reject(new Error('Authentication timeout'));
        }, 5 * 60 * 1000);
      });
      
    } catch (error) {
      throw new Error(`Authentication failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Check if the session is already authenticated with JIRA
   * @param sessionId - The session ID (optional for API token mode)
   * @returns Promise that resolves with connection status
   */
  async isAuthenticated(sessionId?: string): Promise<boolean> {
    try {
      const url = sessionId 
        ? `${this.baseUrl}/jira/status?sessionId=${sessionId}`
        : `${this.baseUrl}/jira/status`;
      const response = await fetch(url);
      
      if (!response.ok) {
        return false;
      }
      
      const { connected } = await response.json();
      return connected;
    } catch (error) {
      console.error('Error checking authentication status:', error);
      return false;
    }
  }

  /**
   * Search for JIRA issues
   * @param sessionId - The session ID (optional for API token mode)
   * @param query - JQL query string
   * @returns Promise that resolves with array of issues
   */
  async searchIssues(sessionId: string | undefined, query: string): Promise<any[]> {
    try {
      const url = sessionId
        ? `${this.baseUrl}/jira/search?sessionId=${sessionId}&q=${encodeURIComponent(query)}`
        : `${this.baseUrl}/jira/search?q=${encodeURIComponent(query)}`;
      const response = await fetch(url);
      
      if (!response.ok) {
        throw new Error(`Failed to search issues: ${response.statusText}`);
      }
      
      const { issues } = await response.json();
      return issues || [];
    } catch (error) {
      throw new Error(`Issue search failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }

  /**
   * Updates Story Points on a given JIRA issue.
   * Note: Backend has a safety rail; it will refuse unless JIRA_WRITE_ENABLED=true.
   */
  async updateStoryPoints(sessionId: string | undefined, issueKey: string, storyPoints: number): Promise<any> {
    const trimmedKey = (issueKey || '').trim();
    if (!trimmedKey) throw new Error('issueKey is required');
    if (!Number.isFinite(storyPoints) || storyPoints < 0) throw new Error('storyPoints must be >= 0');

    const response = await fetch(`${this.baseUrl}/jira/story-points`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        sessionId,
        issueKey: trimmedKey,
        storyPoints: Math.floor(storyPoints),
      }),
    });

    // Backend returns a JSON body for both success and dry-run rejection.
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const msg =
        typeof data?.message === 'string'
          ? data.message
          : `Failed to update story points (${response.status})`;
      const err = new Error(msg);
      (err as any).details = data;
      throw err;
    }
    return data;
  }
}

export default SimpleJiraAuth;
