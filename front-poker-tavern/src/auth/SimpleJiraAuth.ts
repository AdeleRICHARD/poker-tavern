class SimpleJiraAuth {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
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
          // Verify origin for security
          if (event.origin !== 'http://localhost:8080') {
            return;
          }
          
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
   * @param sessionId - The session ID to check
   * @returns Promise that resolves with connection status
   */
  async isAuthenticated(sessionId: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/jira/status?sessionId=${sessionId}`);
      
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
   * @param sessionId - The session ID
   * @param query - JQL query string
   * @returns Promise that resolves with array of issues
   */
  async searchIssues(sessionId: string, query: string): Promise<any[]> {
    try {
      const response = await fetch(`${this.baseUrl}/jira/search?sessionId=${sessionId}&q=${encodeURIComponent(query)}`);
      
      if (!response.ok) {
        throw new Error(`Failed to search issues: ${response.statusText}`);
      }
      
      const { issues } = await response.json();
      return issues || [];
    } catch (error) {
      throw new Error(`Issue search failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
}

export default SimpleJiraAuth;
