import "./App.css";
import { useEffect, useState } from "react";

function App() {
  const [status, setStatus] = useState("loading...");

  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const res = await fetch("http://localhost:5000/health");
        if (!res.ok) {
          throw new Error(`Request failed: ${res.status}`);
        }

        const data = (await res.json()) as { status?: string };
        setStatus(data.status ?? "unknown");
      } catch {
        setStatus("error: cannot reach Go API");
      }
    };

    fetchHealth();
  }, []);

  return (
    <>
      <h1>Web + Go API</h1>
      <p>API health: {status}</p>
    </>
  );
}

export default App;
