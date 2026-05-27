const BASE_URL = import.meta.env.VITE_API_URL;

function getToken() {
  return localStorage.getItem("token");
}

function handleUnauthorized() {
  localStorage.removeItem("token");
  localStorage.removeItem("user");
  window.location.href = "/login";
}

/**
 * Centralized fetch wrapper.
 * @param {string} path - API path e.g. "/api/products"
 * @param {RequestInit} options - fetch options
 * @param {boolean} auth - attach Bearer token (default true)
 * @returns {Promise<any>} parsed JSON data
 * @throws {Error} with .message normalized from backend response
 */
export async function apiRequest(path, options = {}, auth = true) {
  const headers = { "Content-Type": "application/json", ...options.headers };

  if (auth) {
    const token = getToken();
    if (token) headers["Authorization"] = `Bearer ${token}`;
  }

  const response = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  if (response.status === 401 && auth && getToken()) {
    // A token existed but was rejected — it expired or is invalid
    handleUnauthorized();
    throw new Error("Session expired. Please log in again.");
  }

  let data;
  try {
    data = await response.json();
  } catch {
    if (!response.ok) throw new Error("Unexpected server response.");
    return null;
  }

  if (!response.ok) {
    // New shape: { code, message }  |  old shape fallback: { error }
    throw new Error(data.message || data.error || "An unexpected error occurred.");
  }

  return data;
}
