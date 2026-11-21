import { createRulesetStore } from "./ruleset-data.mjs";
import { matchRules, parseQuery } from "./ruleset-matcher.mjs";

const store = createRulesetStore();
const $ = (id) => document.getElementById(id);
const fmt = (n) => n.toLocaleString();
const pageSize = 100;
let meta;
let entries = [];
let filtered = [];
let page = 0;
let browseRequest = 0;

function el(tag, props, ...children) {
  const node = document.createElement(tag);
  if (typeof props === "string") node.className = props;
  else if (props) Object.assign(node, props);
  node.append(...children);
  return node;
}

function setStatus(id, text, error = false) {
  const node = $(id);
  node.textContent = text;
  node.classList.toggle("error", error);
}

function browseLink(name) {
  const link = el("a", { href: "#browse" }, `Browse ${name}`);
  link.dataset.ruleset = name;
  return link;
}

function renderPage() {
  const start = page * pageSize;
  const rows = filtered
    .slice(start, start + pageSize)
    .map((entry) =>
      el("tr", null, el("td", null, entry.type), el("td", null, el("code", null, entry.value))),
    );
  $("entry-rows").replaceChildren(...rows);
  $("entry-empty").hidden = filtered.length !== 0;
  $("previous-page").disabled = page === 0;
  $("next-page").disabled = start + pageSize >= filtered.length;
  const end = Math.min(start + pageSize, filtered.length);
  const totalNote = filtered.length !== entries.length ? ` (${fmt(entries.length)} total)` : "";
  setStatus(
    "entry-count",
    filtered.length
      ? `${fmt(start + 1)}–${fmt(end)} of ${fmt(filtered.length)} entries${totalNote}`
      : `0 of ${fmt(entries.length)} entries`,
  );
}

function filterEntries() {
  const text = $("entry-filter").value.trim().toLowerCase();
  const type = $("entry-type").value;
  filtered =
    !type && !text
      ? entries
      : entries.filter(
          (entry) =>
            (!type || entry.type === type) && (!text || entry.value.toLowerCase().includes(text)),
        );
  page = 0;
  renderPage();
}

async function browse(name) {
  const request = ++browseRequest;
  $("ruleset-select").value = name;
  $("browse-content").hidden = true;
  $("browse-retry").hidden = true;
  setStatus("browse-status", `Loading ${name}…`);
  try {
    const [loaded, artifacts] = await Promise.all([
      store.loadRules(name),
      store.textArtifacts(name),
    ]);
    if (request !== browseRequest) return;
    entries = loaded;
    const types = [...new Set(entries.map((e) => e.type))].sort();
    $("entry-type").replaceChildren(
      new Option("All rule types", ""),
      ...types.map((t) => new Option(t, t)),
    );
    $("entry-filter").value = "";
    $("entry-downloads").replaceChildren(
      ...artifacts.map((a) => el("a", { href: store.artifactUrl(a) }, a.path)),
    );
    setStatus("browse-status", `${fmt(entries.length)} entries in ${name}.`);
    $("browse-content").hidden = false;
    filterEntries();
  } catch (error) {
    if (request !== browseRequest) return;
    setStatus("browse-status", `Unable to load ${name}. ${error.message}`, true);
    $("browse-retry").hidden = false;
  }
}

function renderMatches(query, results) {
  const matched = results.filter((r) => r.matches.length);
  const unchecked = results.filter((r) => r.unsupported.length);
  setStatus(
    "match-status",
    `${query.value}: ${matched.length} matching ruleset${matched.length === 1 ? "" : "s"}${unchecked.length ? " among supported rules" : ""}.`,
  );
  const cards = matched.map((result, index) => {
    const heading = el(
      "h3",
      null,
      result.name,
      ...(index === 0
        ? [el("span", "badge", unchecked.length ? "First confirmed match" : "First match")]
        : []),
    );
    const items = result.matches.map((entry) =>
      el("li", null, el("span", "rule-type", entry.type), el("code", null, entry.value)),
    );
    const summary = `${fmt(result.matches.length)} matching ${result.matches.length === 1 ? "entry" : "entries"}`;
    return el(
      "article",
      "match-card",
      heading,
      browseLink(result.name),
      el(
        "details",
        { open: true },
        el("summary", null, summary),
        el("ul", "match-entries", ...items),
      ),
    );
  });
  if (!matched.length) {
    cards.push(
      el(
        "p",
        "empty-state",
        unchecked.length
          ? "No supported entry matched. Some rules could not be evaluated; see below."
          : `No entries matched. The configured fallback is ${meta.fallback}.`,
      ),
    );
  }
  if (unchecked.length) {
    const items = unchecked.map((r) => {
      const types = [...new Set(r.unsupported.map((e) => e.type))].join(", ");
      return el(
        "li",
        null,
        `${r.name}: ${fmt(r.unsupported.length)} entries (${types}). `,
        browseLink(r.name),
      );
    });
    cards.push(
      el(
        "details",
        "match-card",
        el("summary", null, "Rules that could not be evaluated"),
        el(
          "p",
          null,
          "These rules need additional context or syntax this browser matcher does not support. They may change the first matching ruleset.",
        ),
        el("ul", null, ...items),
      ),
    );
  }
  $("match-results").replaceChildren(...cards);
}

async function testQuery(event) {
  event.preventDefault();
  if ($("match-submit").disabled) return;
  $("match-results").replaceChildren();
  let query;
  try {
    query = parseQuery($("match-query").value);
    $("match-query").removeAttribute("aria-invalid");
  } catch (error) {
    $("match-query").setAttribute("aria-invalid", "true");
    setStatus("match-status", error.message, true);
    return;
  }
  $("match-submit").disabled = true;
  $("match-form").setAttribute("aria-busy", "true");
  setStatus("match-status", `Loading rulesets to check ${query.value}…`);
  try {
    const results = await Promise.all(
      meta.priority.map(async (name) => ({
        name,
        ...matchRules(await store.loadRulesForQuery(name, query.kind), query),
      })),
    );
    renderMatches(query, results);
  } catch (error) {
    setStatus(
      "match-status",
      `Could not check every ruleset. ${error.message} Please try again.`,
      true,
    );
  } finally {
    $("match-submit").disabled = false;
    $("match-form").removeAttribute("aria-busy");
  }
}

async function initialize() {
  $("load-retry").hidden = true;
  $("explorer-status").hidden = false;
  setStatus("explorer-status", "Loading ruleset information…");
  try {
    meta = await store.loadMeta();
    $("ruleset-select").replaceChildren(...meta.priority.map((name) => new Option(name, name)));
    for (const id of ["ruleset-select", "match-query", "match-submit"]) $(id).disabled = false;
    setStatus("explorer-status", "");
    $("explorer-status").hidden = true;
    await browse(meta.priority[0]);
  } catch (error) {
    setStatus("explorer-status", `Unable to load ruleset information. ${error.message}`, true);
    $("load-retry").hidden = false;
  }
}

$("ruleset-select").addEventListener("change", (e) => browse(e.target.value));
$("browse-retry").addEventListener("click", () => browse($("ruleset-select").value));
$("load-retry").addEventListener("click", initialize);
$("entry-filter").addEventListener("input", filterEntries);
$("entry-type").addEventListener("change", filterEntries);
$("previous-page").addEventListener("click", () => {
  page--;
  renderPage();
});
$("next-page").addEventListener("click", () => {
  page++;
  renderPage();
});
$("match-form").addEventListener("submit", testQuery);
document.addEventListener("click", (e) => {
  const link = e.target.closest("a[data-ruleset]");
  if (link && meta) browse(link.dataset.ruleset);
});
initialize();
