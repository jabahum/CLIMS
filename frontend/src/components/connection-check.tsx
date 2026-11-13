"use client";

import { useState } from "react";
import { checkReadiness } from "@/lib/api";

export default function ConnectionCheck() {
  const [status, setStatus] = useState("Connection not checked");
  const [loading, setLoading] = useState(false);
  async function check() {
    setLoading(true);
    setStatus("Checking connection…");
    try {
      setStatus(await checkReadiness());
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Connection failed.");
    } finally {
      setLoading(false);
    }
  }
  return (
    <div className="connection">
      <div>
        <span className="eyebrow">Platform connection</span>
        <p role="status" aria-live="polite">
          {status}
        </p>
      </div>
      <button onClick={check} disabled={loading}>
        {loading ? "Checking…" : "Check connection"}
        <span aria-hidden="true"> ↗</span>
      </button>
    </div>
  );
}
