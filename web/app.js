// GymRat Web Dashboard JavaScript Controller

const API_BASE = '/api';
let activeSessionId = localStorage.getItem('gymrat_session_id') || '';
let exerciseCatalog = [];
let workoutRoutines = [];
let workoutPlans = [];
let historyLogs = [];

// Live Workout State
let liveWorkoutSession = null;
let liveTimerInterval = null;
let liveTimerSeconds = 0;
let restTimerInterval = null;

document.addEventListener('DOMContentLoaded', async () => {
  initTabs();
  initModals();
  initUpload();
  initFilters();

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
    const el = document.getElementById('active-session-id');
    if (el) el.textContent = 'Creating...';

    const res = await fetch(`${API_BASE}/session/new`, { method: 'POST' });
    const data = await res.json();
    activeSessionId = data.sessionId;
    localStorage.setItem('gymrat_session_id', activeSessionId);
    updateSessionUI(activeSessionId);
    await loadDashboardData();
  } catch (err) {
    console.error('Failed to create new session:', err);
    const el = document.getElementById('active-session-id');
    if (el) el.textContent = 'Error Connecting';
  }
}

function updateSessionUI(id) {
  const el = document.getElementById('active-session-id');
  if (el) {
    el.textContent = id ? id.substring(0, 8) + '...' : 'Initializing...';
  }
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
      const targetEl = document.getElementById(targetId);
      if (targetEl) targetEl.classList.add('active');
    });
  });
}

// Data Fetching
async function loadDashboardData() {
  await fetchExercises();
  await fetchWorkouts();
  await fetchPlans();
  await fetchHistory();
}

async function fetchExercises() {
  try {
    const res = await fetch(`${API_BASE}/exercises`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    if (!res.ok) {
      console.warn('Invalid session response, initializing new session...');
      await createNewSession();
      return;
    }
    const data = await res.json();
    exerciseCatalog = Array.isArray(data) ? data : [];
    updateCategoryDropdownOptions();
    renderCatalog();
  } catch (err) {
    console.error('Error fetching exercises:', err);
    exerciseCatalog = [];
  }
}

async function fetchWorkouts() {
  try {
    const res = await fetch(`${API_BASE}/workouts`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    workoutRoutines = Array.isArray(data) ? data : [];
    renderRoutines();
  } catch (err) {
    console.error('Error fetching workout routines:', err);
    workoutRoutines = [];
  }
}

async function fetchPlans() {
  try {
    const res = await fetch(`${API_BASE}/plans`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    workoutPlans = Array.isArray(data) ? data : [];
    renderPlans();
  } catch (err) {
    console.error('Error fetching plans:', err);
    workoutPlans = [];
  }
}

async function fetchHistory() {
  try {
    const res = await fetch(`${API_BASE}/history`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    if (!res.ok) return;
    const data = await res.json();
    historyLogs = Array.isArray(data) ? data : [];
    renderHistory();
  } catch (err) {
    console.error('Error fetching history:', err);
    historyLogs = [];
  }
}

async function fetchHistory() {
  try {
    const res = await fetch(`${API_BASE}/history`, {
      headers: { 'X-Session-ID': activeSessionId }
    });
    historyLogs = await res.json();
    renderHistory();
  } catch (err) {
    console.error('Error fetching history:', err);
  }
}

// Search & Category Filters
function initFilters() {
  document.getElementById('search-exercise').addEventListener('input', renderCatalog);
  document.getElementById('filter-category').addEventListener('change', renderCatalog);
}

function updateCategoryDropdownOptions() {
  const select = document.getElementById('filter-category');
  const currentVal = select.value;
  const categories = [...new Set(exerciseCatalog.map(e => e.category).filter(Boolean))];

  select.innerHTML = '<option value="">All Categories</option>' +
    categories.map(c => `<option value="${escapeHtml(c)}">${escapeHtml(c)}</option>`).join('');
  select.value = currentVal;
}

// TIER 1: Render Exercise Library with Search & Category Filtering
function renderCatalog() {
  const grid = document.getElementById('catalog-grid');
  const showRemoved = document.getElementById('toggle-removed').checked;
  const searchQuery = (document.getElementById('search-exercise').value || '').toLowerCase();
  const selectedCategory = document.getElementById('filter-category').value;

  grid.innerHTML = '';

  if (!exerciseCatalog || exerciseCatalog.length === 0) {
    grid.innerHTML = '<p class="subtitle">No exercises found in library. Click "+ New Exercise".</p>';
    return;
  }

  const filtered = exerciseCatalog.filter(ex => {
    const matchesRemoved = showRemoved || !ex.isRemoved;
    const matchesSearch = (ex.name || '').toLowerCase().includes(searchQuery) ||
                          (ex.description || '').toLowerCase().includes(searchQuery);
    const matchesCategory = !selectedCategory || ex.category === selectedCategory;
    return matchesRemoved && matchesSearch && matchesCategory;
  });

  if (filtered.length === 0) {
    grid.innerHTML = '<p class="subtitle">No exercises match your search or filter criteria.</p>';
    return;
  }

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
        <button class="btn btn-secondary btn-sm" onclick="openUpdateExerciseModal('${ex.id}')">Bump Version</button>
        <button class="btn btn-secondary btn-sm" onclick="toggleSoftDelete('${ex.id}')">${ex.isRemoved ? 'Restore' : 'Soft Delete'}</button>
      </div>
    `;
    grid.appendChild(card);
  });
}

// TIER 2: Render Workout Routines
function renderRoutines() {
  const grid = document.getElementById('routines-grid');
  grid.innerHTML = '';

  if (!workoutRoutines || workoutRoutines.length === 0) {
    grid.innerHTML = '<p class="subtitle">No standalone workout routines created yet. Click "+ Create Workout Routine".</p>';
    return;
  }

  workoutRoutines.forEach(routine => {
    const card = document.createElement('div');
    card.className = 'exercise-card';

    let exSummary = '';
    if (routine.exercises) {
      Object.values(routine.exercises).forEach(we => {
        let setsSummary = (we.sets || []).map((s, i) => {
          return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight}lbs` : `${s.repCount} reps @ ${s.weight}lbs`}`;
        }).join(' | ');
        exSummary += `<div style="font-size: 0.85rem; margin-top: 4px;">• <strong>${escapeHtml(we.nameSnapshot || 'Exercise')} (${(we.sets || []).length} sets):</strong> ${setsSummary}</div>`;
      });
    }

    card.innerHTML = `
      <div class="card-top">
        <h3 class="ex-title">${escapeHtml(routine.name)}</h3>
        <span class="tag-version">Routine Template</span>
      </div>
      <p class="ex-desc">${escapeHtml(routine.description || 'Workout Routine Template')}</p>
      ${exSummary}
      <div class="card-actions" style="display: flex; gap: 8px; margin-top: 12px; flex-wrap: wrap;">
        <button class="btn btn-primary btn-sm" onclick="copyWorkoutTemplate('${routine.id}')">📋 Duplicate Routine</button>
        <button class="btn btn-secondary btn-sm" onclick="editRoutine('${routine.id}')">✏️ Edit Routine</button>
        <button class="btn btn-secondary btn-sm" style="color: var(--accent-rose);" onclick="deleteRoutine('${routine.id}')">🗑️ Delete</button>
      </div>
    `;
    grid.appendChild(card);
  });
}

/**
 * Triggers server-side duplication of a Workout Routine template by ID.
 * Generates a fresh UUID on the server and re-renders the dashboard.
 */
async function copyWorkoutTemplate(workoutId) {
  try {
    const res = await fetch(`${API_BASE}/workouts/copy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify({ workoutId })
    });
    if (!res.ok) throw new Error('Failed to copy workout template');
    await loadDashboardData();
  } catch (err) {
    alert(err.message);
  }
}

/**
 * Triggers server-side duplication of a Training Plan by ID.
 * Generates fresh UUIDs for the plan and all internal scheduled workouts, resetting execution status.
 */
async function copyPlan(planId) {
  try {
    const res = await fetch(`${API_BASE}/plans/copy`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify({ planId })
    });
    if (!res.ok) throw new Error('Failed to copy training plan');
    await loadDashboardData();
  } catch (err) {
    alert(err.message);
  }
}

// TIER 3: Render Training Plans with Live Logger Trigger
function renderPlans() {
  const container = document.getElementById('plans-container');
  container.innerHTML = '';

  if (!workoutPlans || workoutPlans.length === 0) {
    container.innerHTML = '<p class="subtitle">No training plans created yet. Click "+ Create Training Plan".</p>';
    return;
  }

  workoutPlans.forEach(plan => {
    const card = document.createElement('div');
    card.className = 'plan-card';

    let workoutsHtml = '';
    if (plan.workouts && plan.workouts.length > 0) {
      const sortedWorkouts = [...plan.workouts].sort((a, b) => new Date(a.datePlanned) - new Date(b.datePlanned));

      workoutsHtml = sortedWorkouts.map(w => {
        let exHtml = '';
        if (w.exercises) {
          Object.values(w.exercises).forEach(we => {
            let setsStr = we.sets ? we.sets.map((s, i) => {
              return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight}lbs` : `${s.repCount} reps @ ${s.weight}lbs`}`;
            }).join(' | ') : '';

            exHtml += `<div><strong>${escapeHtml(we.nameSnapshot || 'Exercise')}:</strong> ${setsStr}</div>`;
          });
        }

        const dateStr = w.datePlanned ? new Date(w.datePlanned).toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric', year: 'numeric' }) : 'Unscheduled';

        return `
          <div class="workout-item ${w.isExecuted ? 'executed' : ''}">
            <div class="workout-item-header">
              <div>
                <h4>${escapeHtml(w.name)}</h4>
                <div style="font-size: 0.78rem; color: var(--accent-cyan); font-weight: 600; margin-top: 2px;">
                  📅 Planned Date: ${dateStr}
                </div>
              </div>
              <div style="display: flex; gap: 8px;">
                ${!w.isExecuted ? `<button class="btn btn-accent btn-sm" onclick="startLiveWorkout('${plan.id}', '${w.id}')">▶ Start Workout</button>` : ''}
                <button class="btn btn-sm ${w.isExecuted ? 'btn-secondary' : 'btn-secondary'}" onclick="toggleWorkoutExecuted('${plan.id}', '${w.id}', ${!w.isExecuted})">
                  ${w.isExecuted ? '✓ Executed' : 'Mark Executed'}
                </button>
              </div>
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
          <p class="subtitle">${escapeHtml(plan.description || '')} • ${plan.workouts ? plan.workouts.length : 0} Scheduled Sessions</p>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
          <button class="btn btn-secondary btn-sm" onclick="copyPlan('${plan.id}')">📋 Duplicate Plan</button>
          <span class="badge">${plan.status || 'Active'}</span>
        </div>
      </div>
      <div class="workouts-subgrid">
        ${workoutsHtml || '<p class="subtitle">No workout instances scheduled in this plan.</p>'}
      </div>
    `;
    container.appendChild(card);
  });
}

// WORKOUT HISTORY LOG TAB
function renderHistory() {
  const container = document.getElementById('history-container');
  container.innerHTML = '';

  if (!historyLogs || historyLogs.length === 0) {
    container.innerHTML = '<p class="subtitle">No completed workout history logs recorded yet. Complete a workout to start building history!</p>';
    return;
  }

  const sortedLogs = [...historyLogs].sort((a, b) => new Date(b.dateCompleted) - new Date(a.dateCompleted));

  sortedLogs.forEach(log => {
    const card = document.createElement('div');
    card.className = 'plan-card';

    const durationMin = Math.round((log.durationSeconds || 0) / 60);
    const dateStr = log.dateCompleted ? new Date(log.dateCompleted).toLocaleString() : 'Recent';

    let exDetails = '';
    if (log.workoutSnapshot && log.workoutSnapshot.exercises) {
      Object.values(log.workoutSnapshot.exercises).forEach(we => {
        let setPills = (we.sets || []).map((s, i) => {
          return `Set ${i+1}: ${s.setType === 'timed' ? `${s.durationSeconds}s @ ${s.weight} lbs` : `${s.repCount} reps @ ${s.weight} lbs`}`;
        }).join(' | ');
        exDetails += `<div><strong>${escapeHtml(we.nameSnapshot || 'Exercise')}:</strong> ${setPills}</div>`;
      });
    }

    card.innerHTML = `
      <div class="plan-header">
        <div>
          <h3>🏋️‍♂️ ${escapeHtml(log.workoutName || 'Completed Workout')}</h3>
          <p class="subtitle">Completed: ${dateStr} • Duration: ${durationMin} mins</p>
        </div>
        <span class="badge" style="background: var(--gradient-accent); color: #000;">${(log.totalVolumeLbs || 0).toLocaleString()} lbs Volume</span>
      </div>
      <div style="font-size: 0.9rem; line-height: 1.6;">
        ${exDetails}
      </div>
    `;
    container.appendChild(card);
  });
}

// LIVE WORKOUT SESSION LOGGER
function startLiveWorkout(planId, workoutId) {
  const plan = workoutPlans.find(p => p.id === planId);
  if (!plan) return;
  const workout = (plan.workouts || []).find(w => w.id === workoutId);
  if (!workout) return;

  // Clone workout object for live logging
  liveWorkoutSession = {
    planId: plan.id,
    workoutId: workout.id,
    workoutName: workout.name,
    exercises: JSON.parse(JSON.stringify(workout.exercises || {}))
  };

  document.getElementById('live-workout-name').textContent = `Live: ${workout.name}`;
  document.getElementById('live-rest-timer').textContent = '';
  
  // Reset and start live timer
  liveTimerSeconds = 0;
  clearInterval(liveTimerInterval);
  liveTimerInterval = setInterval(() => {
    liveTimerSeconds++;
    const hrs = String(Math.floor(liveTimerSeconds / 3600)).padStart(2, '0');
    const mins = String(Math.floor((liveTimerSeconds % 3600) / 60)).padStart(2, '0');
    const secs = String(liveTimerSeconds % 60).padStart(2, '0');
    document.getElementById('live-workout-timer').textContent = `⏱️ Session Duration: ${hrs}:${mins}:${secs}`;
  }, 1000);

  renderLiveWorkoutBody();
  document.getElementById('modal-live-workout').classList.add('active');
}

function renderLiveWorkoutBody() {
  const body = document.getElementById('live-workout-body');
  body.innerHTML = '';

  if (!liveWorkoutSession || !liveWorkoutSession.exercises) return;

  Object.values(liveWorkoutSession.exercises).forEach(we => {
    const exDiv = document.createElement('div');
    exDiv.style.cssText = 'background: rgba(0,0,0,0.3); padding: 14px; margin-bottom: 14px; border-radius: 10px; border: 1px solid var(--border-color);';

    let setsHtml = (we.sets || []).map((s, idx) => `
      <div style="display: flex; gap: 10px; align-items: center; margin-top: 8px; font-size: 0.88rem;">
        <span style="width: 50px; font-weight: 600;">Set ${idx + 1}:</span>
        <label>Weight (lbs):</label>
        <input type="number" class="live-weight" data-ex="${we.exerciseId}" data-set="${idx}" value="${s.weight || 45}" style="width: 80px; padding: 6px; background: rgba(0,0,0,0.5); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
        <label>${s.setType === 'timed' ? 'Seconds:' : 'Reps:'}</label>
        <input type="number" class="live-reps" data-ex="${we.exerciseId}" data-set="${idx}" value="${s.setType === 'timed' ? (s.durationSeconds || 30) : (s.repCount || 10)}" style="width: 80px; padding: 6px; background: rgba(0,0,0,0.5); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
        <button class="btn btn-secondary btn-sm" onclick="triggerRestTimer(30, this)">✓ Set Done</button>
      </div>
    `).join('');

    exDiv.innerHTML = `
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px;">
        <h4 style="color: var(--accent-emerald); font-size: 1.1rem;">${escapeHtml(we.nameSnapshot || 'Exercise')}</h4>
        <button class="btn btn-secondary btn-sm" onclick="addExtraLiveSet('${we.exerciseId}')">+ Add Set</button>
      </div>
      ${setsHtml}
    `;
    body.appendChild(exDiv);
  });
}

function addExtraLiveSet(exerciseId) {
  if (!liveWorkoutSession || !liveWorkoutSession.exercises[exerciseId]) return;

  const currentSets = liveWorkoutSession.exercises[exerciseId].sets || [];
  const lastSet = currentSets[currentSets.length - 1] || { setType: 'reps', repCount: 10, weight: 45 };

  currentSets.push({
    id: crypto.randomUUID(),
    setType: lastSet.setType,
    repCount: lastSet.repCount,
    durationSeconds: lastSet.durationSeconds,
    weight: lastSet.weight,
    perceivedEffort: 7
  });

  liveWorkoutSession.exercises[exerciseId].sets = currentSets;
  renderLiveWorkoutBody();
}

function triggerRestTimer(seconds = 30, btnEl) {
  if (btnEl) {
    btnEl.style.background = 'var(--accent-emerald)';
    btnEl.style.color = '#000';
    btnEl.textContent = '✓ Completed';
  }

  let restRemaining = seconds;
  clearInterval(restTimerInterval);
  const timerEl = document.getElementById('live-rest-timer');

  restTimerInterval = setInterval(() => {
    restRemaining--;
    if (restRemaining <= 0) {
      clearInterval(restTimerInterval);
      timerEl.textContent = '⚡ Rest Finished! Ready for next set.';
    } else {
      timerEl.textContent = `⏱️ Rest Timer: ${restRemaining}s remaining`;
    }
  }, 1000);
}

// Finish & Log Live Workout Session
async function finishLiveWorkout() {
  if (!liveWorkoutSession) return;

  clearInterval(liveTimerInterval);
  clearInterval(restTimerInterval);

  let totalVolumeLbs = 0;

  const weightInputs = document.querySelectorAll('.live-weight');
  const repsInputs = document.querySelectorAll('.live-reps');

  weightInputs.forEach((wInp, idx) => {
    const weight = parseFloat(wInp.value) || 0;
    const reps = parseFloat(repsInputs[idx] ? repsInputs[idx].value : 0) || 0;
    const exId = wInp.dataset.ex;
    const setIdx = parseInt(wInp.dataset.set);

    if (liveWorkoutSession.exercises[exId] && liveWorkoutSession.exercises[exId].sets[setIdx]) {
      const setObj = liveWorkoutSession.exercises[exId].sets[setIdx];
      setObj.weight = weight;
      if (setObj.setType === 'timed') {
        setObj.durationSeconds = reps;
      } else {
        setObj.repCount = reps;
        totalVolumeLbs += (weight * reps);
      }
    }
  });

  const historyLog = {
    id: crypto.randomUUID(),
    workoutId: liveWorkoutSession.workoutId,
    planId: liveWorkoutSession.planId,
    workoutName: liveWorkoutSession.workoutName,
    dateCompleted: new Date().toISOString(),
    durationSeconds: liveTimerSeconds,
    totalVolumeLbs: totalVolumeLbs,
    workoutSnapshot: {
      id: liveWorkoutSession.workoutId,
      name: liveWorkoutSession.workoutName,
      datePlanned: new Date().toISOString(),
      isExecuted: true,
      exercises: liveWorkoutSession.exercises
    }
  };

  await fetch(`${API_BASE}/history`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Session-ID': activeSessionId
    },
    body: JSON.stringify(historyLog)
  });

  await toggleWorkoutExecuted(liveWorkoutSession.planId, liveWorkoutSession.workoutId, true);

  document.getElementById('modal-live-workout').classList.remove('active');
  liveWorkoutSession = null;
  await loadDashboardData();
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

async function deleteRoutine(id) {
  try {
    await fetch(`${API_BASE}/workouts?id=${id}`, {
      method: 'DELETE',
      headers: { 'X-Session-ID': activeSessionId }
    });
    await fetchWorkouts();
  } catch (err) {
    console.error('Error deleting routine:', err);
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
  const modalRoutine = document.getElementById('modal-routine');
  const modalPlan = document.getElementById('modal-plan');
  const modalLive = document.getElementById('modal-live-workout');

  // Open Exercise Modal
  document.getElementById('btn-add-exercise').addEventListener('click', () => {
    document.getElementById('ex-id').value = '';
    document.getElementById('modal-exercise-title').textContent = 'New Exercise';
    document.getElementById('form-exercise').reset();
    modalEx.classList.add('active');
  });

  // Open Routine Modal
  document.getElementById('btn-add-routine').addEventListener('click', () => {
    document.getElementById('form-routine').reset();
    document.getElementById('routine-exercises-builder').innerHTML = '';
    addRoutineExerciseStep();
    modalRoutine.classList.add('active');
  });

  // Open Plan Modal
  document.getElementById('btn-add-plan').addEventListener('click', () => {
    document.getElementById('form-plan').reset();
    document.getElementById('plan-schedule-builder').innerHTML = '';
    addPlanScheduleStep();
    modalPlan.classList.add('active');
  });

  document.getElementById('btn-finish-live-workout').addEventListener('click', finishLiveWorkout);

  // Close buttons
  document.querySelectorAll('.modal-close, .modal-cancel').forEach(btn => {
    btn.addEventListener('click', () => {
      modalEx.classList.remove('active');
      modalRoutine.classList.remove('active');
      modalPlan.classList.remove('active');
      modalLive.classList.remove('active');
      clearInterval(liveTimerInterval);
      clearInterval(restTimerInterval);
    });
  });

  // Exercise Form Submit
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

  document.getElementById('btn-add-routine-ex-step').addEventListener('click', addRoutineExerciseStep);
  document.getElementById('btn-add-plan-schedule-step').addEventListener('click', addPlanScheduleStep);

  // Routine Form Submit (Multi-Set Builder & Edit Support)
  document.getElementById('form-routine').addEventListener('submit', async (e) => {
    e.preventDefault();
    const routineId = document.getElementById('routine-id').value;
    const name = document.getElementById('routine-name').value;
    const description = document.getElementById('routine-desc').value;

    const exEls = document.querySelectorAll('.routine-ex-builder-item');
    const exercisesMap = {};

    exEls.forEach(el => {
      const exSelect = el.querySelector('.r-ex-select');
      const selectedExId = exSelect ? exSelect.value : '';
      const selectedEx = exerciseCatalog.find(e => e.id === selectedExId) || { name: 'Custom Exercise', version: 1 };

      const setRows = el.querySelectorAll('.routine-set-row');
      const setsArr = [];

      setRows.forEach(sRow => {
        const setType = sRow.querySelector('.r-set-type').value;
        const valCount = parseInt(sRow.querySelector('.r-val').value) || 10;
        const weight = parseFloat(sRow.querySelector('.r-weight').value) || 0;

        setsArr.push({
          id: crypto.randomUUID(),
          setType: setType,
          repCount: setType === 'reps' ? valCount : 0,
          durationSeconds: setType === 'timed' ? valCount : 0,
          weight: weight,
          perceivedEffort: 7
        });
      });

      if (selectedExId) {
        exercisesMap[selectedExId] = {
          exerciseId: selectedExId,
          exerciseVersion: selectedEx.version,
          nameSnapshot: selectedEx.name,
          sets: setsArr
        };
      }
    });

    const targetRoutine = {
      id: routineId || crypto.randomUUID(),
      name,
      description,
      datePlanned: new Date().toISOString(),
      isExecuted: false,
      exercises: exercisesMap
    };

    const method = routineId ? 'PUT' : 'POST';
    await fetch(`${API_BASE}/workouts`, {
      method: method,
      headers: {
        'Content-Type': 'application/json',
        'X-Session-ID': activeSessionId
      },
      body: JSON.stringify(targetRoutine)
    });

    document.getElementById('modal-routine').classList.remove('active');
    await fetchWorkouts();
  });

  // Plan Form Submit (Tier 3)
  document.getElementById('form-plan').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('plan-name').value;
    const description = document.getElementById('plan-desc').value;

    const rowEls = document.querySelectorAll('.plan-schedule-row');
    const selectedWorkouts = [];

    rowEls.forEach(row => {
      const rId = row.querySelector('.p-routine-select').value;
      const dateVal = row.querySelector('.p-date-input').value;
      const foundRoutine = workoutRoutines.find(r => r.id === rId);

      if (foundRoutine) {
        const plannedDate = dateVal ? new Date(dateVal + 'T00:00:00').toISOString() : new Date().toISOString();

        selectedWorkouts.push({
          id: crypto.randomUUID(),
          name: foundRoutine.name,
          description: foundRoutine.description,
          datePlanned: plannedDate,
          isExecuted: false,
          exercises: JSON.parse(JSON.stringify(foundRoutine.exercises || {}))
        });
      }
    });

    const newPlan = {
      id: crypto.randomUUID(),
      name,
      description,
      status: 'active',
      datePlanned: new Date().toISOString(),
      workouts: selectedWorkouts
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

function openUpdateExerciseModal(id) {
  const ex = exerciseCatalog.find(e => e.id === id);
  if (!ex) return;

  document.getElementById('ex-id').value = ex.id;
  document.getElementById('modal-exercise-title').textContent = `Update Exercise (Bumps to v${ex.version + 1})`;
  document.getElementById('ex-name').value = ex.name;
  document.getElementById('ex-category').value = ex.category || '';
  document.getElementById('ex-desc').value = ex.description || '';
  document.getElementById('modal-exercise').classList.add('active');
}

function editRoutine(routineId) {
  const routine = workoutRoutines.find(r => r.id === routineId);
  if (!routine) return;

  document.getElementById('routine-id').value = routine.id;
  document.getElementById('routine-name').value = routine.name;
  document.getElementById('routine-desc').value = routine.description || '';
  document.getElementById('modal-routine-title').textContent = 'Edit Workout Routine';

  const builder = document.getElementById('routine-exercises-builder');
  builder.innerHTML = '';

  if (routine.exercises && Object.keys(routine.exercises).length > 0) {
    Object.values(routine.exercises).forEach(we => {
      addRoutineExerciseStep(we);
    });
  } else {
    addRoutineExerciseStep();
  }

  document.getElementById('modal-routine').classList.add('active');
}

function addRoutineExerciseStep(existingWe = null) {
  const container = document.getElementById('routine-exercises-builder');
  const index = container.children.length + 1;
  const exIdIndex = `ex_builder_${Date.now()}_${index}`;

  let optionsHtml = exerciseCatalog.map(ex => {
    const isSel = existingWe && existingWe.exerciseId === ex.id ? 'selected' : '';
    return `<option value="${ex.id}" ${isSel}>${escapeHtml(ex.name)} (v${ex.version})</option>`;
  }).join('');

  const div = document.createElement('div');
  div.className = 'routine-ex-builder-item';
  div.style.cssText = 'background: rgba(0,0,0,0.3); padding: 14px; margin-bottom: 14px; border-radius: 8px; border: 1px solid var(--border-color)';
  div.innerHTML = `
    <div class="form-group">
      <label>Select Exercise #${index} from Library</label>
      <select class="r-ex-select">${optionsHtml || '<option value="">No Library Exercises Available</option>'}</select>
    </div>
    
    <div class="r-sets-container" id="sets_container_${exIdIndex}">
      <!-- Set rows dynamically added -->
    </div>

    <button type="button" class="btn btn-secondary btn-sm" onclick="addSetRowToRoutine('sets_container_${exIdIndex}')">+ Add Set</button>
  `;
  container.appendChild(div);

  if (existingWe && existingWe.sets && existingWe.sets.length > 0) {
    existingWe.sets.forEach(s => addSetRowToRoutine(`sets_container_${exIdIndex}`, s));
  } else {
    // Add 3 default set rows for convenient workout routine building
    addSetRowToRoutine(`sets_container_${exIdIndex}`);
    addSetRowToRoutine(`sets_container_${exIdIndex}`);
    addSetRowToRoutine(`sets_container_${exIdIndex}`);
  }
}

function addSetRowToRoutine(containerId, setObj = null) {
  const container = document.getElementById(containerId);
  if (!container) return;

  const setIndex = container.children.length + 1;
  const setType = setObj ? (setObj.setType || 'reps') : 'reps';
  const valCount = setObj ? (setType === 'timed' ? (setObj.durationSeconds || 30) : (setObj.repCount || 10)) : 10;
  const weight = setObj ? (setObj.weight || 0) : 45;

  const setRow = document.createElement('div');
  setRow.className = 'routine-set-row';
  setRow.style.cssText = 'display: flex; gap: 10px; align-items: center; margin-bottom: 8px; background: rgba(255,255,255,0.03); padding: 8px; border-radius: 6px;';
  setRow.innerHTML = `
    <span style="font-size: 0.85rem; font-weight: 600; width: 45px;">Set ${setIndex}</span>
    <div style="flex: 1;">
      <select class="r-set-type" style="padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
        <option value="reps" ${setType === 'reps' ? 'selected' : ''}>Counting (Reps)</option>
        <option value="timed" ${setType === 'timed' ? 'selected' : ''}>Timed (Seconds)</option>
      </select>
    </div>
    <div style="flex: 1;">
      <input type="number" class="r-val" value="${valCount}" placeholder="Reps/Sec" style="width: 100%; padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
    </div>
    <div style="flex: 1;">
      <input type="number" class="r-weight" value="${weight}" placeholder="Weight (lbs)" style="width: 100%; padding: 6px; background: rgba(0,0,0,0.4); border: 1px solid var(--border-color); color: #fff; border-radius: 4px;">
    </div>
  `;
  container.appendChild(setRow);
}

function addPlanScheduleStep() {
  const container = document.getElementById('plan-schedule-builder');
  const index = container.children.length + 1;
  const todayStr = new Date().toISOString().split('T')[0];

  let routinesHtml = workoutRoutines.map(r => `<option value="${r.id}">${escapeHtml(r.name)}</option>`).join('');

  const div = document.createElement('div');
  div.className = 'plan-schedule-row';
  div.style.cssText = 'background: rgba(0,0,0,0.3); padding: 12px; margin-bottom: 12px; border-radius: 8px; border: 1px solid var(--border-color)';
  div.innerHTML = `
    <div style="display: flex; gap: 12px; align-items: flex-end;">
      <div class="form-group" style="flex: 2; margin-bottom: 0;">
        <label>Session #${index} Routine Template</label>
        <select class="p-routine-select">${routinesHtml || '<option value="">No Routine Templates Available</option>'}</select>
      </div>
      <div class="form-group" style="flex: 1; margin-bottom: 0;">
        <label>Planned Date</label>
        <input type="date" class="p-date-input" value="${todayStr}">
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
