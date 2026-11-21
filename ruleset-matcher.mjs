/** Browser-safe matching for the rule types this site publishes. */

const BEHAVIORS = new Set(["domain", "ipcidr", "classical"]);
const DOMAIN_TYPES = new Set([
  "DOMAIN",
  "DOMAIN-SUFFIX",
  "DOMAIN-KEYWORD",
  "DOMAIN-WILDCARD",
  "DOMAIN-REGEX",
]);
const IP_TYPES = new Set(["IP-CIDR", "IP-CIDR6"]);
const ASCII_DOMAIN_RE =
  /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)*$/;
const NUMERIC_LAST_LABEL_RE = /(?:^|\.)(?:[0-9]+|0x[0-9a-f]*)$/;
const GO_ONLY_REGEX_RE = /\\(?:A|C|z|Q|E|p|P|a)|\\x\{|\[\[:|\(\?[a-zA-Z-]+:|\(\?P[<=]/;

export const supportsBehavior = (behavior) => BEHAVIORS.has(behavior);
export const matchesQueryKind = (behavior, kind) =>
  behavior === "classical" || behavior === (kind === "ip" ? "ipcidr" : "domain");

const inputError = () => new Error("Enter a valid domain name or IP address.");

function normalizeDomain(value) {
  let input = value.trim().toLowerCase();
  if (input.endsWith(".")) input = input.slice(0, -1);
  if (!input || input.length > 253) throw inputError();
  if (ASCII_DOMAIN_RE.test(input) && !input.includes("xn--") && !NUMERIC_LAST_LABEL_RE.test(input))
    return input;
  if (input.includes("..") || /[\s\\/@:?#%]/.test(input) || /^[0-9.]+$|^0x[0-9a-f]+$/.test(input))
    throw inputError();
  let hostname;
  try {
    hostname = new URL(`http://${input}`).hostname;
  } catch {
    throw inputError();
  }
  if (
    !hostname ||
    hostname.includes(":") ||
    parseIpv4(hostname) !== null ||
    !ASCII_DOMAIN_RE.test(hostname)
  )
    throw inputError();
  return hostname;
}

function parseIpv4(value) {
  const parts = value.split(".");
  if (parts.length !== 4) return null;
  let numeric = 0n;
  for (const part of parts) {
    if (!/^(?:0|[1-9]\d{0,2})$/.test(part)) return null;
    const octet = Number(part);
    if (octet > 255) return null;
    numeric = (numeric << 8n) | BigInt(octet);
  }
  return numeric;
}

function parseIpv6(input) {
  if (!input.includes(":") || !/^[0-9a-f:.]+$/.test(input)) return null;
  let canonical;
  try {
    const { hostname } = new URL(`http://[${input}]/`);
    if (!hostname.startsWith("[") || !hostname.endsWith("]")) return null;
    canonical = hostname.slice(1, -1);
  } catch {
    return null;
  }
  const [leftText, rightText] = canonical.split("::");
  const left = leftText ? leftText.split(":") : [];
  const right = rightText ? rightText.split(":") : [];
  let numeric = 0n;
  for (const group of [...left, ...Array(8 - left.length - right.length).fill("0"), ...right])
    numeric = (numeric << 16n) | BigInt(Number.parseInt(group, 16));
  return { family: 6, numeric, value: canonical };
}

function parseIp(value) {
  const input = value.trim().toLowerCase();
  const ipv4 = parseIpv4(input);
  return ipv4 !== null ? { family: 4, numeric: ipv4, value: input } : parseIpv6(input);
}

function parseCidr(value) {
  const [address, prefixText, ...rest] = value.trim().split("/");
  if (rest.length || !/^\d+$/.test(prefixText ?? "")) return null;
  const ip = parseIp(address);
  if (!ip) return null;
  const width = ip.family === 4 ? 32 : 128;
  const prefix = Number(prefixText);
  if (prefix > width) return null;
  const mask = prefix === 0 ? 0n : ((1n << BigInt(prefix)) - 1n) << BigInt(width - prefix);
  return { ...ip, prefix, mask, network: ip.numeric & mask };
}

function compileGlob(pattern) {
  const source = pattern.replace(/[|\\{}()[\]^$+.*?]/g, (ch) =>
    ch === "*" ? ".*" : ch === "?" ? "." : `\\${ch}`,
  );
  return new RegExp(`^${source}$`);
}

function unsupported(type, value, queryKinds) {
  const kinds =
    queryKinds ??
    (type.startsWith("DOMAIN") ? ["domain"] : IP_TYPES.has(type) ? ["ip"] : ["domain", "ip"]);
  return { type, value, supported: false, queryKinds: kinds };
}

function cidrEntry(type, value, target = value) {
  const cidr = parseCidr(target);
  return cidr ? { type, value, supported: true, cidr } : unsupported(type, value);
}

function domainEntry(raw) {
  if (raw.includes("*")) {
    let value = raw.toLowerCase();
    if (value.endsWith(".")) value = value.slice(0, -1);
    const labels = value.split(".");
    const valid =
      !/[\s,/]/.test(value) &&
      labels.every((label) => label && (label === "*" || !/[?*]/.test(label)));
    return valid
      ? { type: "DOMAIN-WILDCARD", value, supported: true, labels }
      : unsupported("DOMAIN-WILDCARD", raw);
  }
  try {
    return raw.startsWith("+.")
      ? { type: "DOMAIN-SUFFIX", value: normalizeDomain(raw.slice(2)), supported: true }
      : { type: "DOMAIN", value: normalizeDomain(raw), supported: true };
  } catch {
    return unsupported("DOMAIN", raw);
  }
}

function classicalEntry(raw) {
  const comma = raw.indexOf(",");
  const type = (comma < 0 ? raw : raw.slice(0, comma)).trim().toUpperCase();
  const payload = comma < 0 ? "" : raw.slice(comma + 1).trim();
  if (IP_TYPES.has(type)) {
    const [first, ...options] = payload.split(",");
    if (options.some((opt) => opt.trim().toLowerCase() === "src"))
      return unsupported(type, payload, ["domain", "ip"]);
    return cidrEntry(type, payload, first.trim());
  }
  if (!DOMAIN_TYPES.has(type)) return unsupported(type || "UNKNOWN", payload);
  // Regexes may contain commas (e.g. `{0,5}`); other directives put their target before the first comma.
  const target = type === "DOMAIN-REGEX" ? payload : payload.split(",", 1)[0].trim();
  if (!target) return unsupported(type, target);
  try {
    if (type === "DOMAIN" || type === "DOMAIN-SUFFIX")
      return { type, value: normalizeDomain(target), supported: true };
    if (type === "DOMAIN-KEYWORD") return { type, value: target.toLowerCase(), supported: true };
    if (type === "DOMAIN-WILDCARD") {
      const value = target.toLowerCase();
      return { type, value, supported: true, regex: compileGlob(value) };
    }
    if (GO_ONLY_REGEX_RE.test(target)) return unsupported(type, target);
    return { type, value: target, supported: true, regex: new RegExp(target) };
  } catch {
    return unsupported(type, target);
  }
}

export function parseRules(text, behavior) {
  if (typeof text !== "string") throw new TypeError("Ruleset text must be a string.");
  if (!supportsBehavior(behavior)) throw new TypeError("Unknown ruleset behavior.");
  const entries = [];
  for (const line of text.split(/\r?\n/)) {
    const raw = line.trim();
    if (!raw || raw.startsWith("#")) continue;
    entries.push(
      behavior === "domain"
        ? domainEntry(raw)
        : behavior === "ipcidr"
          ? cidrEntry("IP-CIDR", raw)
          : classicalEntry(raw),
    );
  }
  return entries;
}

export function parseQuery(input) {
  if (typeof input !== "string") throw inputError();
  const ip = parseIp(input);
  return ip ? { kind: "ip", ...ip } : { kind: "domain", value: normalizeDomain(input) };
}

function entryMatches(entry, query) {
  if (query.kind === "ip")
    return (
      entry.cidr?.family === query.family &&
      (query.numeric & entry.cidr.mask) === entry.cidr.network
    );
  if (entry.type === "DOMAIN") return query.value === entry.value;
  if (entry.type === "DOMAIN-SUFFIX")
    return query.value === entry.value || query.value.endsWith(`.${entry.value}`);
  if (entry.type === "DOMAIN-KEYWORD") return query.value.includes(entry.value);
  if (entry.type === "DOMAIN-WILDCARD") {
    if (!entry.labels) return entry.regex.test(query.value);
    const actual = query.value.split(".");
    return (
      entry.labels.length === actual.length &&
      entry.labels.every((label, i) => label === "*" || label === actual[i])
    );
  }
  return entry.type === "DOMAIN-REGEX" && entry.regex.test(query.value);
}

export function matchRules(entries, query) {
  if (!Array.isArray(entries)) throw new TypeError("Rules must be an array.");
  if (!query || (query.kind !== "domain" && query.kind !== "ip"))
    throw new TypeError("Query must be parsed first.");
  const matches = [];
  const unsupportedEntries = [];
  for (const entry of entries) {
    if (!entry.supported) {
      if (entry.queryKinds.includes(query.kind)) unsupportedEntries.push(entry);
    } else if (entryMatches(entry, query)) {
      matches.push(entry);
    }
  }
  return { matches, unsupported: unsupportedEntries };
}
