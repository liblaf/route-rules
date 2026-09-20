import { createRulesetStore } from "./ruleset-data.mjs";
import { matchRules, parseQuery } from "./ruleset-matcher.mjs";

const store = createRulesetStore();
const find = (id) => document.getElementById(id);
const count = (value) => value.toLocaleString();
const pageSize = 100;
let meta;
let entries = [];
let filtered = [];
let page = 0;
let browseRequest = 0;

function element(tag, text, className) {
  const node = document.createElement(tag);
  if (text !== undefined) node.textContent = text;
  if (className) node.className = className;
  return node;
}

function status(id, text, error = false) {
  const node = find(id);
  node.textContent = text;
  node.classList.toggle("error", error);
}

function renderPage() {
  const start = page * pageSize;
  const rows = filtered.slice(start, start + pageSize).map((entry) => {
    const row = element("tr");
    row.append(element("td", entry.type));
    const value = element("td");
    value.append(element("code", entry.value));
    row.append(value);
    return row;
  });
  find("entry-rows").replaceChildren(...rows);
  find("entry-empty").hidden = filtered.length !== 0;
  find("previous-page").disabled = page === 0;
  find("next-page").disabled = start + pageSize >= filtered.length;
  status(
    "entry-count",
    filtered.length
      ? `${count(start + 1)}–${count(Math.min(start + pageSize, filtered.length))} of ${count(filtered.length)} entries${filtered.length !== entries.length ? ` (${count(entries.length)} total)` : ""}`
      : `0 of ${count(entries.length)} entries`,
  );
}

function filterEntries() {
  const text = find("entry-filter").value.trim().toLowerCase();
  const type = find("entry-type").value;
  filtered = entries.filter(
    (entry) =>
      (!type || entry.type === type) &&
      entry.value.toLowerCase().includes(text),
  );
  page = 0;
  renderPage();
}

async function browse(name) {
  const request = ++browseRequest;
  find("ruleset-select").value = name;
  find("browse-content").hidden = true;
  find("browse-retry").hidden = true;
  status("browse-status", `Loading ${name}…`);
  try {
    const loaded = await store.loadRules(name);
    if (request !== browseRequest) return;
    entries = loaded;
    const types = [...new Set(entries.map((entry) => entry.type))].sort();
    find("entry-type").replaceChildren(
      new Option("All rule types", ""),
      ...types.map((type) => new Option(type, type)),
    );
    find("entry-filter").value = "";
    const links = meta.artifacts
      .filter(
        (artifact) => artifact.ruleset === name && artifact.format === "text",
      )
      .map((artifact) => {
        const link = element("a", artifact.path);
        link.href = store.artifactUrl(artifact);
        return link;
      });
    find("entry-downloads").replaceChildren(...links);
    status("browse-status", `${count(entries.length)} entries in ${name}.`);
    find("browse-content").hidden = false;
    filterEntries();
  } catch (error) {
    if (request !== browseRequest) return;
    status("browse-status", `Unable to load ${name}. ${error.message}`, true);
    find("browse-retry").hidden = false;
  }
}

function browseLink(name) {
  const link = element("a", `Browse ${name}`);
  link.href = "#browse";
  link.dataset.ruleset = name;
  return link;
}

function renderMatches(query, results) {
  const matched = results.filter((result) => result.matches.length);
  const unchecked = results.filter((result) => result.unsupported.length);
  status(
    "match-status",
    `${query.value}: ${matched.length} matching ruleset${matched.length === 1 ? "" : "s"}${unchecked.length ? " among supported rules" : ""}.`,
  );
  const cards = matched.map((result, index) => {
    const card = element("article", undefined, "match-card");
    const heading = element("h3", result.name);
    if (index === 0)
      heading.append(
        element(
          "span",
          unchecked.length ? "First confirmed match" : "First match",
          "badge",
        ),
      );
    card.append(heading, browseLink(result.name));
    const details = element("details");
    details.open = true;
    details.append(
      element(
        "summary",
        `${count(result.matches.length)} matching ${result.matches.length === 1 ? "entry" : "entries"}`,
      ),
    );
    const list = element("ul", undefined, "match-entries");
    for (const entry of result.matches) {
      const item = element("li");
      item.append(
        element("span", entry.type, "rule-type"),
        element("code", entry.value),
      );
      list.append(item);
    }
    details.append(list);
    card.append(details);
    return card;
  });
  if (!matched.length)
    cards.push(
      element(
        "p",
        unchecked.length
          ? "No supported entry matched. Some rules could not be evaluated; see below."
          : `No entries matched. The configured fallback is ${meta.fallback}.`,
        "empty-state",
      ),
    );
  if (unchecked.length) {
    const details = element("details", undefined, "match-card");
    details.append(element("summary", "Rules that could not be evaluated"));
    details.append(
      element(
        "p",
        "These rules need additional context or syntax this browser matcher does not support. They may change the first matching ruleset.",
      ),
    );
    const list = element("ul");
    for (const result of unchecked) {
      const types = [...new Set(result.unsupported.map((entry) => entry.type))];
      const item = element(
        "li",
        `${result.name}: ${count(result.unsupported.length)} entries (${types.join(", ")}). `,
      );
      item.append(browseLink(result.name));
      list.append(item);
    }
    details.append(list);
    cards.push(details);
  }
  find("match-results").replaceChildren(...cards);
}

async function testQuery(event) {
  event.preventDefault();
  if (find("match-submit").disabled) return;
  find("match-results").replaceChildren();
  let query;
  try {
    query = parseQuery(find("match-query").value);
    find("match-query").removeAttribute("aria-invalid");
  } catch (error) {
    find("match-query").setAttribute("aria-invalid", "true");
    status("match-status", error.message, true);
    return;
  }
  find("match-submit").disabled = true;
  find("match-form").setAttribute("aria-busy", "true");
  status("match-status", `Loading rulesets to check ${query.value}…`);
  try {
    const loaded = await Promise.all(
      meta.priority.map(async (name) => ({
        name,
        entries: await store.loadRules(name),
      })),
    );
    const results = [];
    for (const ruleset of loaded) {
      status("match-status", `Checking ${ruleset.name}…`);
      await new Promise(requestAnimationFrame);
      results.push({
        name: ruleset.name,
        ...matchRules(ruleset.entries, query),
      });
    }
    renderMatches(query, results);
  } catch (error) {
    status(
      "match-status",
      `Could not check every ruleset. ${error.message} Please try again.`,
      true,
    );
  } finally {
    find("match-submit").disabled = false;
    find("match-form").removeAttribute("aria-busy");
  }
}

async function initialize() {
  find("load-retry").hidden = true;
  find("explorer-status").hidden = false;
  status("explorer-status", "Loading ruleset information…");
  try {
    meta = await store.loadMeta();
    find("ruleset-select").replaceChildren(
      ...meta.priority.map((name) => new Option(name, name)),
    );
    find("ruleset-select").disabled = false;
    find("match-query").disabled = false;
    find("match-submit").disabled = false;
    status("explorer-status", "");
    find("explorer-status").hidden = true;
    await browse(meta.priority[0]);
  } catch (error) {
    status(
      "explorer-status",
      `Unable to load ruleset information. ${error.message}`,
      true,
    );
    find("load-retry").hidden = false;
  }
}

find("ruleset-select").addEventListener("change", (event) =>
  browse(event.target.value),
);
find("browse-retry").addEventListener("click", () =>
  browse(find("ruleset-select").value),
);
find("load-retry").addEventListener("click", initialize);
find("entry-filter").addEventListener("input", filterEntries);
find("entry-type").addEventListener("change", filterEntries);
find("previous-page").addEventListener("click", () => {
  page--;
  renderPage();
});
find("next-page").addEventListener("click", () => {
  page++;
  renderPage();
});
find("match-form").addEventListener("submit", testQuery);
document.addEventListener("click", (event) => {
  const link = event.target.closest("a[data-ruleset]");
  if (link && meta) browse(link.dataset.ruleset);
});
initialize();
