import { parseRules } from "./ruleset-matcher.mjs";

export const DEFAULT_BASE_URL = new URL("./", import.meta.url);

const TEXT_BEHAVIORS = new Set(["domain", "ipcidr", "classical"]);

function requestError(url, error) {
  return new Error(
    `Fetch ${url}: ${error instanceof Error ? error.message : String(error)}`,
    { cause: error },
  );
}

async function fetchText(fetcher, url) {
  let response;
  try {
    response = await fetcher(url);
  } catch (error) {
    throw requestError(url, error);
  }
  if (!response || !response.ok) {
    const status = response
      ? `${response.status} ${response.statusText}`.trim()
      : "no response";
    throw new Error(`Fetch ${url}: HTTP ${status}`);
  }
  try {
    return await response.text();
  } catch (error) {
    throw requestError(url, error);
  }
}

function validateMeta(metadata) {
  if (
    !metadata ||
    typeof metadata !== "object" ||
    !Array.isArray(metadata.artifacts)
  ) {
    throw new TypeError("stats.json must contain an artifacts array.");
  }
  if (
    !Array.isArray(metadata.priority) ||
    !metadata.priority.length ||
    !metadata.priority.every((name) => typeof name === "string" && name) ||
    typeof metadata.fallback !== "string" ||
    !metadata.fallback
  ) {
    throw new TypeError(
      "stats.json must contain a ruleset priority and fallback.",
    );
  }
  return metadata;
}

function validateArtifact(artifact) {
  if (
    !artifact ||
    typeof artifact.path !== "string" ||
    !artifact.path ||
    typeof artifact.ruleset !== "string" ||
    !artifact.ruleset ||
    typeof artifact.behavior !== "string" ||
    !artifact.behavior ||
    typeof artifact.format !== "string" ||
    !Number.isSafeInteger(artifact.rule_count) ||
    artifact.rule_count < 0
  ) {
    throw new TypeError("stats.json contains an invalid artifact.");
  }
  return artifact;
}

/** Load a published ruleset manifest and its text artifacts from one origin. */
export function createRulesetStore({
  baseUrl = DEFAULT_BASE_URL,
  fetcher = globalThis.fetch,
} = {}) {
  if (typeof fetcher !== "function")
    throw new TypeError("A fetch function is required.");
  const base = new URL(baseUrl);
  const metadataURL = new URL("stats.json", base);
  let metadataRequest;
  const rulesetRequests = new Map();

  function artifactUrl(artifact) {
    validateArtifact(artifact);
    return new URL(artifact.path, base).href;
  }

  function loadMeta() {
    if (!metadataRequest) {
      metadataRequest = fetchText(fetcher, metadataURL.href).then((text) => {
        try {
          return validateMeta(JSON.parse(text));
        } catch (error) {
          throw new Error(
            `Read ${metadataURL.href}: invalid stats.json (${error instanceof Error ? error.message : String(error)}).`,
            { cause: error },
          );
        }
      });
      metadataRequest.catch(() => {
        metadataRequest = undefined;
      });
    }
    return metadataRequest;
  }

  function loadRules(rulesetName) {
    if (typeof rulesetName !== "string" || !rulesetName)
      throw new TypeError("A ruleset name is required.");
    if (!rulesetRequests.has(rulesetName)) {
      const request = loadMeta().then(async (metadata) => {
        const artifacts = metadata.artifacts
          .filter((artifact) => artifact.ruleset === rulesetName)
          .map(validateArtifact);
        if (!artifacts.length)
          throw new Error(
            `Ruleset ${rulesetName} is not present in stats.json.`,
          );

        const browserArtifacts = artifacts.filter((artifact) =>
          TEXT_BEHAVIORS.has(artifact.behavior),
        );
        const textArtifacts = browserArtifacts.filter(
          (artifact) => artifact.format === "text",
        );
        const binaryBehaviors = new Set(
          browserArtifacts
            .filter((artifact) => artifact.format === "mrs")
            .map((artifact) => artifact.behavior),
        );
        const textBehaviors = new Set(
          textArtifacts.map((artifact) => artifact.behavior),
        );
        const uncovered = [...binaryBehaviors].filter(
          (behavior) => !textBehaviors.has(behavior),
        );
        if (uncovered.length) {
          throw new Error(
            `Ruleset ${rulesetName} has no text artifact for ${uncovered.join(", ")}; it cannot be checked in the browser.`,
          );
        }
        if (!textArtifacts.length)
          throw new Error(
            `Ruleset ${rulesetName} has no text artifacts to browse.`,
          );

        const groups = await Promise.all(
          textArtifacts.map(async (artifact) => {
            const url = artifactUrl(artifact);
            const text = await fetchText(fetcher, url);
            let entries;
            try {
              entries = parseRules(text, artifact.behavior);
            } catch (error) {
              throw new Error(
                `Parse ${url} as ${artifact.behavior}: ${error instanceof Error ? error.message : String(error)}`,
                { cause: error },
              );
            }
            if (entries.length !== artifact.rule_count) {
              throw new Error(
                `Read ${url}: expected ${artifact.rule_count} rules from stats.json, found ${entries.length}.`,
              );
            }
            return entries.map((entry) => ({
              ...entry,
              artifact,
              behavior: artifact.behavior,
            }));
          }),
        );
        return groups.flat();
      });
      rulesetRequests.set(rulesetName, request);
      request.catch(() => {
        rulesetRequests.delete(rulesetName);
      });
    }
    return rulesetRequests.get(rulesetName);
  }

  return { loadMeta, loadRules, artifactUrl };
}
