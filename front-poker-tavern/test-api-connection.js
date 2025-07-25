// Simple test to verify API connection
import { getApiUrl } from './src/config/api.js';

console.log('Testing API connection...');
console.log('API Base URL:', import.meta.env.VITE_API_BASE_URL);
console.log('Health endpoint URL:', getApiUrl('/health'));

try {
  const response = await fetch(getApiUrl('/health'));
  const data = await response.json();
  console.log('✅ API connection successful!');
  console.log('Response:', data);
} catch (error) {
  console.error('❌ API connection failed:', error);
}
