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
const UNSUPPORTED_REASON =
  "This rule needs request context that this checker does not have.";

function inputError() {
  return new Error("Enter a valid domain name or IP address.");
}

function normalizeDomain(value) {
  const input = value.trim().replace(/\.$/, "").toLowerCase();
  if (
    !input ||
    input.length > 253 ||
    input.includes("..") ||
    /[\s\\/@:?#%]/.test(input)
  )
    throw inputError();
  if (/^[0-9.]+$/.test(input) || /^0x[0-9a-f]+$/i.test(input))
    throw inputError();
  let hostname;
  try {
    hostname = new URL(`http://${input}`).hostname;
  } catch {
    throw inputError();
  }
  if (!hostname || hostname.includes(":") || parseIpv4(hostname) !== null)
    throw inputError();
  if (
    !hostname
      .split(".")
      .every((label) => /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/.test(label))
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

function parseIp(value) {
  const input = value.trim().toLowerCase();
  const ipv4 = parseIpv4(input);
  if (ipv4 !== null) return { family: 4, numeric: ipv4, value: input };
  if (!input || input.includes("%")) return null;
  let ipv6 = input;
  if (ipv6.includes(".")) {
    const separator = ipv6.lastIndexOf(":");
    const tail = separator < 0 ? null : parseIpv4(ipv6.slice(separator + 1));
    if (tail === null) return null;
    ipv6 = `${ipv6.slice(0, separator + 1)}${(tail >> 16n).toString(16)}:${(tail & 0xffffn).toString(16)}`;
  }
  const halves = ipv6.split("::");
  if (halves.length > 2) return null;
  const left = halves[0] ? halves[0].split(":") : [];
  const right = halves.length === 2 && halves[1] ? halves[1].split(":") : [];
  if (left.concat(right).some((group) => !/^[0-9a-f]{1,4}$/.test(group)))
    return null;
  if (
    (halves.length === 1 && left.length !== 8) ||
    (halves.length === 2 && left.length + right.length >= 8)
  )
    return null;
  const groups = [
    ...left,
    ...Array(8 - left.length - right.length).fill("0"),
    ...right,
  ].map((group) => Number.parseInt(group, 16));
  let numeric = 0n;
  for (const group of groups) numeric = (numeric << 16n) | BigInt(group);
  return { family: 6, numeric, value: formatIpv6(groups) };
}

function formatIpv6(groups) {
  let start = -1;
  let length = 0;
  for (let index = 0; index < groups.length; ) {
    if (groups[index] !== 0) {
      index += 1;
      continue;
    }
    let end = index;
    while (end < groups.length && groups[end] === 0) end += 1;
    if (end - index > length) [start, length] = [index, end - index];
    index = end;
  }
  if (length < 2) return groups.map((group) => group.toString(16)).join(":");
  const before = groups
    .slice(0, start)
    .map((group) => group.toString(16))
    .join(":");
  const after = groups
    .slice(start + length)
    .map((group) => group.toString(16))
    .join(":");
  return `${before}::${after}`.replace(/^::$/, "::");
}

function parseCidr(value) {
  const [address, prefixText, ...rest] = value.trim().split("/");
  if (rest.length || !/^\d+$/.test(prefixText ?? "")) return null;
  const ip = parseIp(address);
  if (!ip) return null;
  const width = ip.family === 4 ? 32 : 128;
  const prefix = Number(prefixText);
  if (prefix > width) return null;
  const mask =
    prefix === 0 ? 0n : ((1n << BigInt(prefix)) - 1n) << BigInt(width - prefix);
  return { ...ip, prefix, mask, network: ip.numeric & mask };
}

function isWholeLabelWildcard(value) {
  return (
    value.includes("*") &&
    !value.includes("?") &&
    value.split(".").every((label) => !/[?*]/.test(label) || label === "*")
  );
}

function globMatches(pattern, value) {
  let source = "";
  for (const character of pattern) {
    if (character === "*") source += ".*";
    else if (character === "?") source += ".";
    else
      source += /[|\\{}()[\]^$+?.]/.test(character)
        ? `\\${character}`
        : character;
  }
  return new RegExp(`^${source}$`).test(value);
}

function wholeLabelWildcardMatches(pattern, value) {
  const expected = pattern.split(".");
  const actual = value.split(".");
  return (
    expected.length === actual.length &&
    expected.every((label, index) => label === "*" || label === actual[index])
  );
}

function unsupported(type, value, raw, line, reason = UNSUPPORTED_REASON) {
  const queryKinds = type.startsWith("DOMAIN")
    ? ["domain"]
    : IP_TYPES.has(type)
      ? ["ip"]
      : ["domain", "ip"];
  return { type, value, raw, line, supported: false, reason, queryKinds };
}

function regexForBrowser(value) {
  // Mihomo uses Go's RE2. Do not reinterpret its Go-only extensions in JS.
  if (/\\(?:A|C|z|Q|E|p|P|a)|\\x\{|\[\[:|\(\?[a-zA-Z-]+:|\(\?P[<=]/.test(value))
    return null;
  try {
    return new RegExp(value);
  } catch {
    return null;
  }
}

function domainEntry(value, raw, line) {
  if (value.startsWith("+."))
    return {
      type: "DOMAIN-SUFFIX",
      value: normalizeDomain(value.slice(2)),
      raw,
      line,
      supported: true,
    };
  if (value.includes("*")) {
    const normalized = value.trim().toLowerCase();
    return isWholeLabelWildcard(normalized)
      ? {
          type: "DOMAIN-WILDCARD",
          value: normalized,
          raw,
          line,
          supported: true,
          wholeLabel: true,
        }
      : unsupported(
          "DOMAIN-WILDCARD",
          raw,
          raw,
          line,
          "Invalid domain wildcard syntax.",
        );
  }
  return {
    type: "DOMAIN",
    value: normalizeDomain(value),
    raw,
    line,
    supported: true,
  };
}

function classicalEntry(raw, line) {
  const separator = raw.indexOf(",");
  const type = raw
    .slice(0, separator < 0 ? raw.length : separator)
    .trim()
    .toUpperCase();
  const payload = separator < 0 ? "" : raw.slice(separator + 1).trim();
  // A regex may contain commas (for example `{0,5}`), whereas target-IP
  // rules can carry comma-separated options such as `no-resolve`.
  const trimmed =
    type === "DOMAIN-REGEX" ? payload : payload.split(",", 1)[0].trim();
  if (IP_TYPES.has(type)) {
    const cidr = parseCidr(trimmed);
    return cidr
      ? { type, value: trimmed, raw, line, supported: true, cidr }
      : unsupported(type, trimmed, raw, line, "Invalid IP-CIDR rule.");
  }
  if (!DOMAIN_TYPES.has(type))
    return unsupported(type || "UNKNOWN", trimmed, raw, line);
  try {
    if (type === "DOMAIN")
      return {
        type,
        value: normalizeDomain(trimmed),
        raw,
        line,
        supported: true,
      };
    if (type === "DOMAIN-SUFFIX")
      return {
        type,
        value: normalizeDomain(trimmed),
        raw,
        line,
        supported: true,
      };
    if (type === "DOMAIN-KEYWORD")
      return { type, value: trimmed.toLowerCase(), raw, line, supported: true };
    if (type === "DOMAIN-WILDCARD")
      return {
        type,
        value: trimmed.toLowerCase(),
        raw,
        line,
        supported: true,
        wholeLabel: false,
      };
    const regex = regexForBrowser(trimmed);
    return regex
      ? { type, value: trimmed, raw, line, supported: true, regex }
      : unsupported(
          type,
          trimmed,
          raw,
          line,
          "This Go regular expression cannot be matched safely in the browser.",
        );
  } catch {
    return unsupported(type, trimmed, raw, line, "Invalid domain rule.");
  }
}

export function parseRules(text, behavior) {
  if (typeof text !== "string")
    throw new TypeError("Ruleset text must be a string.");
  if (!BEHAVIORS.has(behavior))
    throw new TypeError("Unknown ruleset behavior.");
  return text.split(/\r?\n/).flatMap((source, offset) => {
    const raw = source.trim();
    const line = offset + 1;
    if (!raw || raw.startsWith("#")) return [];
    if (behavior === "domain") {
      try {
        return [domainEntry(raw, raw, line)];
      } catch {
        return [unsupported("DOMAIN", raw, raw, line, "Invalid domain rule.")];
      }
    }
    if (behavior === "ipcidr") {
      const cidr = parseCidr(raw);
      return [
        cidr
          ? { type: "IP-CIDR", value: raw, raw, line, supported: true, cidr }
          : unsupported("IP-CIDR", raw, raw, line, "Invalid IP-CIDR rule."),
      ];
    }
    return [classicalEntry(raw, line)];
  });
}

export function parseQuery(input) {
  if (typeof input !== "string") throw inputError();
  const ip = parseIp(input);
  return ip
    ? { kind: "ip", ...ip }
    : { kind: "domain", value: normalizeDomain(input) };
}

export function matchRules(entries, query) {
  if (!Array.isArray(entries)) throw new TypeError("Rules must be an array.");
  if (!query || (query.kind !== "domain" && query.kind !== "ip"))
    throw new TypeError("Query must be parsed first.");
  const unsupported = entries.filter(
    (entry) => !entry.supported && entry.queryKinds.includes(query.kind),
  );
  const matches = entries.filter((entry) => {
    if (!entry.supported) return false;
    if (query.kind === "ip")
      return (
        entry.cidr?.family === query.family &&
        (query.numeric & entry.cidr.mask) === entry.cidr.network
      );
    if (entry.type === "DOMAIN") return query.value === entry.value;
    if (entry.type === "DOMAIN-SUFFIX")
      return (
        query.value === entry.value || query.value.endsWith(`.${entry.value}`)
      );
    if (entry.type === "DOMAIN-KEYWORD")
      return query.value.includes(entry.value);
    if (entry.type === "DOMAIN-WILDCARD")
      return entry.wholeLabel
        ? wholeLabelWildcardMatches(entry.value, query.value)
        : globMatches(entry.value, query.value);
    return entry.type === "DOMAIN-REGEX" && entry.regex.test(query.value);
  });
  return { matches, unsupported };
}
