export const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export async function checkReadiness(): Promise<string> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);
  try {
    const url = new URL(API_BASE_URL);
    url.pathname = "/health/ready";
    url.search = "";
    const response = await fetch(url.toString(), { signal: controller.signal });
    if (!response.ok) throw new Error("The API is not ready. Please try again.");
    const result = await response.json();
    if (!result.success || result.data?.database !== "connected") {
      throw new Error("The API returned an unexpected readiness response.");
    }
    return "API and database connected";
  } finally {
    clearTimeout(timeout);
  }
}
