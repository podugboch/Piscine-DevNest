import { useEffect, useState } from "react";
import { getUsers } from "../api";

export default function Users({ token }) {
  const [users, setUsers] = useState([]);

  useEffect(() => {
    getUsers(token).then(setUsers);
  }, [token]);

  return (
    <div style={{ maxWidth: 600, margin: "50px auto" }}>
      <h2>Users</h2>
      <ul>
        {users.length > 0 ? (
          users.map((user) => (
            <li key={user.ID}>
              {user.username} ({user.email})
            </li>
          ))
        ) : (
          <li>No users found</li>
        )}
      </ul>
    </div>
  );
}
