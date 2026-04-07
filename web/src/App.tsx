import "./App.css";
import { useState } from "react";
import type { FormEvent } from "react";

type HeadingCount = {
  level: string;
  count: number;
};

type AnalysisPayload = {
  htmlVersion: string;
  pageTitle: string;
  headings: HeadingCount[];
  internalLinks: number;
  externalLinks: number;
  inaccessibleLinks: number;
  hasLoginForm: boolean;
};

type AnalysisResult = {
  statusCode: number;
  message: string;
  body: AnalysisPayload;
};

type ErrorResponse = {
  statusCode: number;
  message: string;
};

function App() {
  const [url, setURL] = useState("");
  const [result, setResult] = useState<AnalysisPayload | null>(null);
  const [error, setError] = useState<ErrorResponse | null>(null);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const response = await fetch("http://localhost:5000/analyze", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url }),
      });

      if (!response.ok) {
        const apiError = (await response.json()) as ErrorResponse;
        setError({
          statusCode: apiError.statusCode ?? response.status,
          message: apiError.message ?? "Unknown error",
        });
        return;
      }

      const data = (await response.json()) as AnalysisResult;
      setResult(data.body);
    } catch {
      setError({
        statusCode: 0,
        message: "Cannot reach API server",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="container">
      <h1>Web Page Analyzer</h1>

      <form onSubmit={handleSubmit} className="analyze-form">
        <label htmlFor="url-input">URL</label>
        <input
          id="url-input"
          type="url"
          placeholder="https://example.com"
          value={url}
          onChange={(event) => setURL(event.target.value)}
          required
        />
        <button type="submit" disabled={loading}>
          {loading ? "Analyzing..." : "Analyze"}
        </button>
      </form>

      {error && (
        <section className="error-box">
          <h2>Error</h2>
          <p>
            HTTP {error.statusCode}: {error.message}
          </p>
        </section>
      )}

      {result && (
        <section className="result-box">
          <h2>Results</h2>
          <p>HTML version: {result.htmlVersion}</p>
          <p>Page title: {result.pageTitle}</p>
          <p>Internal links: {result.internalLinks}</p>
          <p>External links: {result.externalLinks}</p>
          <p>Inaccessible links: {result.inaccessibleLinks}</p>
          <p>Contains login form: {result.hasLoginForm ? "Yes" : "No"}</p>

          <h3>Headings</h3>
          <ul>
            {result.headings.map((heading) => (
              <li key={heading.level}>
                {heading.level}: {heading.count}
              </li>
            ))}
          </ul>
        </section>
      )}
    </main>
  );
}

export default App;
