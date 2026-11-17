const state = {
  equipment: [],
  borrowers: [],
  borrowings: [],
  dashboard: {},
  activeView: "dashboard",
};

const viewDetails = {
  dashboard: ["Dashboard", "LAB OVERVIEW"],
  equipment: ["Equipment", "INVENTORY MANAGEMENT"],
  borrowers: ["Borrowers", "BORROWER DIRECTORY"],
  borrowings: ["Borrowings", "TRANSACTION HISTORY"],
};

const dateToday = new Intl.DateTimeFormat("en", { weekday: "short", month: "short", day: "numeric", year: "numeric" });
document.querySelector("#today-label").textContent = dateToday.format(new Date());

function escapeHTML(value) {
  return String(value ?? "").replace(/[&<>"']/g, (character) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  })[character]);
}

async function api(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: { "Content-Type": "application/json", ...options.headers },
  });
  const text = await response.text();
  let result = null;
  if (text) {
    try { result = JSON.parse(text); } catch { result = { error: text.trim() }; }
  }
  if (!response.ok) throw new Error(result?.error || `Request failed (${response.status})`);
  return result;
}

function showNotice(message = "") {
  const notice = document.querySelector("#notice");
  notice.textContent = message;
  notice.hidden = !message;
}

async function refreshData() {
  try {
    const [equipment, borrowers, borrowings, dashboard] = await Promise.all([
      api("/api/equipment"), api("/api/borrowers"), api("/api/borrowings"), api("/api/dashboard"),
    ]);
    Object.assign(state, { equipment, borrowers, borrowings, dashboard });
    renderAll();
    showNotice();
  } catch (error) {
    showNotice(`Could not load LabBorrow data: ${error.message}. Check that the Go server is running.`);
  }
}

function renderAll() {
  renderDashboard();
  renderEquipment();
  renderBorrowers();
  renderBorrowings();
}

function borrowerName(id) {
  return state.borrowers.find((borrower) => borrower.id === id)?.name || `Borrower #${id}`;
}

function equipmentName(id) {
  return state.equipment.find((equipment) => equipment.id === id)?.name || `Equipment #${id}`;
}

function emptyRow(columns, message) {
  return `<tr><td class="empty-state" colspan="${columns}">${escapeHTML(message)}</td></tr>`;
}

function renderDashboard() {
  document.querySelector("#metric-equipment").textContent = state.dashboard.total_equipment ?? 0;
  document.querySelector("#metric-borrowers").textContent = state.dashboard.total_borrowers ?? 0;
  document.querySelector("#metric-active").textContent = state.dashboard.active_borrowings ?? 0;
  document.querySelector("#metric-available").textContent = state.dashboard.available_items ?? 0;

  const recent = [...state.borrowings].reverse().slice(0, 5);
  document.querySelector("#recent-table").innerHTML = recent.length ? recent.map((borrowing) => `
    <tr>
      <td><strong>${escapeHTML(borrowerName(borrowing.borrower_id))}</strong></td>
      <td>${escapeHTML(equipmentName(borrowing.equipment_id))}</td>
      <td>${escapeHTML(borrowing.borrow_date)}</td>
      <td><span class="status-pill ${borrowing.status === "Returned" ? "returned" : ""}">${escapeHTML(borrowing.status)}</span></td>
    </tr>`).join("") : emptyRow(4, "No borrowing activity yet.");

  const stock = [...state.equipment].sort((a, b) => a.quantity - b.quantity).slice(0, 5);
  document.querySelector("#stock-list").innerHTML = stock.length ? stock.map((equipment) => `
    <div class="stock-item">
      <span class="item-mark" aria-hidden="true">${escapeHTML(equipment.name.slice(0, 1).toUpperCase())}</span>
      <span class="stock-copy"><strong>${escapeHTML(equipment.name)}</strong><small>${escapeHTML(equipment.category)}</small></span>
      <span class="stock-count ${equipment.quantity < 2 ? "low" : ""}">${equipment.quantity} available</span>
    </div>`).join("") : `<p class="empty-state">No equipment in inventory.</p>`;
}

function matchesSearch(values, query) {
  return values.some((value) => String(value ?? "").toLowerCase().includes(query));
}

function renderEquipment() {
  const query = document.querySelector("#equipment-search").value.trim().toLowerCase();
  const rows = state.equipment.filter((item) => matchesSearch([item.name, item.category, item.condition], query));
  document.querySelector("#equipment-count").textContent = `${rows.length} ${rows.length === 1 ? "item" : "items"}`;
  document.querySelector("#equipment-table").innerHTML = rows.length ? rows.map((item) => `
    <tr>
      <td><div class="primary-cell"><span class="item-mark">${escapeHTML(item.name.slice(0, 1).toUpperCase())}</span><strong>${escapeHTML(item.name)}</strong></div></td>
      <td>${escapeHTML(item.category)}</td><td>${escapeHTML(item.condition)}</td>
      <td><strong>${item.quantity}</strong> units</td>
      <td><div class="row-actions"><button class="row-button" data-action="edit" data-kind="equipment" data-id="${item.id}">Edit</button><button class="row-button danger" data-action="delete" data-kind="equipment" data-id="${item.id}">Delete</button></div></td>
    </tr>`).join("") : emptyRow(5, "No matching equipment.");
}

function renderBorrowers() {
  const query = document.querySelector("#borrowers-search").value.trim().toLowerCase();
  const rows = state.borrowers.filter((item) => matchesSearch([item.name, item.student_id, item.email, item.phone], query));
  document.querySelector("#borrowers-count").textContent = `${rows.length} ${rows.length === 1 ? "person" : "people"}`;
  document.querySelector("#borrowers-table").innerHTML = rows.length ? rows.map((item) => `
    <tr>
      <td><div class="primary-cell"><span class="item-mark">${escapeHTML(item.name.slice(0, 1).toUpperCase())}</span><strong>${escapeHTML(item.name)}</strong></div></td>
      <td>${escapeHTML(item.student_id)}</td><td>${escapeHTML(item.email)}</td><td>${escapeHTML(item.phone)}</td>
      <td><div class="row-actions"><button class="row-button" data-action="edit" data-kind="borrower" data-id="${item.id}">Edit</button><button class="row-button danger" data-action="delete" data-kind="borrower" data-id="${item.id}">Delete</button></div></td>
    </tr>`).join("") : emptyRow(5, "No matching borrowers.");
}

function renderBorrowings() {
  const query = document.querySelector("#borrowings-search").value.trim().toLowerCase();
  const rows = state.borrowings.filter((item) => matchesSearch([
    borrowerName(item.borrower_id), equipmentName(item.equipment_id), item.borrow_date, item.return_date, item.status,
  ], query));
  document.querySelector("#borrowings-count").textContent = `${rows.length} ${rows.length === 1 ? "record" : "records"}`;
  document.querySelector("#borrowings-table").innerHTML = rows.length ? rows.map((item) => `
    <tr>
      <td><strong>${escapeHTML(borrowerName(item.borrower_id))}</strong></td>
      <td>${escapeHTML(equipmentName(item.equipment_id))}</td><td>${escapeHTML(item.borrow_date)}</td><td>${escapeHTML(item.return_date)}</td>
      <td><span class="status-pill ${item.status === "Returned" ? "returned" : ""}">${escapeHTML(item.status)}</span></td>
      <td><div class="row-actions">${item.status === "Borrowed" ? `<button class="row-button return" data-action="return" data-id="${item.id}">Return</button>` : ""}<button class="row-button" data-action="edit" data-kind="borrowing" data-id="${item.id}">Edit</button><button class="row-button danger" data-action="delete" data-kind="borrowing" data-id="${item.id}">Delete</button></div></td>
    </tr>`).join("") : emptyRow(6, "No matching borrowing records.");
}

function setView(name) {
  state.activeView = name;
  document.querySelectorAll(".page-view").forEach((view) => {
    const active = view.id === `${name}-view`;
    view.classList.toggle("is-visible", active);
    view.hidden = !active;
  });
  document.querySelectorAll(".nav-link").forEach((button) => {
    button.classList.toggle("is-active", button.dataset.view === name);
  });
  const [title, eyebrow] = viewDetails[name];
  document.querySelector("#page-title").textContent = title;
  document.querySelector("#page-eyebrow").textContent = eyebrow;
  document.title = `LabBorrow | ${title}`;
}

const fieldDefinitions = {
  equipment: [
    { name: "name", label: "Equipment name", required: true },
    { name: "category", label: "Category", required: true },
    { name: "quantity", label: "Available quantity", type: "number", min: 0, required: true },
    { name: "condition", label: "Condition", required: true },
  ],
  borrower: [
    { name: "name", label: "Full name", required: true },
    { name: "student_id", label: "Student ID", required: true },
    { name: "email", label: "Email address", type: "email" },
    { name: "phone", label: "Phone number", type: "tel" },
  ],
  borrowing: [
    { name: "borrower_id", label: "Borrower", type: "borrower-select", required: true },
    { name: "equipment_id", label: "Equipment", type: "equipment-select", required: true },
    { name: "borrow_date", label: "Borrow date", type: "date", required: true },
    { name: "return_date", label: "Due date", type: "date", required: true },
  ],
};

function fieldMarkup(field, value) {
  const required = field.required ? "required" : "";
  let control;
  if (field.type === "borrower-select" || field.type === "equipment-select") {
    const records = field.type === "borrower-select" ? state.borrowers : state.equipment;
    const label = field.type === "borrower-select" ? (item) => `${item.name} · ${item.student_id}` : (item) => `${item.name} · ${item.quantity} available`;
    control = `<select name="${field.name}" ${required}><option value="">Choose ${field.type === "borrower-select" ? "a borrower" : "equipment"}</option>${records.map((item) => `<option value="${item.id}" ${Number(value) === item.id ? "selected" : ""}>${escapeHTML(label(item))}</option>`).join("")}</select>`;
  } else {
    const type = field.type || "text";
    control = `<input name="${field.name}" type="${type}" value="${escapeHTML(value ?? "")}" ${field.min !== undefined ? `min="${field.min}"` : ""} ${required}>`;
  }
  return `<div class="field"><label>${escapeHTML(field.label)}</label>${control}</div>`;
}

function openForm(kind, record = null) {
  const dialog = document.querySelector("#record-dialog");
  const title = `${record ? "Edit" : "Add"} ${kind === "borrower" ? "borrower" : kind}`;
  document.querySelector("#dialog-title").textContent = title;
  document.querySelector("#dialog-eyebrow").textContent = record ? "UPDATE RECORD" : "NEW RECORD";
  document.querySelector("#dialog-submit").textContent = record ? "Save changes" : `Create ${kind}`;
  const fields = [...fieldDefinitions[kind]];
  if (kind === "borrowing" && record) fields.push({ name: "status", label: "Status", type: "status-select" });
  if (fields.at(-1)?.name === "status") {
    fields[fields.length - 1] = { ...fields.at(-1), type: "status-select" };
  }
  document.querySelector("#form-fields").innerHTML = fields.map((field) => {
    if (field.type === "status-select") {
      return `<div class="field"><label>Status</label><select name="status"><option ${record.status === "Borrowed" ? "selected" : ""}>Borrowed</option><option ${record.status === "Returned" ? "selected" : ""}>Returned</option></select></div>`;
    }
    return fieldMarkup(field, record?.[field.name]);
  }).join("");
  const form = document.querySelector("#record-form");
  form.dataset.kind = kind;
  form.dataset.id = record?.id ?? "";
  dialog.showModal();
}

async function deleteRecord(kind, id) {
  const collection = kind === "equipment" ? "equipment" : `${kind}s`;
  if (!window.confirm(`Delete this ${kind} record?`)) return;
  await api(`/api/${collection}/${id}`, { method: "DELETE" });
  showNotice(`${kind[0].toUpperCase()}${kind.slice(1)} deleted.`);
  await refreshData();
}

document.querySelectorAll(".nav-link").forEach((button) => button.addEventListener("click", () => setView(button.dataset.view)));
document.querySelectorAll("[data-open-view]").forEach((button) => button.addEventListener("click", () => setView(button.dataset.openView)));
document.querySelectorAll("[data-create]").forEach((button) => button.addEventListener("click", () => openForm(button.dataset.create)));
document.querySelector("#quick-create").addEventListener("click", () => openForm("borrowing"));
document.querySelectorAll("[data-close-dialog]").forEach((button) => button.addEventListener("click", () => document.querySelector("#record-dialog").close()));
document.querySelector("#record-dialog").addEventListener("click", (event) => {
  if (event.target === event.currentTarget) event.currentTarget.close();
});

document.querySelector("#record-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  const form = event.currentTarget;
  const kind = form.dataset.kind;
  const id = form.dataset.id;
  const payload = Object.fromEntries(new FormData(form).entries());
  for (const key of ["quantity", "equipment_id", "borrower_id"]) {
    if (key in payload) payload[key] = Number(payload[key]);
  }
  const collection = kind === "equipment" ? "equipment" : `${kind}s`;
  const method = id ? "PUT" : "POST";
  const path = id ? `/api/${collection}/${id}` : `/api/${collection}`;
  try {
    await api(path, { method, body: JSON.stringify(payload) });
    document.querySelector("#record-dialog").close();
    showNotice(`${kind[0].toUpperCase()}${kind.slice(1)} ${id ? "updated" : "created"}.`);
    await refreshData();
  } catch (error) {
    showNotice(error.message);
  }
});

document.querySelectorAll("#equipment-search, #borrowers-search, #borrowings-search").forEach((input) => {
  input.addEventListener("input", () => {
    if (input.id === "equipment-search") renderEquipment();
    if (input.id === "borrowers-search") renderBorrowers();
    if (input.id === "borrowings-search") renderBorrowings();
  });
});

document.querySelector("main").addEventListener("click", async (event) => {
  const button = event.target.closest("button[data-action]");
  if (!button) return;
  const { action, kind, id } = button.dataset;
  try {
    if (action === "edit") {
      const source = kind === "equipment" ? state.equipment : kind === "borrower" ? state.borrowers : state.borrowings;
      const record = source.find((item) => item.id === Number(id));
      if (record) openForm(kind, record);
    } else if (action === "delete") {
      await deleteRecord(kind, id);
    } else if (action === "return") {
      await api(`/api/borrowings/${id}/return`, { method: "PATCH" });
      showNotice("Equipment return recorded.");
      await refreshData();
    }
  } catch (error) {
    showNotice(error.message);
  }
});

refreshData();