// GymRat Web Dashboard JavaScript Controller

const API_BASE = '/api';
let activeSessionId = localStorage.getItem('gymrat_session_id') || '';
let exerciseCatalog = [];
let workoutPlans = [];

document.addEventListener('DOMContentLoaded', async () => {
  initTabs();
  initModals();
  initUpload();

  if (!activeSessionId) {
    await createNewSession();
  } else {
    updateSessionUI(activeSessionId);
    await loadDashboardData();
  }

  document.getElementById('btn-new-session').addEventListener('click', createNewSession);
  document.getElementById('btn-export-vault').addEventListener('click', exportVault);
  document.getElementById('toggle-removed').addEventListener('change', renderCatalog);
});

// Session Lifecycle
async function createNewSession() {
  try {
    const res = await fetch(`${API_BASE}/session/new`, { method: 'POST' });
    const data = await res.json();
    activeSessionId = data.sessionId;
    localStorage.setItem('gymrat_session_id', activeSessionId);
    updateSessionUI(activeSessionId);
    await loadDashboardData();
  } catch (err) {
    console.error('Failed to create new session:', err);
  }
}

function updateSessionUI(id) {
  document.getElementById('active-session-id').textContent = id.substring(0, 8) + '...';
}

// Navigation Tabs
function initTabs() {
  const tabs = document.querySelectorAll('.tab-btn');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => t.classList.remove('active'));
      document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));

      tab.classList.add('active');
      const targetId = `tab-${tab.dataset.tab}`;
      document.getElementById(targetId).classList.add('active');
    });
  });
}

// Data Fetching
async function loadDashboardData() {
  await fetchExercises();
  await fetchPlans();
}

async function fetchExercises() {
  try {
    const res = await fetch(`${API_BASE}/exercises`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    exerciseCatalog = await res.json();
    renderCatalog();
  } catch (err) {
    console.error('Error fetching exercises:', err);
  }
}

async function fetchPlans() {
  try {
    const res = await fetch(`${API_BASE}/plans`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    workoutPlans = await res.json();
    renderPlans();
  } catch (err) {
    console.error('Error fetching plans:', err);
  }
}

// Render Exercise Catalog
function renderCatalog() {
  const grid = document.getElementById('catalog-grid');
  const showRemoved = document.getElementById('toggle-removed').checked;
  grid.innerHTML = '';

  if (!exerciseCatalog || exerciseCatalog.length === 0) {
    grid.innerHTML = '<p class="subtitle">No exercises found. Click "+ New Exercise" to get started.</p>';
    return;
  }

  // Filter exercises
  const filtered = exerciseCatalog.filter(ex => showRemoved || !ex.isRemoved);

  filtered.forEach(ex => {
    const card = document.createElement('div');
    card.className = `exercise-card ${ex.isRemoved ? 'removed' : ''}`;
    card.innerHTML = `
      <div class="card-top">
        <h3 class="ex-title">${escapeHtml(ex.name)}</h3>
        <span class="tag-version">v${ex.version}</span>
      </div>
      <span class="tag-category">${escapeHtml(ex.category || 'General')}</span>
      <p class="ex-desc">${escapeHtml(ex.description || 'No description provided.')}</p>
      ${ex.isRemoved ? '<p class="tag-category" style="color: var(--accent-red)">Soft-Deleted (IsRemoved)</p>' : ''}
      <div class="card-actions">
        <button class="btn btn-secondary btn-sm" onclick="openUpdateModal('${ex.id}')">Bump Version</button>
        <button class="btn btn-secondary btn-sm" onclick="toggleSoftDelete('${ex.id}')">${ex.isRemoved ? 'Restore' : 'Soft Delete'}</button>
      </div>
    `;
    grid.appendChild(card);
  });
}

// Render Workout Plans
function renderPlans() {
  const container = document.getElementById('plans-container');
  container.innerHTML = '';

  if (!workoutPlans || workoutPlans.length === 0) {
    container.innerHTML = '<p class="subtitle">No workout plans created yet. Click "+ Create Plan".</p>';
    return;
  }

  workoutPlans.forEach(plan => {
    const card = document.createElement('div');
    card.className = 'plan-card';

    let workoutsHtml = '';
    if (plan.workouts && plan.workouts.length > 0) {
      workoutsHtml = plan.workouts.map(w => {
        let exHtml = '';
        if (w.exercises) {
          Object.values(w.exercises).forEach(we => {
            let setsStr = we.sets ? we.sets.map(s => {
              return s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight}lbs` : `${s.repCount} reps @ ${s.weight}lbs`;
            }).join(' | ') : '';

            exHtml += `<div><strong>${escapeHtml(we.nameSnapshot || 'Exercise')}:</strong> ${setsStr}</div>`;
          });
        }

        return `
          <div class="workout-item ${w.isExecuted ? 'executed' : ''}">
            <div class="workout-item-header">
              <h4>${escapeHtml(w.name)}</h4>
              <button class="btn btn-sm ${w.isExecuted ? 'btn-secondary' : 'btn-accent'}" onclick="toggleWorkoutExecuted('${plan.id}', '${w.id}', ${!w.isExecuted})">
                ${w.isExecuted ? '✓ Executed' : 'Mark Executed'}
              </button>
            </div>
            <p class="ex-desc">${escapeHtml(w.description || '')}</p>
            ${exHtml}
          </div>
        `;
      }).join('');
    }

    card.innerHTML = `
      <div class="plan-header">
        <div>
          <h3>${escapeHtml(plan.name)}</h3>
          <p class="subtitle">${escapeHtml(plan.description || '')} • Planned: ${new Date(plan.datePlanned).toLocaleDateString()}</p>
        </div>
        <span class="badge">${plan.status || 'Active'}</span>
      </div>
      <div class="workouts-subgrid">
        ${workoutsHtml}
      </div>
    `;
    container.appendChild(card);
  });
}

// Exercise Actions
async function toggleSoftDelete(id) {
  try {
    await fetch(`${API_BASE}/exercises?id=${id}`, {
      method: 'DELETE',
      headers: { 'X-Session-ID': activeSessionId }
    });
    await fetchExercises();
  } catch (err) {
    console.error('Error soft deleting exercise:', err);
  }
}

async function toggleWorkoutExecuted(planId, workoutId, isExecuted) {
  try {
    await fetch(`${API_BASE}/plans`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify({ planId, workoutId, isExecuted })
    });
    await fetchPlans();
  } catch (err) {
    console.error('Error toggling workout executed:', err);
  }
}

// Modals Setup
function initModals() {
  const modalEx = document.getElementById('modal-exercise');
  const modalPlan = document.getElementById('modal-plan');

  document.getElementById('btn-add-exercise').addEventListener('click', () => {
    document.getElementById('ex-id').value = '';
    document.getElementById('modal-exercise-title').textContent = 'New Exercise';
    document.getElementById('form-exercise').reset();
    modalEx.classList.add('active');
  });

  document.getElementById('btn-add-plan').addEventListener('click', () => {
    document.getElementById('form-plan').reset();
    document.getElementById('plan-workouts-builder').innerHTML = '';
    addWorkoutStep(); // default 1 workout row
    modalPlan.classList.add('active');
  });

  document.querySelectorAll('.modal-close, .modal-cancel').forEach(btn => {
    btn.addEventListener('click', () => {
      modalEx.classList.remove('active');
      modalPlan.classList.remove('active');
    });
  });

  document.getElementById('form-exercise').addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = document.getElementById('ex-id').value;
    const name = document.getElementById('ex-name').value;
    const category = document.getElementById('ex-category').value;
    const description = document.getElementById('ex-desc').value;

    const method = id ? 'PUT' : 'POST';
    await fetch(`${API_BASE}/exercises`, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify({ id, name, category, description })
    });

    modalEx.classList.remove('active');
    await fetchExercises();
  });

  document.getElementById('btn-add-workout-step').addEventListener('click', addWorkoutStep);

  document.getElementById('form-plan').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('plan-name').value;
    const description = document.getElementById('plan-desc').value;

    const workoutEls = document.querySelectorAll('.workout-builder-item');
    const workouts = [];

    workoutEls.forEach(wEl => {
      const wName = wEl.querySelector('.w-name').value;
      const wDesc = wEl.querySelector('.w-desc').value;
      const exSelect = wEl.querySelector('.w-ex-select');
      const setType = wEl.querySelector('.w-set-type').value;
      const valCount = parseInt(wEl.querySelector('.w-val').value) || 10;
      const weight = parseFloat(wEl.querySelector('.w-weight').value) || 0;

      const selectedExId = exSelect.value;
      const selectedEx = exerciseCatalog.find(e => e.id === selectedExId) || { name: 'Custom Exercise', version: 1 };

      const setObj = {
        id: crypto.randomUUID(),
        setType: setType,
        repCount: setType === 'reps' ? valCount : 0,
        durationSeconds: setType === 'timed' ? valCount : 0,
        weight: weight,
        perceivedEffort: 7
      };

      const exercisesMap = {};
      if (selectedExId) {
        exercisesMap[selectedExId] = {
          exerciseId: selectedExId,
          exerciseVersion: selectedEx.version,
          nameSnapshot: selectedEx.name,
          sets: [setObj]
        };
      }

      workouts.push({
        id: crypto.randomUUID(),
        name: wName,
        description: wDesc,
        datePlanned: new Date().toISOString(),
        isExecuted: false,
        exercises: exercisesMap
      });
    });

    const newPlan = {
      name,
      description,
      status: 'active',
      datePlanned: new Date().toISOString(),
      workouts
    };

    await fetch(`${API_BASE}/plans`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify(newPlan)
    });

    modalPlan.classList.remove('active');
    await fetchPlans();
  });
}

function openUpdateModal(id) {
  const ex = exerciseCatalog.find(e => e.id === id);
  if (!ex) return;

  document.getElementById('ex-id').value = ex.id;
  document.getElementById('modal-exercise-title').textContent = `Update Exercise (Bumps to v${ex.version + 1})`;
  document.getElementById('ex-name').value = ex.name;
  document.getElementById('ex-category').value = ex.category || '';
  document.getElementById('ex-desc').value = ex.description || '';
  document.getElementById('modal-exercise').classList.add('active');
}

function addWorkoutStep() {
  const container = document.getElementById('plan-workouts-builder');
  const index = container.children.length + 1;

  let optionsHtml = exerciseCatalog.map(ex => `<option value="${ex.id}">${escapeHtml(ex.name)} (v${ex.version})</option>`).join('');

  const div = document.createElement('div');
  div.className = 'workout-builder-item';
  div.style.cssText = 'background: rgba(0,0,0,0.3); padding: 12px; margin-bottom: 12px; border-radius: 8px; border: 1px solid var(--border-color)';
  div.innerHTML = `
    <div class="form-group">
      <label>Workout #${index} Name</label>
      <input type="text" class="w-name" required value="Workout ${index}">
    </div>
    <div class="form-group">
      <label>Description</label>
      <input type="text" class="w-desc" placeholder="Target focus">
    </div>
    <div class="form-group">
      <label>Select Exercise from Catalog</label>
      <select class="w-ex-select">${optionsHtml || '<option value="">No Catalog Exercises Available</option>'}</select>
    </div>
    <div style="display: flex; gap: 10px;">
      <div class="form-group" style="flex: 1;">
        <label>Set Type</label>
        <select class="w-set-type">
          <option value="reps">Counting Sets (Reps)</option>
          <option value="timed">Timed Sets (Seconds)</option>
        </select>
      </div>
      <div class="form-group" style="flex: 1;">
        <label>Reps or Duration (sec)</label>
        <input type="number" class="w-val" value="10">
      </div>
      <div class="form-group" style="flex: 1;">
        <label>Weight (lbs)</label>
        <input type="number" class="w-weight" value="45">
      </div>
    </div>
  `;
  container.appendChild(div);
}

// Upload & Export Handlers
function initUpload() {
  const dropZone = document.getElementById('drop-zone');
  const fileInput = document.getElementById('file-input');
  const btnBrowse = document.getElementById('btn-browse-file');
  const statusMsg = document.getElementById('upload-status');

  btnBrowse.addEventListener('click', () => fileInput.click());

  fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
      uploadFile(e.target.files[0]);
    }
  });

  dropZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    dropZone.style.borderColor = 'var(--accent-cyan)';
  });

  dropZone.addEventListener('dragleave', () => {
    dropZone.style.borderColor = 'var(--border-accent)';
  });

  dropZone.addEventListener('drop', (e) => {
    e.preventDefault();
    dropZone.style.borderColor = 'var(--border-accent)';
    if (e.dataTransfer.files.length > 0) {
      uploadFile(e.dataTransfer.files[0]);
    }
  });
}

async function uploadFile(file) {
  const statusMsg = document.getElementById('upload-status');
  statusMsg.textContent = 'Uploading and validating JSON vault...';

  try {
    const text = await file.text();
    const res = await fetch(`${API_BASE}/vault/upload`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: text
    });

    if (res.ok) {
      statusMsg.textContent = 'Vault imported successfully!';
      await loadDashboardData();
    } else {
      const err = await res.text();
      statusMsg.textContent = `Upload failed: ${err}`;
    }
  } catch (err) {
    statusMsg.textContent = `Upload error: ${err.message}`;
  }
}

function exportVault() {
  window.location.href = `${API_BASE}/vault/export?sessionId=${activeSessionId}`;
}

function escapeHtml(str) {
  return (str || '').replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
