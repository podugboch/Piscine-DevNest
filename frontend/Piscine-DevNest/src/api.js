const API_URL = import.meta.env.VITE_API_URL; // Make sure your .env has VITE_API_URL=http://localhost:8080

// Ping backend
export const pingBackend = async () => {
  const res = await fetch(`${API_URL}/api/ping`);
  return res.json();
};

// Login function
export const login = async (email, password) => {
  const res = await fetch(`${API_URL}/api/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  return res.json();
};

// Get all users (requires JWT token)
export const getUsers = async (token) => {
  const res = await fetch(`${API_URL}/api/users`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  return res.json();
};

