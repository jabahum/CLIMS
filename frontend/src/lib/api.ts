export const API_BASE_URL = "/api/v1";

export async function checkReadiness(): Promise<string> {
  const response = await fetch("/health/ready", {
    cache: "no-store",
    signal: AbortSignal.timeout(8000),
  });
  if (!response.ok) throw new Error("The API is not ready. Please try again.");
  const result = await response.json();
  if (!result.success || result.data?.database !== "connected") {
    throw new Error("The API returned an unexpected readiness response.");
  }
  return "API and database connected";
}
