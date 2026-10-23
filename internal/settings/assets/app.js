"use strict";

// State of the page: the selected lights (id + name), the lights known to the
// bridge and the last saved settings to detect unsaved changes.
const state = {
  selected: [],
  bridgeLights: [],
  saved: "",
  activeResult: -1,
};

const $ = (id) => document.getElementById(id);
const elements = {
  selected: $("selected-lights"),
  search: $("light-search"),
  results: $("search-results"),
  bridgeStatus: $("bridge-status"),
  latitude: $("latitude"),
  longitude: $("longitude"),
  locationSource: $("location-source"),
  detect: $("detect-location"),
  save: $("save"),
  status: $("status"),
};

async function api(method, path, body) {
  const response = await fetch(path, {
    method,
    headers: body ? { "Content-Type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || `Request failed (${response.status})`);
  }
  return data;
}

function currentSettings() {
  return {
    location: {
      latitude: Number(elements.latitude.value),
      longitude: Number(elements.longitude.value),
    },
    lights: state.selected.map(({ id, name }) => ({ id, name })),
  };
}

function setStatus(message, kind = "") {
  elements.status.textContent = message;
  elements.status.className = `status ${kind}`;
}

function updateSaveButton() {
  const valid = elements.latitude.checkValidity() && elements.longitude.checkValidity()
    && elements.latitude.value !== "" && elements.longitude.value !== ""
    && state.selected.length > 0;
  const dirty = JSON.stringify(currentSettings()) !== state.saved;
  elements.save.disabled = !(valid && dirty);
  if (dirty && elements.status.classList.contains("success")) {
    setStatus("");
  }
}

// --- Selected lights -------------------------------------------------------

function renderSelected() {
  elements.selected.replaceChildren();

  if (state.selected.length === 0) {
    const empty = document.createElement("li");
    empty.className = "empty";
    empty.textContent = "No lights selected yet. Search for lights below to add them.";
    elements.selected.append(empty);
  }

  for (const light of state.selected) {
    const row = document.createElement("li");

    const name = document.createElement("span");
    name.className = "light-name";
    name.textContent = light.name || light.id;
    row.append(name);

    if (light.available === false) {
      const badge = document.createElement("span");
      badge.className = "badge";
      badge.textContent = "Not found on bridge";
      badge.title = `Light ID ${light.id} is unknown to the bridge.`;
      row.append(badge);
    }

    const remove = document.createElement("button");
    remove.type = "button";
    remove.className = "remove";
    remove.textContent = "×";
    remove.title = `Remove ${light.name || light.id}`;
    remove.setAttribute("aria-label", remove.title);
    remove.addEventListener("click", () => removeLight(light.id));
    row.append(remove);

    elements.selected.append(row);
  }

  updateSaveButton();
}

function addLight(light) {
  if (!state.selected.some((selected) => selected.id === light.id)) {
    state.selected.push({ id: light.id, name: light.name, available: true });
    renderSelected();
  }
  elements.search.value = "";
  closeResults();
  elements.search.focus();
}

function removeLight(id) {
  state.selected = state.selected.filter((light) => light.id !== id);
  renderSelected();
}

// --- Search ------------------------------------------------------------------

function matchingLights() {
  const query = elements.search.value.trim().toLowerCase();
  return state.bridgeLights.filter((light) =>
    !state.selected.some((selected) => selected.id === light.id)
    && light.name.toLowerCase().includes(query));
}

function renderResults() {
  const matches = matchingLights();
  elements.results.replaceChildren();

  if (matches.length === 0) {
    const empty = document.createElement("li");
    empty.className = "empty";
    empty.textContent = state.bridgeLights.length === 0 ? "No lights available." : "No matching lights.";
    elements.results.append(empty);
  }

  matches.forEach((light, index) => {
    const option = document.createElement("li");
    option.id = `result-${index}`;
    option.setAttribute("role", "option");
    option.setAttribute("aria-selected", String(index === state.activeResult));
    option.textContent = light.name;
    // mousedown instead of click, so the search field does not lose focus first.
    option.addEventListener("mousedown", (event) => {
      event.preventDefault();
      addLight(light);
    });
    elements.results.append(option);
  });

  elements.results.hidden = false;
  elements.search.setAttribute("aria-expanded", "true");
  elements.search.setAttribute("aria-activedescendant",
    state.activeResult >= 0 ? `result-${state.activeResult}` : "");
  return matches;
}

function closeResults() {
  state.activeResult = -1;
  elements.results.hidden = true;
  elements.search.setAttribute("aria-expanded", "false");
}

elements.search.addEventListener("input", () => {
  state.activeResult = -1;
  renderResults();
});
elements.search.addEventListener("focus", renderResults);
elements.search.addEventListener("blur", closeResults);
elements.search.addEventListener("keydown", (event) => {
  const matches = matchingLights();
  if (event.key === "ArrowDown") {
    state.activeResult = Math.min(state.activeResult + 1, matches.length - 1);
  } else if (event.key === "ArrowUp") {
    state.activeResult = Math.max(state.activeResult - 1, 0);
  } else if (event.key === "Enter") {
    const light = matches[state.activeResult >= 0 ? state.activeResult : 0];
    if (light && (state.activeResult >= 0 || matches.length === 1)) {
      addLight(light);
    }
    event.preventDefault();
    return;
  } else if (event.key === "Escape") {
    closeResults();
    return;
  } else {
    return;
  }
  event.preventDefault();
  renderResults();
});

async function loadBridgeLights() {
  try {
    state.bridgeLights = await api("GET", "/api/bridge/lights");
    elements.bridgeStatus.hidden = true;
  } catch (error) {
    state.bridgeLights = [];
    elements.bridgeStatus.textContent = error.message;
    elements.bridgeStatus.hidden = false;
  }
}

// --- Location ------------------------------------------------------------------

const locationSources = {
  "location-service": "Detected via your system's location service.",
  "time-zone": "Based on your time zone. Adjust it for more precise times.",
};

async function detectLocation() {
  elements.detect.disabled = true;
  elements.locationSource.textContent = "Detecting your location…";
  try {
    const location = await api("GET", "/api/location/default");
    elements.latitude.value = location.latitude;
    elements.longitude.value = location.longitude;
    elements.locationSource.textContent = locationSources[location.source] || "";
  } catch (error) {
    elements.locationSource.textContent = error.message;
  } finally {
    elements.detect.disabled = false;
    updateSaveButton();
  }
}

elements.detect.addEventListener("click", detectLocation);
elements.latitude.addEventListener("input", updateSaveButton);
elements.longitude.addEventListener("input", updateSaveButton);

// --- Load & save -------------------------------------------------------------

function showSettings(settings) {
  state.selected = settings.lights;
  elements.latitude.value = settings.location.latitude;
  elements.longitude.value = settings.location.longitude;
  state.saved = JSON.stringify(currentSettings());
  renderSelected();
}

elements.save.addEventListener("click", async () => {
  elements.save.disabled = true;
  setStatus("Saving…");
  try {
    showSettings(await api("PUT", "/api/settings", currentSettings()));
    setStatus("Settings saved.", "success");
  } catch (error) {
    setStatus(error.message, "error");
    updateSaveButton();
  }
});

window.addEventListener("beforeunload", (event) => {
  if (!elements.save.disabled) {
    event.preventDefault();
  }
});

async function init() {
  try {
    const [settings] = await Promise.all([api("GET", "/api/settings"), loadBridgeLights()]);
    showSettings(settings);
    if (settings.location.latitude === 0 && settings.location.longitude === 0) {
      await detectLocation();
    }
  } catch (error) {
    setStatus(error.message, "error");
  }
}

init();
