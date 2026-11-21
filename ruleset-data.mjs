import { matchesQueryKind, parseRules, supportsBehavior } from "./ruleset-matcher.mjs";

export const DEFAULT_BASE_URL = new URL("./", import.meta.url);

const nonEmpty = (v) => typeof v === "string" && v.length > 0;

async function fetchText(fetcher, url) {
  try {
    const response = await fetcher(url);
    if (!response?.ok) {
      const status = response ? `${response.status} ${response.statusText}`.trim() : "no response";
      throw new Error(`HTTP ${status}`);
    }
    return await response.text();
  } catch (error) {
    const msg = error instanceof Error ? error.message : String(error);
    throw new Error(`Fetch ${url}: ${msg}`, { cause: error });
  }
}

function validateMeta(meta) {
  if (!meta || typeof meta !== "object" || !Array.isArray(meta.artifacts))
    throw new TypeError("stats.json must contain an artifacts array.");
  if (
    !Array.isArray(meta.priority) ||
    !meta.priority.length ||
    !meta.priority.every(nonEmpty) ||
    !nonEmpty(meta.fallback)
  )
    throw new TypeError("stats.json must contain a ruleset priority and fallback.");
  return meta;
}

function resolveArtifacts(meta, name) {
  const all = meta.artifacts.filter((a) => a?.ruleset === name);
  if (!all.length) throw new Error(`Ruleset ${name} is not present in stats.json.`);
  for (const a of all) {
    if (
      !nonEmpty(a.path) ||
      !nonEmpty(a.behavior) ||
      typeof a.format !== "string" ||
      !Number.isSafeInteger(a.rule_count) ||
      a.rule_count < 0
    )
      throw new TypeError("stats.json contains an invalid artifact.");
  }
  const supported = all.filter((a) => supportsBehavior(a.behavior));
  const text = supported.filter((a) => a.format === "text");
  const textBehaviors = new Set(text.map((a) => a.behavior));
  const uncovered = [
    ...new Set(
      supported
        .filter((a) => a.format === "mrs" && !textBehaviors.has(a.behavior))
        .map((a) => a.behavior),
    ),
  ];
  if (uncovered.length) {
    throw new Error(
      `Ruleset ${name} has no text artifact for ${uncovered.join(", ")}; it cannot be checked in the browser.`,
    );
  }
  if (!text.length) throw new Error(`Ruleset ${name} has no text artifacts to browse.`);
  return text;
}

/** Load a published ruleset manifest and its text artifacts from one origin. */
export function createRulesetStore({
  baseUrl = DEFAULT_BASE_URL,
  fetcher = globalThis.fetch,
} = {}) {
  if (typeof fetcher !== "function") throw new TypeError("A fetch function is required.");
  const base = new URL(baseUrl);
  const metaURL = new URL("stats.json", base).href;
  const cache = new Map();

  function cached(key, create) {
    if (!cache.has(key)) {
      const promise = create();
      cache.set(key, promise);
      promise.catch(() => {
        if (cache.get(key) === promise) cache.delete(key);
      });
    }
    return cache.get(key);
  }

  function artifactUrl(artifact) {
    if (!nonEmpty(artifact?.path)) throw new TypeError("Artifact path is required.");
    return new URL(artifact.path, base).href;
  }

  const loadMeta = () =>
    cached("meta", async () => {
      const text = await fetchText(fetcher, metaURL);
      try {
        return validateMeta(JSON.parse(text));
      } catch (error) {
        const msg = error instanceof Error ? error.message : String(error);
        throw new Error(`Read ${metaURL}: invalid stats.json (${msg}).`, { cause: error });
      }
    });

  const loadArtifact = (artifact) =>
    cached(`artifact:${artifact.path}`, async () => {
      const url = artifactUrl(artifact);
      const text = await fetchText(fetcher, url);
      const entries = parseRules(text, artifact.behavior);
      if (entries.length !== artifact.rule_count) {
        throw new Error(
          `Read ${url}: expected ${artifact.rule_count} rules from stats.json, found ${entries.length}.`,
        );
      }
      for (const entry of entries) entry.artifact = artifact;
      return entries;
    });

  function textArtifacts(name) {
    if (!nonEmpty(name)) throw new TypeError("A ruleset name is required.");
    return loadMeta().then((meta) => resolveArtifacts(meta, name));
  }

  const loadRules = (name) => {
    if (!nonEmpty(name)) throw new TypeError("A ruleset name is required.");
    return cached(`ruleset:${name}`, async () =>
      (await Promise.all((await textArtifacts(name)).map(loadArtifact))).flat(),
    );
  };

  const loadRulesForQuery = (name, kind) => {
    if (!nonEmpty(name)) throw new TypeError("A ruleset name is required.");
    if (kind !== "domain" && kind !== "ip") throw new TypeError("Query kind must be domain or ip.");
    return cached(`query:${name}:${kind}`, async () => {
      const artifacts = (await textArtifacts(name)).filter((a) =>
        matchesQueryKind(a.behavior, kind),
      );
      return (await Promise.all(artifacts.map(loadArtifact))).flat();
    });
  };

  return { loadMeta, textArtifacts, loadRules, loadRulesForQuery, artifactUrl };
}
