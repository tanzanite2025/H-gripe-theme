const API_BASE_URL = 'http://localhost:9000/api/admin';

/**
 * Enhanced fetch wrapper with Fail Loudly principle
 */
export async function http(endpoint, options = {}) {
  // Determine full URL
  const isAbsolute = endpoint.startsWith('http');
  const isCustomBase = endpoint.startsWith('/api/'); 
  
  let url = endpoint;
  if (!isAbsolute) {
    if (isCustomBase) {
      url = `http://localhost:9000${endpoint}`;
    } else {
      const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
      url = `${API_BASE_URL}${cleanEndpoint}`;
    }
  }

  // Handle headers
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const fetchOptions = {
    ...options,
    headers,
  };

  try {
    const response = await fetch(url, fetchOptions);

    const text = await response.text();
    let data;
    try {
      data = text ? JSON.parse(text) : null;
    } catch (e) {
      data = text;
    }

    if (!response.ok) {
      const errorMessage = data?.error || data?.message || response.statusText || 'Unknown error';
      console.error(`[CRITICAL] API Request failed for ${url}: ${response.status} - ${errorMessage}`);
      throw new Error(errorMessage);
    }

    return data;
  } catch (error) {
    console.error(`[CRITICAL] Network or parsing error for ${url}:`, error);
    throw error;
  }
}

export default http;
